package repository

import (
	"context"
	"fmt"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type VehicleRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewVehicleRepository(db *pgxpool.Pool, logger *zap.Logger) *VehicleRepository {
	return &VehicleRepository{
		db:     db,
		logger: logger,
	}
}

func (r *VehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	query := `
		INSERT INTO vehicles (
			id, org_id, license_plate, model, make, year, color, status, capacity, fuel_type, 
			vin, registration, insurance, odometer, last_service, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), $16, NOW(), $17)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		vehicle.ID,
		vehicle.OrgID,
		vehicle.LicensePlate,
		vehicle.Model,
		vehicle.Make,
		vehicle.Year,
		vehicle.Color,
		vehicle.Status,
		vehicle.Capacity,
		vehicle.FuelType,
		vehicle.VIN,
		vehicle.Registration,
		vehicle.Insurance,
		vehicle.Odometer,
		vehicle.LastService,
		vehicle.CreatedBy,
		vehicle.UpdatedBy,
	).Scan(&vehicle.ID, &vehicle.CreatedAt, &vehicle.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create vehicle", zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	query := `
		SELECT id, org_id, license_plate, model, make, year, color, status, capacity, fuel_type, 
		       vin, registration, insurance, odometer, last_service, created_at, created_by, updated_at, updated_by
		FROM vehicles 
		WHERE id = $1`

	vehicle := &models.Vehicle{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&vehicle.ID,
		&vehicle.OrgID,
		&vehicle.LicensePlate,
		&vehicle.Model,
		&vehicle.Make,
		&vehicle.Year,
		&vehicle.Color,
		&vehicle.Status,
		&vehicle.Capacity,
		&vehicle.FuelType,
		&vehicle.VIN,
		&vehicle.Registration,
		&vehicle.Insurance,
		&vehicle.Odometer,
		&vehicle.LastService,
		&vehicle.CreatedAt,
		&vehicle.CreatedBy,
		&vehicle.UpdatedAt,
		&vehicle.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("vehicle not found")
		}
		r.logger.Error("failed to get vehicle by id", zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	return vehicle, nil
}

func (r *VehicleRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Vehicle, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM vehicles WHERE org_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count vehicles", zap.Error(err))
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	// Get vehicles
	query := `
		SELECT id, org_id, license_plate, model, make, year, color, status, capacity, fuel_type, 
		       vin, registration, insurance, odometer, last_service, created_at, created_by, updated_at, updated_by
		FROM vehicles 
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		r.logger.Error("failed to query vehicles", zap.Error(err))
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var vehicles []*models.Vehicle
	for rows.Next() {
		vehicle := &models.Vehicle{}
		err := rows.Scan(
			&vehicle.ID,
			&vehicle.OrgID,
			&vehicle.LicensePlate,
			&vehicle.Model,
			&vehicle.Make,
			&vehicle.Year,
			&vehicle.Color,
			&vehicle.Status,
			&vehicle.Capacity,
			&vehicle.FuelType,
			&vehicle.VIN,
			&vehicle.Registration,
			&vehicle.Insurance,
			&vehicle.Odometer,
			&vehicle.LastService,
			&vehicle.CreatedAt,
			&vehicle.CreatedBy,
			&vehicle.UpdatedAt,
			&vehicle.UpdatedBy,
		)
		if err != nil {
			r.logger.Error("failed to scan vehicle", zap.Error(err))
			return nil, 0, fmt.Errorf("scan error: %w", err)
		}
		vehicles = append(vehicles, vehicle)
	}

	return vehicles, total, nil
}

func (r *VehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	query := `
		UPDATE vehicles SET
			license_plate = COALESCE($1, license_plate),
			model = COALESCE($2, model),
			make = COALESCE($3, make),
			year = COALESCE($4, year),
			color = COALESCE($5, color),
			status = COALESCE($6, status),
			capacity = COALESCE($7, capacity),
			fuel_type = COALESCE($8, fuel_type),
			vin = COALESCE($9, vin),
			registration = COALESCE($10, registration),
			insurance = COALESCE($11, insurance),
			odometer = COALESCE($12, odometer),
			last_service = COALESCE($13, last_service),
			updated_at = NOW(),
			updated_by = $14
		WHERE id = $15
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query,
		vehicle.LicensePlate,
		vehicle.Model,
		vehicle.Make,
		vehicle.Year,
		vehicle.Color,
		vehicle.Status,
		vehicle.Capacity,
		vehicle.FuelType,
		vehicle.VIN,
		vehicle.Registration,
		vehicle.Insurance,
		vehicle.Odometer,
		vehicle.LastService,
		vehicle.UpdatedBy,
		vehicle.ID,
	).Scan(&vehicle.UpdatedAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return fmt.Errorf("vehicle not found")
		}
		r.logger.Error("failed to update vehicle", zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (r *VehicleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM vehicles WHERE id = $1`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete vehicle", zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("vehicle not found")
	}
	
	return nil
}

func (r *VehicleRepository) GetStats(ctx context.Context) (*models.VehicleStats, error) {
	stats := &models.VehicleStats{
		TotalVehicles:  0,
		ActiveVehicles: 0,
		TotalDistance:  0,
		TotalFuel:      0,
	}

	// 1. Vehicle Counts
	countQuery := `SELECT status, COUNT(*) FROM vehicles GROUP BY status`
	rows, err := r.db.Query(ctx, countQuery)
	if err != nil {
		r.logger.Error("failed to query vehicle counts", zap.Error(err))
	} else {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if err := rows.Scan(&status, &count); err == nil {
				stats.TotalVehicles += count
				if status == "active" || status == "ACTIVE" {
					stats.ActiveVehicles += count
				}
			}
		}
	}

	// 2. Distance and Fuel from Trips
	metricsQuery := `SELECT SUM(actual_distance_km), SUM(fuel_consumed_liters) FROM trips WHERE is_deleted = FALSE`
	err = r.db.QueryRow(ctx, metricsQuery).Scan(&stats.TotalDistance, &stats.TotalFuel)
	if err != nil {
		r.logger.Error("failed to query trip metrics", zap.Error(err))
	}

	return stats, nil
}

func (r *VehicleRepository) AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) error {
	query := `UPDATE vehicles SET current_driver_id = $1, updated_at = NOW() WHERE id = $2`
	
	result, err := r.db.Exec(ctx, query, driverID, vehicleID)
	if err != nil {
		r.logger.Error("failed to assign driver", zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("vehicle not found")
	}
	
	return nil
}