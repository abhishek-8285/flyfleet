package postgres

import (
	"context"
	"fmt"

	"github.com/flyfleet/maintenance-service/internal/domain/maintenance_schedule"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MaintenanceScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewMaintenanceScheduleRepository(pool *pgxpool.Pool) *MaintenanceScheduleRepository {
	return &MaintenanceScheduleRepository{pool: pool}
}

func (r *MaintenanceScheduleRepository) Create(ctx context.Context, s *maintenance_schedule.MaintenanceSchedule) error {
	query := `
		INSERT INTO maintenance_schedules (
			org_id, vehicle_id, type, name, description, due_date, due_odometer,
			status, priority, assigned_to, estimated_cost, parts_required, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		s.OrgID, s.VehicleID, s.Type, s.Name, s.Description, s.DueDate, s.DueOdometer,
		s.Status, s.Priority, s.AssignedTo, s.EstimatedCost, s.PartsRequired, s.Notes,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create maintenance schedule: %w", err)
	}
	return nil
}

func (r *MaintenanceScheduleRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*maintenance_schedule.MaintenanceSchedule, error) {
	query := `
		SELECT id, org_id, vehicle_id, type, name, description, due_date, due_odometer,
		       status, priority, assigned_to, estimated_cost, actual_cost, parts_required,
		       notes, completed_at, completed_by, created_at, updated_at
		FROM maintenance_schedules
		WHERE id = $1 AND org_id = $2 AND is_deleted = false
	`
	s := &maintenance_schedule.MaintenanceSchedule{}
	err := r.pool.QueryRow(ctx, query, id, orgID).Scan(
		&s.ID, &s.OrgID, &s.VehicleID, &s.Type, &s.Name, &s.Description, &s.DueDate, &s.DueOdometer,
		&s.Status, &s.Priority, &s.AssignedTo, &s.EstimatedCost, &s.ActualCost, &s.PartsRequired,
		&s.Notes, &s.CompletedAt, &s.CompletedBy, &s.CreatedAt, &s.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Not found handled in service
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance schedule: %w", err)
	}
	return s, nil
}

func (r *MaintenanceScheduleRepository) Update(ctx context.Context, s *maintenance_schedule.MaintenanceSchedule) error {
	query := `
		UPDATE maintenance_schedules SET
			type = $3, name = $4, description = $5, due_date = $6, due_odometer = $7,
			status = $8, priority = $9, assigned_to = $10, estimated_cost = $11,
			actual_cost = $12, parts_required = $13, notes = $14,
			completed_at = $15, completed_by = $16, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND org_id = $2 AND is_deleted = false
		RETURNING updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		s.ID, s.OrgID, s.Type, s.Name, s.Description, s.DueDate, s.DueOdometer,
		s.Status, s.Priority, s.AssignedTo, s.EstimatedCost, s.ActualCost, s.PartsRequired,
		s.Notes, s.CompletedAt, s.CompletedBy,
	).Scan(&s.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update maintenance schedule: %w", err)
	}
	return nil
}

func (r *MaintenanceScheduleRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	query := `
		UPDATE maintenance_schedules
		SET is_deleted = true, deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND org_id = $2 AND is_deleted = false
	`
	_, err := r.pool.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to soft delete maintenance schedule: %w", err)
	}
	return nil
}

func (r *MaintenanceScheduleRepository) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*maintenance_schedule.MaintenanceSchedule, int64, error) {
	countQuery := `SELECT COUNT(*) FROM maintenance_schedules WHERE org_id = $1 AND is_deleted = false`
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count maintenance schedules: %w", err)
	}

	query := `
		SELECT id, org_id, vehicle_id, type, name, description, due_date, due_odometer,
		       status, priority, assigned_to, estimated_cost, actual_cost, parts_required,
		       notes, completed_at, completed_by, created_at, updated_at
		FROM maintenance_schedules
		WHERE org_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list maintenance schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*maintenance_schedule.MaintenanceSchedule
	for rows.Next() {
		s := &maintenance_schedule.MaintenanceSchedule{}
		err := rows.Scan(
			&s.ID, &s.OrgID, &s.VehicleID, &s.Type, &s.Name, &s.Description, &s.DueDate, &s.DueOdometer,
			&s.Status, &s.Priority, &s.AssignedTo, &s.EstimatedCost, &s.ActualCost, &s.PartsRequired,
			&s.Notes, &s.CompletedAt, &s.CompletedBy, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan maintenance schedule: %w", err)
		}
		schedules = append(schedules, s)
	}

	return schedules, total, nil
}

