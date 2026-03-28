package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type GeofenceEngine struct {
	db     *pgxpool.Pool
	nc     *nats.Conn
	redis  *redis.Client
	logger *zap.Logger

	// Cache of vehicle last known geofence state
	vehicleStates map[uuid.UUID]*VehicleState
}

type VehicleState struct {
	VehicleID       uuid.UUID
	InsideGeofences map[uuid.UUID]time.Time // geofence_id -> entry_time
	LastChecked     time.Time
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
		redis:         nil,
		logger:        logger,
		vehicleStates: make(map[uuid.UUID]*VehicleState),
	}
}

func NewGeofenceEngineWithRedis(db *pgxpool.Pool, nc *nats.Conn, redisClient *redis.Client, logger *zap.Logger) *GeofenceEngine {
	return &GeofenceEngine{
		db:            db,
		nc:            nc,
		redis:         redisClient,
		logger:        logger,
		vehicleStates: make(map[uuid.UUID]*VehicleState),
	}
}

var (
	geofenceEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_events_total",
		Help: "Total geofence events triggered",
	}, []string{"event_type"})
	geofenceEngineProcessingTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "geofence_engine_processing_time",
		Help:    "Processing time for a single GPS message in geofence engine",
		Buckets: prometheus.DefBuckets,
	})
	geofenceEngineErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "geofence_engine_errors_total",
		Help: "Total geofence engine processing errors",
	})
	geofenceGPSMessagesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "geofence_gps_messages_processed",
		Help: "Total GPS messages processed by geofence engine",
	})
)

