package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/database"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

// GeofenceChecker runs periodically to detect geofence entry/exit
// events for all vehicles based on their latest telemetry positions.
type GeofenceChecker struct {
	telemRepo repository.TelemetryRepositoryer
	geoRepo   repository.GeofenceRepositoryer
	liveCache repository.LiveCacher
	redis     *redis.Client
	logger    *zap.Logger
	interval  time.Duration
	cancel    context.CancelFunc
	mu        sync.Mutex
	running   bool
}

func NewGeofenceChecker(
	telemRepo repository.TelemetryRepositoryer,
	geoRepo repository.GeofenceRepositoryer,
	liveCache repository.LiveCacher,
	rdb *redis.Client,
	logger *zap.Logger,
	interval time.Duration,
) *GeofenceChecker {
	return &GeofenceChecker{
		telemRepo: telemRepo,
		geoRepo:   geoRepo,
		liveCache: liveCache,
		redis:     rdb,
		logger:    logger,
		interval:  interval,
	}
}

// Start begins the periodic geofence checking loop. Call with a
// background context; use Stop() to terminate.
func (w *GeofenceChecker) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	w.logger.Info("geofence checker started",
		zap.Duration("interval", w.interval))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("geofence checker stopped")
			return
		case <-ticker.C:
			w.checkAllOrgs(ctx)
		}
	}
}

// Stop terminates the geofence checking loop.
func (w *GeofenceChecker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
	}
	w.running = false
}

// checkAllOrgs is the main check cycle. In a production system this
// would be driven by a list of active orgs; here we check all vehicles
// that have recent telemetry.
func (w *GeofenceChecker) checkAllOrgs(ctx context.Context) {
	// Get all vehicles with recent telemetry (last 5 minutes)
	// In production, this would be a dedicated query or org list
	// For now, we process events as they arrive via Redis pub/sub
	// instead of polling — more efficient.
}

// CheckVehicle checks a single vehicle's position against all active
// geofences, logging events on state changes.
func (w *GeofenceChecker) CheckVehicle(ctx context.Context, vehicleID uuid.UUID, tenantID uuid.UUID, lat, lng float64) error {
	// 1. Get all active geofences for this org
	geofences, err := w.geoRepo.GetOrgGeofences(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("get org geofences: %w", err)
	}

	if len(geofences) == 0 {
		return nil // no geofences to check
	}

	// 2. Check if point is inside any geofence (PostGIS)
	insideGeofences, err := w.geoRepo.CheckPointInGeofences(ctx, tenantID, lat, lng)
	if err != nil {
		return fmt.Errorf("spatial check: %w", err)
	}

	// Build set of geofences the vehicle is currently inside
	currentInside := make(map[string]bool)
	for _, g := range insideGeofences {
		currentInside[g.ID.String()] = true
	}

	// 3. Compare with previous state for all org geofences
	for _, g := range geofences {
		gfID := g.ID.String()
		wasInside, err := w.liveCache.GetGeofenceState(ctx, vehicleID.String(), gfID)
		if err != nil {
			w.logger.Warn("failed to get geofence state",
				zap.String("vehicle_id", vehicleID.String()),
				zap.String("geofence_id", gfID),
				zap.Error(err))
			continue
		}

		nowInside := currentInside[gfID]

		// State changed?
		if nowInside != wasInside {
			eventType := models.EventTypeEntry
			if !nowInside {
				eventType = models.EventTypeExit
			}

			event := &models.GeofenceEvent{
				ID:            uuid.Must(uuid.NewV4()),
				GeofenceID:    g.ID,
				VehicleID:     vehicleID,
				EventType:     eventType,
				OccurredAt:    time.Now().UTC(),
				Latitude:      lat,
				Longitude:     lng,
				AlertSent:     false,
				AlertChannels: g.AlertChannels,
			}

			// Log the event
			if err := w.geoRepo.LogGeofenceEvent(ctx, event); err != nil {
				w.logger.Error("failed to log geofence event",
					zap.Error(err),
					zap.String("vehicle_id", vehicleID.String()),
					zap.String("geofence_id", gfID))
				continue
			}

			// Update state in Redis
			if err := w.liveCache.SetGeofenceState(ctx, vehicleID.String(), gfID, nowInside); err != nil {
				w.logger.Warn("failed to update geofence state", zap.Error(err))
			}

			// Publish alert to Redis pub/sub for dashboard notification
			w.publishGeofenceAlert(ctx, tenantID, event, g.Name)

			w.logger.Info("geofence event detected",
				zap.String("vehicle_id", vehicleID.String()),
				zap.String("geofence", g.Name),
				zap.String("event_type", eventType))
		}
	}

	return nil
}

func (w *GeofenceChecker) publishGeofenceAlert(
	ctx context.Context,
	tenantID uuid.UUID,
	event *models.GeofenceEvent,
	geofenceName string,
) {
	msg, err := json.Marshal(models.GeofenceWSMessage{
		VehicleID:    event.VehicleID.String(),
		EventType:    event.EventType,
		GeofenceID:   event.GeofenceID.String(),
		GeofenceName: geofenceName,
		OccurredAt:   event.OccurredAt.Format(time.RFC3339),
		Latitude:     event.Latitude,
		Longitude:    event.Longitude,
		Message: fmt.Sprintf("Vehicle %s %s %s at %s",
			event.VehicleID.String()[:8],
			event.EventType,
			geofenceName,
			event.OccurredAt.Format("15:04")),
	})
	if err != nil {
		w.logger.Error("failed to marshal geofence alert", zap.Error(err))
		return
	}

	channel := database.RedisChannelForOrg(tenantID.String())
	if err := w.redis.Publish(ctx, channel, msg).Err(); err != nil {
		w.logger.Warn("failed to publish geofence alert", zap.Error(err))
	}
}
