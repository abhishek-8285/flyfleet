package worker

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	trModels "github.com/flyfleet/fleet-core/internal/tracking/models"
	tr "github.com/flyfleet/fleet-core/internal/tracking/repository"
	tripModels "github.com/flyfleet/fleet-core/internal/domains/trip/models"
)

// TripQueryPort defines what the auto-status worker needs from the trip domain.
// The real *tripRepo.PostgresTripRepository implements this interface.
type TripQueryPort interface {
	GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*tripModels.Trip, error)
	GetByID(ctx context.Context, id string) (*tripModels.Trip, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
}

// AutoStatusWorker consumes telemetry position updates from the Redis
// pub/sub telemetry stream and automatically transitions trip states:
//   - Auto-start:  vehicle reaches pickup geofence with ignition ON
//   - Auto-complete: vehicle arrives at drop geofence and comes to a stop
type AutoStatusWorker struct {
	redis     *redis.Client
	liveCache *tr.LiveCacheRepository
	tripQuery TripQueryPort
	logger    *zap.Logger
}

// NewAutoStatusWorker creates a new auto-status worker.
func NewAutoStatusWorker(
	rdb *redis.Client,
	liveCache *tr.LiveCacheRepository,
	tripQuery TripQueryPort,
	logger *zap.Logger,
) *AutoStatusWorker {
	return &AutoStatusWorker{
		redis:     rdb,
		liveCache: liveCache,
		tripQuery: tripQuery,
		logger:    logger,
	}
}

// Start subscribes to the telemetry pub/sub pattern (org:*:telemetry)
// and processes every position update until the context is cancelled.
func (w *AutoStatusWorker) Start(ctx context.Context) error {
	w.logger.Info("AutoStatusWorker starting — subscribing to telemetry stream")

	pubsub := w.redis.PSubscribe(ctx, "org:*:telemetry")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			w.processTelemetry(ctx, msg)
		}
	}
}

// processTelemetry handles a single position update from the telemetry stream.
func (w *AutoStatusWorker) processTelemetry(ctx context.Context, msg *redis.Message) {
	var pos trModels.PositionUpdate
	if err := json.Unmarshal([]byte(msg.Payload), &pos); err != nil {
		w.logger.Debug("failed to unmarshal telemetry event", zap.Error(err))
		return
	}

	// Extract tenant ID from channel name: org:{tenant_id}:telemetry
	tenantID := extractTenantFromChannel(msg.Channel)

	// Resolve vehicle UUID for live cache lookup
	vehicleUUID, err := uuid.FromString(pos.VehicleID)
	if err != nil {
		w.logger.Debug("invalid vehicle_id in telemetry",
			zap.String("vehicle_id", pos.VehicleID))
		return
	}

	// Get full live location (includes ignition) from Redis cache
	loc, err := w.liveCache.GetLiveLocation(ctx, vehicleUUID)
	if err != nil || loc == nil {
		// Vehicle offline or cache miss — skip this update
		return
	}

	// Look up active trips for this vehicle
	trips, err := w.tripQuery.GetActiveTripsForVehicle(ctx, pos.VehicleID)
	if err != nil {
		w.logger.Warn("failed to get active trips for vehicle",
			zap.String("vehicle_id", pos.VehicleID),
			zap.Error(err))
		return
	}

	for _, trip := range trips {
		// Defense-in-depth: skip trips from other tenants
		if trip.TenantID != tenantID {
			continue
		}

		status := strings.ToUpper(trip.Status)

		// Auto-Start: trip waiting (CREATED/SCHEDULED), vehicle near pickup,
		// ignition ON and not moving fast.
		if status == tripModels.TripStatusCreated || status == tripModels.TripStatusScheduled {
			if loc.Ignition && loc.Speed < 30 {
				fullTrip, err := w.tripQuery.GetByID(ctx, trip.ID)
				if err != nil || fullTrip == nil {
					continue
				}
				if w.isNearLocation(loc.Lat, loc.Lng, fullTrip.PickupGeo, 50) {
					if err := w.autoStartTrip(ctx, fullTrip, loc); err != nil {
						w.logger.Error("auto-start failed",
							zap.String("trip_id", trip.ID),
							zap.Error(err))
					} else {
						w.logger.Info("trip auto-started",
							zap.String("trip_id", trip.ID),
							zap.String("vehicle_id", pos.VehicleID))
					}
				}
			}
		}

		// Auto-Complete: trip in progress (STARTED/RESUMED/PAUSED),
		// vehicle near drop and stopped.
		if status == tripModels.TripStatusStarted ||
			status == tripModels.TripStatusResumed ||
			status == tripModels.TripStatusPaused {
			if loc.Speed < 2.0 {
				fullTrip, err := w.tripQuery.GetByID(ctx, trip.ID)
				if err != nil || fullTrip == nil {
					continue
				}
				if w.isNearLocation(loc.Lat, loc.Lng, fullTrip.DropGeo, 50) {
					if err := w.autoCompleteTrip(ctx, fullTrip, loc); err != nil {
						w.logger.Error("auto-complete failed",
							zap.String("trip_id", trip.ID),
							zap.Error(err))
					} else {
						w.logger.Info("trip auto-completed",
							zap.String("trip_id", trip.ID),
							zap.String("vehicle_id", pos.VehicleID))
					}
				}
			}
		}
	}
}

