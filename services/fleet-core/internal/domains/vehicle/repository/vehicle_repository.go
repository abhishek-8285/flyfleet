package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type VehicleRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewVehicleRepository(db *pgxpool.Pool, logger *zap.Logger) *VehicleRepository {
	return &VehicleRepository{db: db, logger: logger}
}

func (r *VehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	query := `
		INSERT INTO vehicles (
			id, tenant_id, registration_number, chassis_number, engine_number, vin_number,
			vehicle_type, capacity_tonnage, fuel_type, fuel_tank_capacity, ownership_type,
			parent_vehicle_id, branch_id, cost_center_id, status,
			gps_device_id, immobilizer_enabled, ais140_compliant,
			make, model, year, color, odometer, current_driver_id, speed_limit,
			created_at, updated_at, version
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25,
			NOW(), NOW(), 1
		) RETURNING id, created_at, updated_at, version`

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		vehicle.ID, vehicle.TenantID, vehicle.RegistrationNumber,
		vehicle.ChassisNumber, vehicle.EngineNumber, vehicle.VINNumber,
		vehicle.VehicleType, vehicle.CapacityTonnage, vehicle.FuelType,
		vehicle.FuelTankCapacity, vehicle.OwnershipType,
		vehicle.ParentVehicleID, vehicle.BranchID, vehicle.CostCenterID, vehicle.Status,
		vehicle.GPSDeviceID, vehicle.ImmobilizerEnabled, vehicle.AIS140Compliant,
		vehicle.Make, vehicle.Model, vehicle.Year, vehicle.Color, vehicle.Odometer, vehicle.CurrentDriverID, vehicle.SpeedLimit,
	).Scan(&vehicle.ID, &vehicle.CreatedAt, &vehicle.UpdatedAt, &vehicle.Version)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	query := `
		SELECT id, tenant_id, registration_number, chassis_number, engine_number, vin_number,
		       vehicle_type, capacity_tonnage, fuel_type, fuel_tank_capacity, ownership_type,
		       parent_vehicle_id, branch_id, cost_center_id, status,
		       gps_device_id, immobilizer_enabled, ais140_compliant,
		       make, model, year, color, odometer, current_driver_id, speed_limit,
		       created_at, updated_at, deleted_at, version
		FROM vehicles WHERE id = $1 AND deleted_at IS NULL`

	v := &models.Vehicle{}
	start := time.Now()
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.TenantID, &v.RegistrationNumber, &v.ChassisNumber, &v.EngineNumber, &v.VINNumber,
		&v.VehicleType, &v.CapacityTonnage, &v.FuelType, &v.FuelTankCapacity, &v.OwnershipType,
		&v.ParentVehicleID, &v.BranchID, &v.CostCenterID, &v.Status,
		&v.GPSDeviceID, &v.ImmobilizerEnabled, &v.AIS140Compliant,
		&v.Make, &v.Model, &v.Year, &v.Color, &v.Odometer, &v.CurrentDriverID, &v.SpeedLimit,
		&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt, &v.Version,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("vehicle not found")
		}
		r.logger.Error("failed to get vehicle by id", zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}
	return v, nil
}

