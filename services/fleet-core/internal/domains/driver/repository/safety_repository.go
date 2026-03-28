package repository

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SafetyRepository struct {
	db *pgxpool.Pool
}

func NewSafetyRepository(db *pgxpool.Pool) *SafetyRepository {
	return &SafetyRepository{db: db}
}

func (r *SafetyRepository) CreateSafetyEvent(ctx context.Context, event *models.SafetyEvent) error {
	query := `
		INSERT INTO driver_safety_events (
			id, driver_id, trip_id, event_type, severity, gps_lat, gps_lng, timestamp, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		event.ID, event.DriverID, event.TripID, event.EventType,
		event.Severity, event.GPSLat, event.GPSLng, event.Timestamp, event.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}