// Start - Subscribe to GPS events and process geofence checks
func (ge *GeofenceEngine) Start(ctx context.Context) error {
	ge.logger.Info("Starting geofence engine...")
	_ = ge.rebuildVehicleState(ctx)

	// Subscribe to GPS telemetry stream
	sub, err := ge.nc.Subscribe("fleet.telemetry.gps", func(msg *nats.Msg) {
		var event GPSEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			ge.logger.Error("Failed to parse GPS event", zap.Error(err))
			geofenceEngineErrorsTotal.Inc()
			return
		}

		vehicleID, err := uuid.Parse(event.VehicleID)
		if err != nil {
			geofenceEngineErrorsTotal.Inc()
			return
		}
		geofenceGPSMessagesProcessed.Inc()

		// Process geofence checks for this GPS point
		if err := ge.processGPSEvent(ctx, vehicleID, event); err != nil {
			geofenceEngineErrorsTotal.Inc()
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
	start := time.Now()
	defer func() {
		geofenceEngineProcessingTime.Observe(time.Since(start).Seconds())
	}()
	ctx, span := otel.Tracer("geofencing.engine").Start(ctx, "process_gps_event")
	span.SetAttributes(attribute.String("vehicle_id", vehicleID.String()))
	defer span.End()

	// Query all active geofences assigned to this vehicle
	rows, err := ge.db.Query(ctx, `
		SELECT g.id, g.name, g.type, g.center, g.radius_meters, g.coordinates,
		       ga.assignment_type, ga.dwell_minutes
		FROM geofences g
		LEFT JOIN geofence_assignments ga ON ga.geofence_id = g.id
		WHERE ga.vehicle_id = $1
		  AND ga.is_active = true
		  AND g.enabled = true
	`, vehicleID)

	if err != nil {
		return fmt.Errorf("failed to query geofences: %w", err)
	}
	defer rows.Close()

	point := fmt.Sprintf("ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography", event.Longitude, event.Latitude)

	// Get or create vehicle state (Redis-backed for multi-instance safety).
	state, exists := ge.vehicleStates[vehicleID]
	if !exists {
		if loaded, lerr := ge.loadVehicleState(ctx, vehicleID); lerr == nil && loaded != nil {
			state = loaded
		}
	}
	if state == nil {
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
			center         string
			radiusMeters   *int
			coordinates    string
			assignmentType string
			dwellMinutes   *int
		)

		if err := rows.Scan(&geofenceID, &name, &geoType, &center, &radiusMeters, &coordinates, &assignmentType, &dwellMinutes); err != nil {
			ge.logger.Error("Failed to scan geofence", zap.Error(err))
			continue
		}

		// Check if vehicle is inside this geofence using PostGIS
		isInside, err := ge.checkGeofenceContainment(ctx, geoType, point, center, radiusMeters, coordinates)
		if err != nil {
			ge.logger.Error("Failed to check geofence containment", zap.Error(err))
			continue
		}

		if isInside {
			currentlyInside[geofenceID] = true

			// Check for ENTER event
			if _, wasInside := state.InsideGeofences[geofenceID]; !wasInside {
				if assignmentType == "enter" || assignmentType == "both" {
					ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, name, "enter", event)
				}
				state.InsideGeofences[geofenceID] = time.Unix(event.Timestamp, 0)
			} else {
				// Check for DWELL event
				if assignmentType == "dwell" && dwellMinutes != nil {
					entryTime := state.InsideGeofences[geofenceID]
					duration := time.Unix(event.Timestamp, 0).Sub(entryTime)
					if duration >= time.Duration(*dwellMinutes)*time.Minute {
						ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, name, "dwell", event)
						// Reset entry time to avoid duplicate dwell alerts
						state.InsideGeofences[geofenceID] = time.Unix(event.Timestamp, 0)
					}
				}
			}
		}
	}

	// Check for EXIT events
	// Pre-fetch all geofence names in batch to avoid N+1 queries
	geofenceIDs := make([]uuid.UUID, 0, len(state.InsideGeofences))
	for geofenceID := range state.InsideGeofences {
		geofenceIDs = append(geofenceIDs, geofenceID)
	}
	geofenceNames, err := ge.getGeofenceNamesBatch(ctx, geofenceIDs)
	if err != nil {
		ge.logger.Error("Failed to batch fetch geofence names", zap.Error(err))
		// Fall back to individual queries if batch fails
		geofenceNames = make(map[uuid.UUID]string)
	}

	// Collect exit events for batch insert
	type exitEvent struct {
		geofenceID uuid.UUID
		duration   int
	}
	var exits []exitEvent

	for geofenceID, entryTime := range state.InsideGeofences {
		if !currentlyInside[geofenceID] {
			duration := time.Unix(event.Timestamp, 0).Sub(entryTime)

			// Get geofence name from batch (no fallback to avoid N+1)
			geofenceName := geofenceNames[geofenceID]

			ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, geofenceName, "exit", event)
			delete(state.InsideGeofences, geofenceID)

			// Collect for batch insert instead of immediate insert
			exits = append(exits, exitEvent{
				geofenceID: geofenceID,
				duration:   int(duration.Minutes()),
			})
		}
	}

	// Batch insert exit events to avoid N+1 INSERTs
	if len(exits) > 0 {
		var args []interface{}
		var placeholders []string
		for i, ex := range exits {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, 'exit', ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
			args = append(args,
				time.Unix(event.Timestamp, 0), vehicleID, ex.geofenceID,
				event.Longitude, event.Latitude, ex.duration,
			)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, location, duration_minutes) VALUES %s ON CONFLICT DO NOTHING",
			strings.Join(placeholders, ", "),
		)
		_, _ = ge.db.Exec(ctx, batchQuery, args...)
	}

	state.LastChecked = time.Unix(event.Timestamp, 0)
	ge.vehicleStates[vehicleID] = state
	_ = ge.saveVehicleState(ctx, state)
	return nil
}

// checkGeofenceContainment checks if a point is inside a geofence
func (ge *GeofenceEngine) checkGeofenceContainment(ctx context.Context, geoType, point, center string, radiusMeters *int, coordinates string) (bool, error) {
	ctx, span := otel.Tracer("geofencing.engine").Start(ctx, "check_geofence_containment")
	defer span.End()
	var query string
	var args []interface{}

	switch geoType {
	case "circle":
		query = fmt.Sprintf(`SELECT ST_DWithin(%s::geography, $1::geography, $2)`, center)
		args = []interface{}{point, *radiusMeters}
	case "polygon":
		query = fmt.Sprintf(`SELECT ST_Within(%s::geography, $1::geography)`, point)
		args = []interface{}{coordinates}
	default:
		return false, nil
	}

	var isInside bool
	err := ge.db.QueryRow(ctx, query, args...).Scan(&isInside)
	if err != nil {
		return false, err
	}

	return isInside, nil
}

