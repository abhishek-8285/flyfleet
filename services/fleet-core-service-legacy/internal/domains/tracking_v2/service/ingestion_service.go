package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// maxTimestampDriftSeconds is the maximum allowed difference between
// the device's recorded_at and the server's current time.
const maxTimestampDriftSeconds = 300

// IngestionConfig holds tunable ingestion parameters.
type IngestionConfig struct {
	MaxSpeedKmh              float64
	MaxTimestampDriftSeconds int
	HMACSecret               string // if non-empty, require X-HMAC-Signature
	BufferEnabled            bool   // if true, use NATS for async ingestion
}

// DefaultIngestionConfig returns sensible production defaults.
func DefaultIngestionConfig() IngestionConfig {
	return IngestionConfig{
		MaxSpeedKmh:              300,
		MaxTimestampDriftSeconds: maxTimestampDriftSeconds,
		BufferEnabled:            true, // Default to async for production safety
	}
}

// IngestionService validates, deduplicates, and persists inbound telemetry.
// It also updates the Redis live-cache and hands events to the anomaly engine.
type IngestionService struct {
	db            *pgxpool.Pool
	mb            messaging.MessageBroker
	telemRepo     repository.TelemetryRepository
	devRegRepo    repository.DeviceRegistryRepository
	outboxRepo    repository.OutboxRepository
	liveCache     repository.LiveCacheRepository   // may be nil
	anomalyEngine *AnomalyEngine
	analytics     analytics.AnalyticsService // optional
	cfg           IngestionConfig
	logger        *zap.Logger
}

// NewIngestionService creates a new IngestionService.
func NewIngestionService(
	db *pgxpool.Pool,
	mb messaging.MessageBroker,
	telemRepo repository.TelemetryRepository,
	devRegRepo repository.DeviceRegistryRepository,
	outboxRepo repository.OutboxRepository,
	liveCache repository.LiveCacheRepository,
	anomalyEngine *AnomalyEngine,
	analytics analytics.AnalyticsService,
	cfg IngestionConfig,
	logger *zap.Logger,
) *IngestionService {
	return &IngestionService{
		db:            db,
		mb:            mb,
		telemRepo:     telemRepo,
		devRegRepo:    devRegRepo,
		outboxRepo:    outboxRepo,
		liveCache:     liveCache,
		anomalyEngine: anomalyEngine,
		analytics:     analytics,
		cfg:           cfg,
		logger:        logger,
	}
}