// isNearLocation returns true if the given lat/lng is within thresholdMeters
// of the target location (which may be nil if geo data wasn't set).
func (w *AutoStatusWorker) isNearLocation(lat, lng float64, target *tripModels.Location, thresholdMeters float64) bool {
	if target == nil {
		return false
	}
	return haversineMeters(lat, lng, target.Latitude, target.Longitude) <= thresholdMeters
}

// autoStartTrip transitions a trip from SCHEDULED → STARTED via a direct
// repository update. We use the repo directly rather than TripService.Start to
// avoid the service-layer status check that expects uppercase DB status.
func (w *AutoStatusWorker) autoStartTrip(ctx context.Context, trip *tripModels.Trip, loc *trModels.LiveLocation) error {
	now := time.Now()

	odometer := loc.Odometer
	fuelLevel := loc.FuelLevel

	return w.tripQuery.Update(ctx, trip.ID, map[string]interface{}{
		"status":              tripModels.TripStatusStarted,
		"start_time":          &now,
		"start_odometer":      odometer,
		"fuel_level_snapshot": fuelLevel,
	})
}

// autoCompleteTrip transitions a trip from STARTED → COMPLETED via a direct
// repository update, populating end-trip fields from the last telemetry point.
func (w *AutoStatusWorker) autoCompleteTrip(ctx context.Context, trip *tripModels.Trip, loc *trModels.LiveLocation) error {
	now := time.Now()

	endOdometer := loc.Odometer
	totalDistance := endOdometer
	if trip.StartOdometer != nil && *trip.StartOdometer > 0 {
		totalDistance = endOdometer - *trip.StartOdometer
		if totalDistance < 0 {
			totalDistance = 0
		}
	}

	return w.tripQuery.Update(ctx, trip.ID, map[string]interface{}{
		"status":       tripModels.TripStatusCompleted,
		"end_time":     &now,
		"end_odometer": endOdometer,
		"total_distance": totalDistance,
	})
}

// extractTenantFromChannel parses "org:{tenant_id}:telemetry" → tenant_id.
func extractTenantFromChannel(channel string) string {
	parts := strings.Split(channel, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// haversineMeters computes the great-circle distance between two coordinates
// in meters using the Haversine formula.
func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0
	radLat1 := lat1 * math.Pi / 180.0
	radLat2 := lat2 * math.Pi / 180.0
	dLat := radLat2 - radLat1
	dLon := (lng2 - lng1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
