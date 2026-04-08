package service

import (
	"context"
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

// IngestionService validates, stores, and broadcasts incoming
// telemetry events from GPS devices.
type IngestionService struct {
	telemRepo repository.TelemetryRepositoryer
	liveCache repository.LiveCacher
	redis     *redis.Client
	logger    *zap.Logger
}

func NewIngestionService(
	telemRepo repository.TelemetryRepositoryer,
	liveCache repository.LiveCacher,
	rdb *redis.Client,
	logger *zap.Logger,
) *IngestionService {
	return &IngestionService{
		telemRepo: telemRepo,
		liveCache: liveCache,
		redis:     rdb,
		logger:    logger,
	}
}

// Ingest validates and persists a telemetry event, updates the Redis
// live cache, and publishes a position update for WebSocket fans.
func (s *IngestionService) Ingest(ctx context.Context, event *models.TelemetryEvent) error {
	// Validate coordinates
	if event.Lat < -90 || event.Lat > 90 {
		return fmt.Errorf("invalid latitude: %f", event.Lat)
	}
	if event.Lng < -180 || event.Lng > 180 {
		return fmt.Errorf("invalid longitude: %f", event.Lng)
	}

	// Write to TimescaleDB
	if err := s.telemRepo.Write(ctx, event); err != nil {
		return fmt.Errorf("write telemetry: %w", err)
	}

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

	s.logger.Debug("telemetry ingested",
		zap.String("vehicle_id", event.VehicleID.String()),
		zap.String("status", status),
		zap.Float64("lat", event.Lat),
		zap.Float64("lng", event.Lng),
	)
	return nil
}

func (s *IngestionService) publishPositionUpdate(
	ctx context.Context,
	orgID uuid.UUID,
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

	channel := database.RedisChannelForOrg(orgID.String())
	return s.redis.Publish(ctx, channel, msg).Err()
}