func (r *VehicleRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Vehicle, int64, error) {
	var total int64
	start := time.Now()
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM vehicles WHERE tenant_id = $1 AND deleted_at IS NULL`, orgID).Scan(&total); err != nil {
		metrics.RecordDBQuery("SELECT COUNT(*) FROM vehicles WHERE tenant_id = $1 AND deleted_at IS NULL", time.Since(start))
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	metrics.RecordDBQuery("SELECT COUNT(*) FROM vehicles WHERE tenant_id = $1 AND deleted_at IS NULL", time.Since(start))

	query := `
		SELECT id, tenant_id, registration_number, chassis_number, engine_number, vin_number,
		       vehicle_type, capacity_tonnage, fuel_type, fuel_tank_capacity, ownership_type,
		       parent_vehicle_id, branch_id, cost_center_id, status,
		       gps_device_id, immobilizer_enabled, ais140_compliant,
		       make, model, year, color, odometer, current_driver_id, speed_limit,
		       created_at, updated_at, deleted_at, version
		FROM vehicles WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	queryStart := time.Now()
	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	metrics.RecordDBQuery(query, time.Since(queryStart))
	if err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var vehicles []*models.Vehicle
	for rows.Next() {
		v := &models.Vehicle{}
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.RegistrationNumber, &v.ChassisNumber, &v.EngineNumber, &v.VINNumber,
			&v.VehicleType, &v.CapacityTonnage, &v.FuelType, &v.FuelTankCapacity, &v.OwnershipType,
			&v.ParentVehicleID, &v.BranchID, &v.CostCenterID, &v.Status,
			&v.GPSDeviceID, &v.ImmobilizerEnabled, &v.AIS140Compliant,
			&v.Make, &v.Model, &v.Year, &v.Color, &v.Odometer, &v.CurrentDriverID, &v.SpeedLimit,
			&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt, &v.Version,
		); err != nil {
			return nil, 0, fmt.Errorf("scan error: %w", err)
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, total, nil
}

// Update performs an optimistic-locked update (increments version).
func (r *VehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	query := `
		UPDATE vehicles SET
			registration_number  = COALESCE($1, registration_number),
			chassis_number       = COALESCE($2, chassis_number),
			engine_number        = COALESCE($3, engine_number),
			vin_number           = COALESCE($4, vin_number),
			vehicle_type         = COALESCE($5, vehicle_type),
			capacity_tonnage     = COALESCE($6, capacity_tonnage),
			fuel_type            = COALESCE($7, fuel_type),
			fuel_tank_capacity   = COALESCE($8, fuel_tank_capacity),
			ownership_type       = COALESCE($9, ownership_type),
			parent_vehicle_id    = COALESCE($10, parent_vehicle_id),
			branch_id            = COALESCE($11, branch_id),
			cost_center_id       = COALESCE($12, cost_center_id),
			status               = COALESCE($13, status),
			gps_device_id        = COALESCE($14, gps_device_id),
			immobilizer_enabled  = $15,
			ais140_compliant     = $16,
			make                 = COALESCE($17, make),
			model                = COALESCE($18, model),
			year                 = COALESCE($19, year),
			color                = COALESCE($20, color),
			odometer             = COALESCE($21, odometer),
			current_driver_id    = COALESCE($22, current_driver_id),
			speed_limit          = $23,
			updated_at           = NOW(),
			version              = version + 1
		WHERE id = $24 AND deleted_at IS NULL AND version = $25
		RETURNING updated_at, version`

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		vehicle.RegistrationNumber, vehicle.ChassisNumber, vehicle.EngineNumber, vehicle.VINNumber,
		vehicle.VehicleType, vehicle.CapacityTonnage, vehicle.FuelType, vehicle.FuelTankCapacity,
		vehicle.OwnershipType, vehicle.ParentVehicleID, vehicle.BranchID, vehicle.CostCenterID,
		vehicle.Status, vehicle.GPSDeviceID, vehicle.ImmobilizerEnabled, vehicle.AIS140Compliant,
		vehicle.Make, vehicle.Model, vehicle.Year, vehicle.Color, vehicle.Odometer, vehicle.CurrentDriverID, vehicle.SpeedLimit,
		vehicle.ID, vehicle.Version,
	).Scan(&vehicle.UpdatedAt, &vehicle.Version)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		if err.Error() == "no rows in result set" {
			return fmt.Errorf("vehicle not found or version conflict")
		}
		r.logger.Error("failed to update vehicle", zap.Error(err))
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

// Delete performs a soft delete.
func (r *VehicleRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	query := `UPDATE vehicles SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
	start := time.Now()
	result, err := r.db.Exec(ctx, query, id, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("vehicle not found")
	}
	return nil
}

func (r *VehicleRepository) AssignDriver(ctx context.Context, vehicleID, orgID, driverID uuid.UUID) error {
	query := `UPDATE vehicles SET updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
	start := time.Now()
	result, err := r.db.Exec(ctx, query, vehicleID, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("vehicle not found")
	}
	return nil
}

func (r *VehicleRepository) GetStats(ctx context.Context) (*models.VehicleStats, error) {
	stats := &models.VehicleStats{}

	start := time.Now()
	rows, err := r.db.Query(ctx, `SELECT status, COUNT(*) FROM vehicles WHERE deleted_at IS NULL GROUP BY status`)
	metrics.RecordDBQuery(`SELECT status, COUNT(*) FROM vehicles WHERE deleted_at IS NULL GROUP BY status`, time.Since(start))
	if err != nil {
		r.logger.Error("failed to query vehicle counts", zap.Error(err))
	} else {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if rows.Scan(&status, &count) == nil {
				stats.TotalVehicles += count
				switch status {
				case models.StatusActive:
					stats.ActiveVehicles += count
				case models.StatusMaintenance:
					stats.MaintenanceCount += count
				case models.StatusSuspended:
					stats.SuspendedCount += count
				}
			}
		}
	}

	// Distance and fuel from trips
	start = time.Now()
	_ = r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(actual_distance_km),0), COALESCE(SUM(fuel_consumed_liters),0)
		 FROM trips WHERE is_deleted = FALSE`).
		Scan(&stats.TotalDistance, &stats.TotalFuel)
	metrics.RecordDBQuery(`SELECT COALESCE(SUM(actual_distance_km),0), COALESCE(SUM(fuel_consumed_liters),0) FROM trips WHERE is_deleted = FALSE`, time.Since(start))

	return stats, nil
}

func (r *VehicleRepository) GetByRegistrationNumber(ctx context.Context, tenantID uuid.UUID, regNo string) (*models.Vehicle, error) {
	v := &models.Vehicle{}
	start := time.Now()
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, registration_number, status, version, created_at, updated_at
		 FROM vehicles WHERE tenant_id = $1 AND registration_number = $2 AND deleted_at IS NULL`,
		tenantID, regNo,
	).Scan(&v.ID, &v.TenantID, &v.RegistrationNumber, &v.Status, &v.Version, &v.CreatedAt, &v.UpdatedAt)
	metrics.RecordDBQuery(`SELECT id, tenant_id, registration_number, status, version, created_at, updated_at FROM vehicles WHERE tenant_id = $1 AND registration_number = $2 AND deleted_at IS NULL`, time.Since(start))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // not found (nil, nil) = caller checks
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return v, nil
}

