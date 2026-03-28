package repository

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShiftRepository struct {
	db *pgxpool.Pool
}

func NewShiftRepository(db *pgxpool.Pool) *ShiftRepository {
	return &ShiftRepository{db: db}
}

func (r *ShiftRepository) CreateShift(ctx context.Context, shift *models.Shift) error {
	query := `
		INSERT INTO driver_shifts (
			id, driver_id, shift_start, shift_end, shift_type, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if shift.ID == "" {
		shift.ID = uuid.New().String()
	}
	if shift.CreatedAt.IsZero() {
		shift.CreatedAt = time.Now()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		shift.ID, shift.DriverID, shift.ShiftStart, shift.ShiftEnd,
		shift.ShiftType, shift.Status, shift.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *ShiftRepository) AssignVehicle(ctx context.Context, shiftID string, vehicleID string) error {
	query := `
		INSERT INTO driver_shift_assignments (
			id, shift_id, vehicle_id, assigned_at
		) VALUES ($1, $2, $3, $4)`

	id := uuid.New().String()
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id, shiftID, vehicleID, time.Now())
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *ShiftRepository) UpdateShiftStatus(ctx context.Context, shiftID string, status string) error {
	query := `UPDATE driver_shifts SET status = $2 WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, shiftID, status)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *ShiftRepository) GetShiftsByDriverID(ctx context.Context, driverID string, start, end time.Time) ([]*models.Shift, error) {
	query := `
		SELECT id, driver_id, shift_start, shift_end, shift_type, status, created_at
		FROM driver_shifts
		WHERE driver_id = $1 AND shift_start >= $2 AND shift_start <= $3
		ORDER BY shift_start ASC`

	queryStart := time.Now()
	rows, err := r.db.Query(ctx, query, driverID, start, end)
	metrics.RecordDBQuery(query, time.Since(queryStart))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*models.Shift
	for rows.Next() {
		var s models.Shift
		err := rows.Scan(&s.ID, &s.DriverID, &s.ShiftStart, &s.ShiftEnd, &s.ShiftType, &s.Status, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, &s)
	}
	return shifts, nil
}
