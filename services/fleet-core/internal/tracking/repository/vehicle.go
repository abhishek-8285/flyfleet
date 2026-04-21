package repository

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

// VehicleRepository provides vehicle and driver data for
// enriching tracking responses.
type VehicleRepository struct {
	pool *pgxpool.Pool
}

func NewVehicleRepository(pool *pgxpool.Pool) *VehicleRepository {
	return &VehicleRepository{pool: pool}
}

// VehicleRecord is the DB row for a vehicle.
type VehicleRecord struct {
	ID              uuid.UUID `json:"id"`
	Registration    string    `json:"registration"`
	Make            string    `json:"make"`
	Model           string    `json:"model"`
	CurrentDriverID uuid.UUID `json:"current_driver_id"`
}

// GetOrgVehicles returns all active vehicles for an organization.
func (r *VehicleRepository) GetOrgVehicles(ctx context.Context, tenantID uuid.UUID) ([]VehicleRecord, error) {
	q := `SELECT id, registration, make, model, COALESCE(current_driver_id, '00000000-0000-0000-0000-000000000000')
	      FROM vehicles
	      WHERE tenant_id = $1 AND status = 'active'
	      ORDER BY registration`

	rows, err := r.pool.Query(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query org vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []VehicleRecord
	for rows.Next() {
		var v VehicleRecord
		if err := rows.Scan(&v.ID, &v.Registration, &v.Make, &v.Model, &v.CurrentDriverID); err != nil {
			return nil, fmt.Errorf("scan vehicle: %w", err)
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

// GetDriverInfo returns a driver's name and phone by ID.
func (r *VehicleRepository) GetDriverInfo(ctx context.Context, driverID uuid.UUID) (*models.DriverInfo, error) {
	q := `SELECT id, name, COALESCE(phone, '') FROM drivers WHERE id = $1`
	var d models.DriverInfo
	err := r.pool.QueryRow(ctx, q, driverID).Scan(&d.ID, &d.Name, &d.Phone)
	if err != nil {
		return nil, fmt.Errorf("query driver info: %w", err)
	}
	return &d, nil
}

// GetDriversBatch returns driver info for multiple driver IDs.
func (r *VehicleRepository) GetDriversBatch(ctx context.Context, driverIDs []uuid.UUID) (map[uuid.UUID]*models.DriverInfo, error) {
	if len(driverIDs) == 0 {
		return map[uuid.UUID]*models.DriverInfo{}, nil
	}

	q := `SELECT id, name, COALESCE(phone, '') FROM drivers WHERE id = ANY($1)`
	rows, err := r.pool.Query(ctx, q, driverIDs)
	if err != nil {
		return nil, fmt.Errorf("batch query drivers: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*models.DriverInfo)
	for rows.Next() {
		var d models.DriverInfo
		if err := rows.Scan(&d.ID, &d.Name, &d.Phone); err != nil {
			return nil, fmt.Errorf("scan driver: %w", err)
		}
		id, _ := uuid.FromString(d.ID)
		result[id] = &d
	}
	return result, nil
}
