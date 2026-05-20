package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/database"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

type geofenceChecker interface {
	CheckVehicle(ctx context.Context, vehicleID uuid.UUID, tenantID uuid.UUID, lat, lng float64) error
}

// IngestionService validates, stores, and broadcasts incoming
// telemetry events from GPS devices.
type IngestionService struct {
	telemRepo  repository.TelemetryRepositoryer
	liveCache  repository.LiveCacher
	geoChecker geofenceChecker
	batcher    *TelemetryBatcher // Added for high-throughput batching
	redis      *redis.Client
	logger     *zap.Logger
	hmacSecret string
}

func NewIngestionService(
	telemRepo repository.TelemetryRepositoryer,
	liveCache repository.LiveCacher,
	geoChecker geofenceChecker,
	rdb *redis.Client,
	logger *zap.Logger,
) *IngestionService {
	batcher := NewTelemetryBatcher(telemRepo, logger, 1000, 5*time.Second)
	batcher.SetRedisClient(rdb)
	batcher.Start()

	return &IngestionService{
		telemRepo:  telemRepo,
		liveCache:  liveCache,
		geoChecker: geoChecker,
		batcher:    batcher,
		redis:      rdb,
		logger:     logger,
	}
}

// Stop shuts down the internal batcher gracefully.
func (s *IngestionService) Stop() {
	if s.batcher != nil {
		s.batcher.Stop()
	}
}

// ForceFlush triggers an immediate database write of all buffered telemetry events.
func (s *IngestionService) ForceFlush() {
	if s.batcher != nil {
		s.batcher.ForceFlush()
	}
}

// SetHMACSecret updates the HMAC secret for ingestion validation.
func (s *IngestionService) SetHMACSecret(secret string) {
	s.hmacSecret = secret
}

// Health returns the status of the ingestion service and its dependencies.
func (s *IngestionService) Health(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})
	stats["status"] = "healthy"

	// Check Batcher
	if s.batcher != nil {
		depth, capacity := s.batcher.QueueStats()
		stats["batcher"] = map[string]interface{}{
			"queue_depth":    depth,
			"queue_capacity": capacity,
			"queue_usage":    float64(depth) / float64(capacity),
		}
		if depth > int(float64(capacity)*0.9) {
			stats["status"] = "degraded"
			stats["reason"] = "telemetry_queue_near_capacity"
		}
	}

	// Check Redis
	if err := s.redis.Ping(ctx).Err(); err != nil {
		stats["status"] = "unhealthy"
		stats["redis"] = "disconnected"
	} else {
		stats["redis"] = "connected"
	}

	return stats
}

// verifyHMAC computes HMAC-SHA256(key, message) and compares with sig.
func (s *IngestionService) verifyHMAC(message, sig string) bool {
	if s.hmacSecret == "" {
		return true
	}

	mac := hmac.New(sha256.New, []byte(s.hmacSecret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(expected), []byte(sig)) == 1
}

// Ingest validates and persists a telemetry event, updates the Redis
// live cache, and publishes a position update for WebSocket fans.
func (s *IngestionService) Ingest(ctx context.Context, event *models.TelemetryEvent) error {
	// Resolve VehicleID and TenantID if they are Nil (e.g. Protobuf/MQTT edge gateway ingestion)
	if event.VehicleID == uuid.Nil || event.TenantID == uuid.Nil {
		tenantID, vehicleID, err := s.telemRepo.ResolveDeviceAssignment(ctx, event.DeviceID)
		if err != nil {
			s.logger.Warn("Failed to resolve device assignment from database", zap.String("device_id", event.DeviceID), zap.Error(err))
			return fmt.Errorf("resolve device assignment: %w", err)
		}
		event.TenantID = tenantID
		event.VehicleID = vehicleID
	}

	// HMAC validation (if secret is configured)
	if s.hmacSecret != "" {
		if event.Signature == "" {
			return fmt.Errorf("X-HMAC-Signature is required for hardware ingestion")
		}
		// Message for signature is EventID + DeviceID
		if !s.verifyHMAC(event.EventID.String()+event.DeviceID, event.Signature) {
			return fmt.Errorf("invalid HMAC signature")
		}
	}

	// Validate coordinates
	if event.Lat < -90 || event.Lat > 90 {
		return fmt.Errorf("invalid latitude: %f", event.Lat)
	}
	if event.Lng < -180 || event.Lng > 180 {
		return fmt.Errorf("invalid longitude: %f", event.Lng)
	}

	// Write to TimescaleDB asynchronously via the batcher
	s.batcher.Add(*event)

	// Update Redis live cache
	status := models.CalculateStatus(event.Speed, event.Ignition, event.RecordedAt)
	loc := &models.LiveLocation{
		VehicleID: event.VehicleID,
		Lat:       event.Lat,
		Lng:       event.Lng,
		Speed:     event.Speed,
		Heading:   event.Heading,
		Ignition:  event.Ignition,
		FuelLevel: event.FuelLevel,
		Odometer:  event.Odometer,
		LastSeen:  event.RecordedAt,
		Status:    status,
	}
	if err := s.liveCache.SetLiveLocation(ctx, event.VehicleID, loc); err != nil {
		s.logger.Warn("failed to update live cache",
			zap.String("vehicle_id", event.VehicleID.String()),
			zap.Error(err))
		// Non-fatal — DB write succeeded
	}

	// Publish to Redis pub/sub for WebSocket fan-out
	if err := s.publishPositionUpdate(ctx, event.TenantID, event, status); err != nil {
		s.logger.Warn("failed to publish position update",
			zap.String("vehicle_id", event.VehicleID.String()),
			zap.Error(err))
		// Non-fatal — core ingest succeeded
	}

	// Trigger geofence checks
	if s.geoChecker != nil {
		if err := s.geoChecker.CheckVehicle(ctx, event.VehicleID, event.TenantID, event.Lat, event.Lng); err != nil {
			s.logger.Error("geofence check failed",
				zap.String("vehicle_id", event.VehicleID.String()),
				zap.Error(err))
		}
	}

	s.logger.Info("telemetry ingested successfully",
		zap.String("vehicle_id", event.VehicleID.String()),
		zap.String("device_id", event.DeviceID),
		zap.String("status", status),
		zap.Float64("lat", event.Lat),
		zap.Float64("lng", event.Lng),
	)
	return nil
}

func (s *IngestionService) publishPositionUpdate(
	ctx context.Context,
	tenantID uuid.UUID,
	event *models.TelemetryEvent,
	status string,
) error {
	msg, err := json.Marshal(models.PositionUpdate{
		VehicleID: event.VehicleID.String(),
		Latitude:  event.Lat,
		Longitude: event.Lng,
		SpeedKmh:  event.Speed,
		Heading:   event.Heading,
		Status:    status,
		Timestamp: event.RecordedAt.Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("marshal position update: %w", err)
	}

	channel := database.RedisChannelForOrg(tenantID.String())
	return s.redis.Publish(ctx, channel, msg).Err()
}