func (r *VehicleRepository) Search(ctx context.Context, params models.VehicleSearchParams) ([]*models.Vehicle, int64, error) {
	args := []interface{}{}
	where := "deleted_at IS NULL"
	argN := 1

	if params.TenantID != nil {
		where += fmt.Sprintf(" AND tenant_id = $%d", argN)
		args = append(args, *params.TenantID)
		argN++
	}
	if params.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, params.Status)
		argN++
	}
	if params.VehicleType != "" {
		where += fmt.Sprintf(" AND vehicle_type = $%d", argN)
		args = append(args, params.VehicleType)
		argN++
	}
	if params.FuelType != "" {
		where += fmt.Sprintf(" AND fuel_type = $%d", argN)
		args = append(args, params.FuelType)
		argN++
	}
	if params.BranchID != nil {
		where += fmt.Sprintf(" AND branch_id = $%d", argN)
		args = append(args, *params.BranchID)
		argN++
	}
	if params.Query != "" {
		where += fmt.Sprintf(" AND (registration_number ILIKE $%d OR vin_number ILIKE $%d)", argN, argN)
		args = append(args, "%"+params.Query+"%")
		argN++
	}

	var total int64
	start := time.Now()
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles WHERE "+where, args...).Scan(&total); err != nil {
		metrics.RecordDBQuery("SELECT COUNT(*) FROM vehicles WHERE "+where, time.Since(start))
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	metrics.RecordDBQuery("SELECT COUNT(*) FROM vehicles WHERE "+where, time.Since(start))

	offset := (params.Page - 1) * params.Limit
	args = append(args, params.Limit, offset)
	query := fmt.Sprintf(`
		SELECT id, tenant_id, registration_number, chassis_number, engine_number, vin_number,
		       vehicle_type, capacity_tonnage, fuel_type, fuel_tank_capacity, ownership_type,
		       parent_vehicle_id, branch_id, cost_center_id, status,
		       gps_device_id, immobilizer_enabled, ais140_compliant,
		       make, model, year, color, odometer, current_driver_id, speed_limit, created_at, updated_at, deleted_at, version
		FROM vehicles WHERE %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argN, argN+1)

	queryStart := time.Now()
	rows, err := r.db.Query(ctx, query, args...)
	metrics.RecordDBQuery(query, time.Since(queryStart))
	if err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var vehicles []*models.Vehicle
	for rows.Next() {
		v := &models.Vehicle{}
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.RegistrationNumber, &v.ChassisNumber, &v.EngineNumber, &v.VINNumber,
			&v.VehicleType, &v.CapacityTonnage, &v.FuelType, &v.FuelTankCapacity, &v.OwnershipType,
			&v.ParentVehicleID, &v.BranchID, &v.CostCenterID, &v.Status,
			&v.GPSDeviceID, &v.ImmobilizerEnabled, &v.AIS140Compliant,
			&v.Make, &v.Model, &v.Year, &v.Color, &v.Odometer, &v.CurrentDriverID, &v.SpeedLimit,
			&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt, &v.Version,
		); err != nil {
			return nil, 0, fmt.Errorf("scan error: %w", err)
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, total, nil
}
