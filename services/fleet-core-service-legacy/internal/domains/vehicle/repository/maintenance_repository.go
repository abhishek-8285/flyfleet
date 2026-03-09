package repository

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MaintenanceScheduleRepository struct {
	db *pgxpool.Pool
}

func NewMaintenanceScheduleRepository(db *pgxpool.Pool) *MaintenanceScheduleRepository {
	return &MaintenanceScheduleRepository{db: db}
}

func (r *MaintenanceScheduleRepository) Upsert(ctx context.Context, s *models.MaintenanceSchedule) error {
	query := `
		INSERT INTO maintenance_schedule (
			vehicle_id, next_service_date, next_service_odometer, engine_hours,
			predictive_score, last_service_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (vehicle_id) DO UPDATE SET
			next_service_date     = EXCLUDED.next_service_date,
			next_service_odometer = EXCLUDED.next_service_odometer,
			engine_hours          = EXCLUDED.engine_hours,
			predictive_score      = EXCLUDED.predictive_score,
			last_service_date     = EXCLUDED.last_service_date,
			updated_at            = NOW()
		RETURNING created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		s.VehicleID, s.NextServiceDate, s.NextServiceOdometer,
		s.EngineHours, s.PredictiveScore, s.LastServiceDate,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (r *MaintenanceScheduleRepository) Get(ctx context.Context, vehicleID uuid.UUID) (*models.MaintenanceSchedule, error) {
	s := &models.MaintenanceSchedule{}
	err := r.db.QueryRow(ctx, `
		SELECT vehicle_id, next_service_date, next_service_odometer, engine_hours,
		       predictive_score, last_service_date, created_at, updated_at
		FROM maintenance_schedule WHERE vehicle_id = $1`, vehicleID,
	).Scan(
		&s.VehicleID, &s.NextServiceDate, &s.NextServiceOdometer,
		&s.EngineHours, &s.PredictiveScore, &s.LastServiceDate,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return s, nil
}