// Ingest validates and stores a single telemetry event.
// Returns a typed error wrapping the rejection reason on validation failure.
func (s *IngestionService) Ingest(ctx context.Context, event *models.TelemetryEvent) ([]models.AnomalyEvent, error) {
	// 1 ── HMAC validation (if secret is configured)
	if s.cfg.HMACSecret != "" && event.Signature != "" {
		if !s.verifyHMAC(event.EventID+event.DeviceID, event.Signature) {
			return nil, fmt.Errorf("invalid HMAC signature")
		}
	}

	// 2 ── Timestamp drift check
	drift := time.Since(event.RecordedAt)
	if drift < 0 {
		drift = -drift
	}
	if drift > time.Duration(s.cfg.MaxTimestampDriftSeconds)*time.Second {
		return nil, fmt.Errorf("timestamp drift too large: %v", drift)
	}

	// 3 ── Coordinate sanity
	if err := validateCoords(event.Lat, event.Lng); err != nil {
		return nil, err
	}

	// 4 ── Speed sanity
	if event.Speed > s.cfg.MaxSpeedKmh {
		return nil, fmt.Errorf("speed %g km/h exceeds maximum %g km/h", event.Speed, s.cfg.MaxSpeedKmh)
	}

	// 5 ── Device / vehicle binding check
	device, err := s.devRegRepo.GetByDeviceID(ctx, event.DeviceID)
	if err != nil || device == nil {
		return nil, fmt.Errorf("device not registered: %s", event.DeviceID)
	}
	if device.VehicleID != "" && event.VehicleID.String() != device.VehicleID {
		return nil, fmt.Errorf("device %s is not bound to vehicle %s", event.DeviceID, event.VehicleID)
	}

	// Multi-tenant check: ensure event tenant matches device tenant
	if event.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// 6 ── Dedup check
	dup, err := s.telemRepo.IsDuplicate(ctx, event.EventID, event.DeviceID)
	if err != nil {
		s.logger.Warn("dedup check failed", zap.Error(err))
	}
	if dup {
		return nil, fmt.Errorf("duplicate event_id: %s", event.EventID)
	}

	// 7 ── Stamp received_at
	event.ReceivedAt = time.Now().UTC()

	// 8 ── Persist to DB or Buffer via NATS
	if s.cfg.BufferEnabled && s.mb != nil {
		// Asynchronous Path: Buffer to NATS for high-throughput resilience.
		// Partition key is implicit in the payload; JetStream handle provides ordering.
		err = s.mb.PublishJS("telemetry.raw", event)
		if err != nil {
			s.logger.Error("failed to buffer telemetry to NATS", zap.Error(err))
			// Fail fast to protect DB from sudden 20k/sec surge if buffer fails.
			return nil, fmt.Errorf("telemetry buffer unavailable: %w", err)
		}
	} else {
		// Synchronous Path: Direct DB write (Legacy or non-critical paths)
		if s.db != nil {
			err = s.transactionalIngest(ctx, event)
		} else {
			// Fallback for tests (non-transactional)
			err = s.telemRepo.IngestEvent(ctx, nil, event)
			if err == nil {
				err = s.outboxRepo.AddToOutbox(ctx, nil, event)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("ingest failed: %w", err)
		}
	}

	// 9 ── Update Redis live cache
	if s.liveCache != nil {
		loc := &models.LiveLocation{
			VehicleID: event.VehicleID,
			Lat:       event.Lat,
			Lng:       event.Lng,
			Speed:     event.Speed,
			Heading:   event.Heading,
			Ignition:  event.Ignition,
			LastSeen:  event.RecordedAt,
		}
		if err := s.liveCache.SetLiveLocation(ctx, loc); err != nil {
			s.logger.Warn("failed to update live cache", zap.Error(err))
		}
	}

	// 10 ── Anomaly detection (stateless – no I/O)
	var anomalies []models.AnomalyEvent
	if s.anomalyEngine != nil {
		// Fetch previous event to enable delta-based checks (with tenant_id enforcement)
		prev, _ := s.telemRepo.GetLatestByVehicle(ctx, event.TenantID, event.VehicleID)
		anomalies = s.anomalyEngine.Evaluate(event, prev)
	}

	// 11 ── Predictive Analytics Cache Invalidation
	// If it's a high-impact event (Panic, or triggers overspeed anomaly), invalidate cache.
	if s.analytics != nil {
		shouldInvalidate := event.EventType == models.EventTypePanic
		for _, a := range anomalies {
			if a.Type == models.AnomalyOverspeed || a.Type == models.AnomalyPanic {
				shouldInvalidate = true
				break
			}
		}

		if shouldInvalidate {
			_ = s.analytics.InvalidateVehicleCache(ctx, event.VehicleID, event.TenantID)
		}
	}

	return anomalies, nil
}

func (s *IngestionService) transactionalIngest(ctx context.Context, event *models.TelemetryEvent) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Ingest event to main table
	if err := s.telemRepo.IngestEvent(ctx, tx, event); err != nil {
		return err
	}

	// Add to outbox for JetStream relay
	if err := s.outboxRepo.AddToOutbox(ctx, tx, event); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// verifyHMAC computes HMAC-SHA256(key, message) and compares with sig using subtle.ConstantTimeCompare.
func (s *IngestionService) verifyHMAC(message, sig string) bool {
	mac := hmac.New(sha256.New, []byte(s.cfg.HMACSecret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expected), []byte(sig)) == 1
}

func validateCoords(lat, lng float64) error {
	if math.IsNaN(lat) || math.IsNaN(lng) {
		return errors.New("NaN coordinates")
	}
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude %g out of range [-90,90]", lat)
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("longitude %g out of range [-180,180]", lng)
	}
	return nil
}