// getGeofenceName retrieves the name of a geofence by ID
func (ge *GeofenceEngine) getGeofenceName(ctx context.Context, geofenceID uuid.UUID) (string, error) {
	var name string
	err := ge.db.QueryRow(ctx, `SELECT name FROM geofences WHERE id = $1`, geofenceID).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

// getGeofenceNamesBatch retrieves names for multiple geofences in a single query
func (ge *GeofenceEngine) getGeofenceNamesBatch(ctx context.Context, geofenceIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(geofenceIDs) == 0 {
		return make(map[uuid.UUID]string), nil
	}

	// Build query with UNNEST for array parameter
	query := `SELECT id, name FROM geofences WHERE id = ANY($1)`

	rows, err := ge.db.Query(ctx, query, geofenceIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make(map[uuid.UUID]string, len(geofenceIDs))
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		names[id] = name
	}

	return names, nil
}

func (ge *GeofenceEngine) triggerGeofenceEvent(ctx context.Context, vehicleID, geofenceID uuid.UUID, geofenceName, eventType string, event GPSEvent) {
	ctx, span := otel.Tracer("geofencing.engine").Start(ctx, "trigger_geofence_event")
	span.SetAttributes(
		attribute.String("vehicle_id", vehicleID.String()),
		attribute.String("geofence_id", geofenceID.String()),
		attribute.String("event_type", eventType),
	)
	defer span.End()

	// Log to database
	_, err := ge.db.Exec(ctx, `
		INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, location)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography)
		ON CONFLICT DO NOTHING
	`, time.Unix(event.Timestamp, 0), vehicleID, geofenceID, eventType, event.Longitude, event.Latitude)

	if err != nil {
		ge.logger.Error("Failed to log geofence event", zap.Error(err))
		geofenceEngineErrorsTotal.Inc()
		return
	}

	// Publish to NATS for alert service
	eventData, _ := json.Marshal(map[string]interface{}{
		"vehicle_id":    vehicleID.String(),
		"geofence_id":   geofenceID.String(),
		"geofence_name": geofenceName,
		"event_type":    eventType,
		"latitude":      event.Latitude,
		"longitude":     event.Longitude,
		"timestamp":     event.Timestamp,
	})

	if ge.nc != nil {
		if err := ge.nc.Publish("fleet.geofence.events", eventData); err != nil {
			ge.logger.Error("Failed to publish geofence event", zap.Error(err))
		}
	}
	geofenceEventsTotal.WithLabelValues(eventType).Inc()

	ge.logger.Info("Geofence event triggered",
		zap.String("vehicle_id", vehicleID.String()),
		zap.String("geofence_id", geofenceID.String()),
		zap.String("type", eventType))
}

func (ge *GeofenceEngine) stateKey(vehicleID uuid.UUID) string {
	return "vehicle_state:" + vehicleID.String()
}

func (ge *GeofenceEngine) loadVehicleState(ctx context.Context, vehicleID uuid.UUID) (*VehicleState, error) {
	if ge.redis == nil {
		return nil, nil
	}
	raw, err := ge.redis.Get(ctx, ge.stateKey(vehicleID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var state VehicleState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, err
	}
	if state.InsideGeofences == nil {
		state.InsideGeofences = make(map[uuid.UUID]time.Time)
	}
	return &state, nil
}

func (ge *GeofenceEngine) saveVehicleState(ctx context.Context, state *VehicleState) error {
	if ge.redis == nil || state == nil {
		return nil
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return ge.redis.Set(ctx, ge.stateKey(state.VehicleID), string(raw), 30*24*time.Hour).Err()
}

func (ge *GeofenceEngine) rebuildVehicleState(ctx context.Context) error {
	if ge.redis == nil {
		return nil
	}
	rows, err := ge.db.Query(ctx, `
		SELECT vehicle_id, geofence_id, event_type, time
		FROM geofence_events
		WHERE time >= NOW() - INTERVAL '24 hours'
		ORDER BY vehicle_id, time ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stateMap := map[uuid.UUID]*VehicleState{}
	for rows.Next() {
		var vehicleID, geofenceID uuid.UUID
		var eventType string
		var ts time.Time
		if err := rows.Scan(&vehicleID, &geofenceID, &eventType, &ts); err != nil {
			continue
		}
		st := stateMap[vehicleID]
		if st == nil {
			st = &VehicleState{
				VehicleID:       vehicleID,
				InsideGeofences: make(map[uuid.UUID]time.Time),
				LastChecked:     ts,
			}
			stateMap[vehicleID] = st
		}
		switch eventType {
		case "enter", "dwell":
			st.InsideGeofences[geofenceID] = ts
		case "exit":
			delete(st.InsideGeofences, geofenceID)
		}
		if ts.After(st.LastChecked) {
			st.LastChecked = ts
		}
	}
	for vehicleID, state := range stateMap {
		ge.vehicleStates[vehicleID] = state
		_ = ge.saveVehicleState(ctx, state)
	}
	return nil
}
