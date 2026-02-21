package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type GeofenceEngine struct {
	db     *pgxpool.Pool
	nc     *nats.Conn
	logger *zap.Logger
	
	// Cache of vehicle last known geofence state
	vehicleStates map[uuid.UUID]*VehicleState
}

type VehicleState struct {
	VehicleID      uuid.UUID
	InsideGeofences map[uuid.UUID]time.Time // geofence_id -> entry_time
	LastChecked    time.Time
}

type GPSEvent struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
	Ignition  bool    `json:"ignition"`
}

func NewGeofenceEngine(db *pgxpool.Pool, nc *nats.Conn, logger *zap.Logger) *GeofenceEngine {
	return &GeofenceEngine{
		db:            db,
		nc:            nc,
		logger:        logger,
		vehicleStates: make(map[uuid.UUID]*VehicleState),
	}
}

// Start - Subscribe to GPS events and process geofence checks
func (ge *GeofenceEngine) Start(ctx context.Context) error {
	ge.logger.Info("Starting geofence engine...")

	// Subscribe to GPS telemetry stream
	sub, err := ge.nc.Subscribe("fleet.telemetry.gps", func(msg *nats.Msg) {
		var event GPSEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			ge.logger.Error("Failed to parse GPS event", zap.Error(err))
			return
		}

		vehicleID, err := uuid.Parse(event.VehicleID)
		if err != nil {
			return
		}

		// Process geofence checks for this GPS point
		if err := ge.processGPSEvent(ctx, vehicleID, event); err != nil {
			ge.logger.Error("Failed to process geofence event",
				zap.String("vehicle_id", event.VehicleID),
				zap.Error(err))
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to GPS events: %w", err)
	}

	ge.logger.Info("Geofence engine started", zap.String("subject", "fleet.telemetry.gps"))

	// Wait for context cancellation
	<-ctx.Done()
	sub.Unsubscribe()
	return nil
}

func (ge *GeofenceEngine) processGPSEvent(ctx context.Context, vehicleID uuid.UUID, event GPSEvent) error {
	// Query all active geofences assigned to this vehicle
	rows, err := ge.db.Query(ctx, `
		SELECT g.id, g.name, g.type, g.geometry, g.radius_meters,
		       ga.assignment_type, ga.dwell_minutes
		FROM geofences g
		JOIN geofence_assignments ga ON ga.geofence_id = g.id
		WHERE ga.vehicle_id = $1 
		  AND ga.is_active = true 
		  AND g.is_active = true
	`, vehicleID)

	if err != nil {
		return err
	}
	defer rows.Close()

	point := fmt.Sprintf("ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography", event.Longitude, event.Latitude)
	
	// Get or create vehicle state
	state, exists := ge.vehicleStates[vehicleID]
	if !exists {
		state = &VehicleState{
			VehicleID:       vehicleID,
			InsideGeofences: make(map[uuid.UUID]time.Time),
			LastChecked:     time.Unix(event.Timestamp, 0),
		}
		ge.vehicleStates[vehicleID] = state
	}

	currentlyInside := make(map[uuid.UUID]bool)

	for rows.Next() {
		var (
			geofenceID     uuid.UUID
			name           string
			geoType        string
			geometry       string
			radiusMeters   *int
			assignmentType string
			dwellMinutes   *int
		)

		if err := rows.Scan(&geofenceID, &name, &geoType, &geometry, &radiusMeters, &assignmentType, &dwellMinutes); err != nil {
			continue
		}

		// Check if vehicle is inside this geofence using PostGIS
		var isInside bool
		query := fmt.Sprintf(`
			SELECT ST_DWithin(
				'%s'::geography,
				%s,
				COALESCE($1, 0)
			)
		`, geometry, point)

		err = ge.db.QueryRow(ctx, query, radiusMeters).Scan(&isInside)
		if err != nil {
			continue
		}

		if isInside {
			currentlyInside[geofenceID] = true

			// Check for ENTER event
			if _, wasInside := state.InsideGeofences[geofenceID]; !wasInside {
				if assignmentType == "enter" || assignmentType == "both" {
					ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, "enter", event)
				}
				state.InsideGeofences[geofenceID] = time.Unix(event.Timestamp, 0)
			} else {
				// Check for DWELL event
				if assignmentType == "dwell" && dwellMinutes != nil {
					entryTime := state.InsideGeofences[geofenceID]
					duration := time.Unix(event.Timestamp, 0).Sub(entryTime)
					if duration >= time.Duration(*dwellMinutes)*time.Minute {
						ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, "dwell", event)
						// Reset entry time to avoid duplicate dwell alerts
						state.InsideGeofences[geofenceID] = time.Unix(event.Timestamp, 0)
					}
				}
			}
		}
	}

	// Check for EXIT events
	for geofenceID, entryTime := range state.InsideGeofences {
		if !currentlyInside[geofenceID] {
			duration := time.Unix(event.Timestamp, 0).Sub(entryTime)
			ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, "exit", event)
			delete(state.InsideGeofences, geofenceID)
			
			// Log to TimescaleDB
			_, _ = ge.db.Exec(ctx, `
				INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, location, duration_minutes)
				VALUES ($1, $2, $3, 'exit', ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, $6)
			`, time.Unix(event.Timestamp, 0), vehicleID, geofenceID, event.Longitude, event.Latitude, int(duration.Minutes()))
		}
	}

	state.LastChecked = time.Unix(event.Timestamp, 0)
	return nil
}

func (ge *GeofenceEngine) triggerGeofenceEvent(ctx context.Context, vehicleID, geofenceID uuid.UUID, eventType string, event GPSEvent) {
	// Log to TimescaleDB
	_, err := ge.db.Exec(ctx, `
		INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, location)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography)
	`, time.Unix(event.Timestamp, 0), vehicleID, geofenceID, eventType, event.Longitude, event.Latitude)

	if err != nil {
		ge.logger.Error("Failed to log geofence event", zap.Error(err))
	}

	// Publish to NATS for alert service
	eventData, _ := json.Marshal(map[string]interface{}{
		"vehicle_id":   vehicleID.String(),
		"geofence_id":  geofenceID.String(),
		"event_type":   eventType,
		"latitude":     event.Latitude,
		"longitude":    event.Longitude,
		"timestamp":    event.Timestamp,
	})

	if err := ge.nc.Publish("fleet.geofence.events", eventData); err != nil {
		ge.logger.Error("Failed to publish geofence event", zap.Error(err))
	}

	ge.logger.Info("Geofence event triggered",
		zap.String("vehicle_id", vehicleID.String()),
		zap.String("geofence_id", geofenceID.String()),
		zap.String("type", eventType))
}
