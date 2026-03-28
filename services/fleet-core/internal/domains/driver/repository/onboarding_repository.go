package repository

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OnboardingRepository struct {
	db *pgxpool.Pool
}

func NewOnboardingRepository(db *pgxpool.Pool) *OnboardingRepository {
	return &OnboardingRepository{db: db}
}

func (r *OnboardingRepository) CreateStep(ctx context.Context, step *models.OnboardingStep) error {
	query := `
		INSERT INTO driver_onboarding_steps (
			id, driver_id, step, status, remarks, completed_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if step.ID == "" {
		step.ID = uuid.New().String()
	}
	if step.CreatedAt.IsZero() {
		step.CreatedAt = time.Now()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		step.ID, step.DriverID, step.Step, step.Status,
		step.Remarks, step.CompletedAt, step.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *OnboardingRepository) UpdateStep(ctx context.Context, step *models.OnboardingStep) error {
	query := `
		UPDATE driver_onboarding_steps
		SET status = $2, remarks = $3, completed_at = $4
		WHERE id = $1`

	start := time.Now()
	_, err := r.db.Exec(ctx, query, step.ID, step.Status, step.Remarks, step.CompletedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *OnboardingRepository) GetStepsByDriverID(ctx context.Context, driverID string) ([]*models.OnboardingStep, error) {
	query := `
		SELECT id, driver_id, step, status, remarks, completed_at, created_at
		FROM driver_onboarding_steps
		WHERE driver_id = $1
		ORDER BY created_at ASC`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []*models.OnboardingStep
	for rows.Next() {
		var s models.OnboardingStep
		err := rows.Scan(&s.ID, &s.DriverID, &s.Step, &s.Status, &s.Remarks, &s.CompletedAt, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		steps = append(steps, &s)
	}
	return steps, nil
}
