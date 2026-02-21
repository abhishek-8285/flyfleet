package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core-service-new/internal/models"
	"github.com/google/uuid"
)

type FleetRepository struct {
	db *sql.DB
}

func NewFleetRepository(db *sql.DB) *FleetRepository {
	return &FleetRepository{db: db}
}

func (r *FleetRepository) CreateVehicle(ctx context.Context, v *models.Vehicle) error {
	query := `
		INSERT INTO vehicles (id, org_id, license_plate, model, make, year, status, vin, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.OrgID, v.LicensePlate, v.Model, v.Make, v.Year, v.Status, v.VIN, v.CreatedAt, v.CreatedBy, v.UpdatedAt, v.UpdatedBy,
	)
	return err
}

func (r *FleetRepository) GetVehicle(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	query := `
		SELECT id, org_id, license_plate, model, make, year, status, vin, created_at, created_by, updated_at, updated_by
		FROM vehicles
		WHERE id = $1 AND is_deleted = false
	`
	var v models.Vehicle
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.OrgID, &v.LicensePlate, &v.Model, &v.Make, &v.Year, &v.Status, &v.VIN, &v.CreatedAt, &v.CreatedBy, &v.UpdatedAt, &v.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vehicle not found")
		}
		return nil, err
	}
	return &v, nil
}

func (r *FleetRepository) ListVehicles(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]models.Vehicle, int64, error) {
	// First get total
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vehicles WHERE org_id = $1 AND is_deleted = false", orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, org_id, license_plate, model, make, year, status, vin, created_at, created_by, updated_at, updated_by
		FROM vehicles
		WHERE org_id = $1 AND is_deleted = false
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(&v.ID, &v.OrgID, &v.LicensePlate, &v.Model, &v.Make, &v.Year, &v.Status, &v.VIN, &v.CreatedAt, &v.CreatedBy, &v.UpdatedAt, &v.UpdatedBy); err != nil {
			return nil, 0, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, total, nil
}

func (r *FleetRepository) UpdateVehicle(ctx context.Context, v *models.Vehicle) error {
	query := `
		UPDATE vehicles
		SET model = $1, make = $2, year = $3, status = $4, updated_at = $5, updated_by = $6
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query, v.Model, v.Make, v.Year, v.Status, v.UpdatedAt, v.UpdatedBy, v.ID)
	return err
}

func (r *FleetRepository) DeleteVehicle(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE vehicles SET is_deleted = true, updated_at = NOW() WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// AssignDriver updates both the vehicle and the driver records for an assignment
func (r *FleetRepository) AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update vehicle
	_, err = tx.ExecContext(ctx, "UPDATE vehicles SET current_driver_id = $1, updated_at = NOW() WHERE id = $2", driverID, vehicleID)
	if err != nil {
		return err
	}

	// 2. Update driver
	_, err = tx.ExecContext(ctx, "UPDATE drivers SET current_vehicle_id = $1, updated_at = NOW() WHERE id = $2", vehicleID, driverID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *FleetRepository) CreateDriver(ctx context.Context, d *models.Driver) error {
	query := `
		INSERT INTO drivers (
			id, org_id, first_name, last_name, email, phone,
			license_number, license_type, license_expiry, status,
			rating, behavior_score, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.OrgID, d.FirstName, d.LastName, d.Email, d.Phone,
		d.LicenseNumber, d.LicenseType, d.LicenseExpiry, d.Status,
		d.Rating, d.BehaviorScore, d.CreatedAt, d.CreatedBy, d.UpdatedAt, d.UpdatedBy,
	)
	return err
}

func (r *FleetRepository) GetDriver(ctx context.Context, id uuid.UUID) (*models.Driver, error) {
	query := `
		SELECT 
			id, org_id, first_name, last_name, email, phone,
			license_number, license_type, license_expiry, status,
			current_vehicle_id, rating, behavior_score, created_at, created_by, updated_at, updated_by
		FROM drivers
		WHERE id = $1 AND deleted_at IS NULL
	`
	var d models.Driver
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.OrgID, &d.FirstName, &d.LastName, &d.Email, &d.Phone,
		&d.LicenseNumber, &d.LicenseType, &d.LicenseExpiry, &d.Status,
		&d.CurrentVehicleID, &d.Rating, &d.BehaviorScore, &d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("driver not found")
		}
		return nil, err
	}
	return &d, nil
}

func (r *FleetRepository) ListDrivers(ctx context.Context, orgID uuid.UUID, limit, offset int32, status string) ([]models.Driver, int64, error) {
	var total int64
	countQuery := "SELECT COUNT(*) FROM drivers WHERE org_id = $1 AND deleted_at IS NULL"
	if status != "" {
		countQuery += " AND status = $2"
		err := r.db.QueryRowContext(ctx, countQuery, orgID, status).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery, orgID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	query := `
		SELECT 
			id, org_id, first_name, last_name, email, phone,
			license_number, license_type, license_expiry, status,
			current_vehicle_id, rating, behavior_score, created_at, created_by, updated_at, updated_by
		FROM drivers
		WHERE org_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{orgID}
	nextArg := 2
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", nextArg)
		args = append(args, status)
		nextArg++
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", nextArg, nextArg+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var drivers []models.Driver
	for rows.Next() {
		var d models.Driver
		err := rows.Scan(
			&d.ID, &d.OrgID, &d.FirstName, &d.LastName, &d.Email, &d.Phone,
			&d.LicenseNumber, &d.LicenseType, &d.LicenseExpiry, &d.Status,
			&d.CurrentVehicleID, &d.Rating, &d.BehaviorScore, &d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		drivers = append(drivers, d)
	}
	return drivers, total, nil
}

func (r *FleetRepository) UpdateDriver(ctx context.Context, d *models.Driver) error {
	query := `
		UPDATE drivers
		SET first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5, updated_by = $6
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query, d.FirstName, d.LastName, d.Email, d.Phone, d.UpdatedAt, d.UpdatedBy, d.ID)
	return err
}

func (r *FleetRepository) DeleteDriver(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE drivers SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *FleetRepository) UpdateDriverStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := "UPDATE drivers SET status = $1, updated_at = NOW() WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// Maintenance methods

func (r *FleetRepository) CreateWorkOrder(ctx context.Context, wo *models.WorkOrder) error {
	query := `
		INSERT INTO work_orders (
			id, org_id, vehicle_id, title, description,
			maintenance_type, priority, status, assigned_to,
			scheduled_date, estimated_cost, actual_cost,
			vendor_id, approval_status, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err := r.db.ExecContext(ctx, query,
		wo.ID, wo.OrgID, wo.VehicleID, wo.Title, wo.Description,
		wo.Type, wo.Priority, wo.Status, wo.AssignedTo,
		wo.ScheduledDate, wo.EstimatedCost, wo.ActualCost,
		wo.VendorID, wo.ApprovalStatus, wo.CreatedAt, wo.CreatedBy, wo.UpdatedAt, wo.UpdatedBy,
	)
	return err
}

func (r *FleetRepository) GetWorkOrder(ctx context.Context, id uuid.UUID) (*models.WorkOrder, error) {
	query := `
		SELECT 
			id, org_id, vehicle_id, title, description,
			maintenance_type, priority, status, assigned_to,
			scheduled_date, completed_date, estimated_cost, actual_cost,
			vendor_id, approval_status, created_at, created_by, updated_at, updated_by
		FROM work_orders
		WHERE id = $1
	`
	var wo models.WorkOrder
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wo.ID, &wo.OrgID, &wo.VehicleID, &wo.Title, &wo.Description,
		&wo.Type, &wo.Priority, &wo.Status, &wo.AssignedTo,
		&wo.ScheduledDate, &wo.CompletedDate, &wo.EstimatedCost, &wo.ActualCost,
		&wo.VendorID, &wo.ApprovalStatus, &wo.CreatedAt, &wo.CreatedBy, &wo.UpdatedAt, &wo.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("work order not found")
		}
		return nil, err
	}
	return &wo, nil
}

func (r *FleetRepository) ListWorkOrders(ctx context.Context, orgID uuid.UUID, limit, offset int32, vehicleID *uuid.UUID, status string) ([]models.WorkOrder, int64, error) {
	var total int64
	countQuery := "SELECT COUNT(*) FROM work_orders WHERE org_id = $1"
	args := []interface{}{orgID}
	nextArg := 2

	if vehicleID != nil {
		countQuery += fmt.Sprintf(" AND vehicle_id = $%d", nextArg)
		args = append(args, *vehicleID)
		nextArg++
	}
	if status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", nextArg)
		args = append(args, status)
		nextArg++
	}

	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, org_id, vehicle_id, title, description,
			maintenance_type, priority, status, assigned_to,
			scheduled_date, completed_date, estimated_cost, actual_cost,
			vendor_id, approval_status, created_at, created_by, updated_at, updated_by
		FROM work_orders
		WHERE org_id = $1
	`
	// Reuse args and nextArg for filtering
	queryArgs := []interface{}{orgID}
	queryNextArg := 2
	if vehicleID != nil {
		query += fmt.Sprintf(" AND vehicle_id = $%d", queryNextArg)
		queryArgs = append(queryArgs, *vehicleID)
		queryNextArg++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", queryNextArg)
		queryArgs = append(queryArgs, status)
		queryNextArg++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", queryNextArg, queryNextArg+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var wos []models.WorkOrder
	for rows.Next() {
		var wo models.WorkOrder
		err := rows.Scan(
			&wo.ID, &wo.OrgID, &wo.VehicleID, &wo.Title, &wo.Description,
			&wo.Type, &wo.Priority, &wo.Status, &wo.AssignedTo,
			&wo.ScheduledDate, &wo.CompletedDate, &wo.EstimatedCost, &wo.ActualCost,
			&wo.VendorID, &wo.ApprovalStatus, &wo.CreatedAt, &wo.CreatedBy, &wo.UpdatedAt, &wo.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		wos = append(wos, wo)
	}
	return wos, total, nil
}

func (r *FleetRepository) UpdateWorkOrder(ctx context.Context, wo *models.WorkOrder) error {
	query := `
		UPDATE work_orders
		SET title = $1, description = $2, priority = $3, assigned_to = $4,
		    estimated_cost = $5, actual_cost = $6, updated_at = $7, updated_by = $8
		WHERE id = $9
	`
	_, err := r.db.ExecContext(ctx, query,
		wo.Title, wo.Description, wo.Priority, wo.AssignedTo,
		wo.EstimatedCost, wo.ActualCost, wo.UpdatedAt, wo.UpdatedBy, wo.ID,
	)
	return err
}

func (r *FleetRepository) ChangeWorkOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := "UPDATE work_orders SET status = $1, updated_at = NOW() WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *FleetRepository) CreateInspection(ctx context.Context, i *models.Inspection) error {
	query := `
		INSERT INTO vehicle_inspections (
			id, org_id, vehicle_id, inspector_id,
			status, notes, odometer_at_inspection,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		i.ID, i.OrgID, i.VehicleID, i.InspectorID,
		i.Status, i.Notes, i.Odometer,
		i.CreatedAt, i.CreatedBy, i.UpdatedAt, i.UpdatedBy,
	)
	return err
}

func (r *FleetRepository) ListInspections(ctx context.Context, orgID uuid.UUID, limit, offset int32, vehicleID *uuid.UUID) ([]models.Inspection, int64, error) {
	var total int64
	countQuery := "SELECT COUNT(*) FROM vehicle_inspections WHERE org_id = $1"
	args := []interface{}{orgID}
	nextArg := 2

	if vehicleID != nil {
		countQuery += fmt.Sprintf(" AND vehicle_id = $%d", nextArg)
		args = append(args, *vehicleID)
		nextArg++
	}

	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, org_id, vehicle_id, inspector_id,
			status, notes, odometer_at_inspection,
			created_at, created_by, updated_at, updated_by
		FROM vehicle_inspections
		WHERE org_id = $1
	`
	queryArgs := []interface{}{orgID}
	queryNextArg := 2
	if vehicleID != nil {
		query += fmt.Sprintf(" AND vehicle_id = $%d", queryNextArg)
		queryArgs = append(queryArgs, *vehicleID)
		queryNextArg++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", queryNextArg, queryNextArg+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var inspections []models.Inspection
	for rows.Next() {
		var i models.Inspection
		err := rows.Scan(
			&i.ID, &i.OrgID, &i.VehicleID, &i.InspectorID,
			&i.Status, &i.Notes, &i.Odometer,
			&i.CreatedAt, &i.CreatedBy, &i.UpdatedAt, &i.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		inspections = append(inspections, i)
	}
	return inspections, total, nil
}

// Fuel & Telemetry methods

func (r *FleetRepository) CreateFuelLog(ctx context.Context, fl *models.FuelLog) error {
	query := `
		INSERT INTO fuel_logs (
			id, org_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, receipt_url, notes, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err := r.db.ExecContext(ctx, query,
		fl.ID, fl.OrgID, fl.VehicleID, fl.DriverID, fl.StationName, fl.FuelType,
		fl.Quantity, fl.UnitPrice, fl.TotalCost, fl.Odometer, fl.TransactionDate,
		fl.Status, fl.ReceiptURL, fl.Notes, fl.CreatedAt, fl.CreatedBy, fl.UpdatedAt, fl.UpdatedBy,
	)
	return err
}

func (r *FleetRepository) ListFuelLogs(ctx context.Context, orgID uuid.UUID, limit, offset int32, vehicleID, driverID *uuid.UUID) ([]models.FuelLog, int64, error) {
	where := "WHERE org_id = $1"
	args := []interface{}{orgID}
	nextArg := 2

	if vehicleID != nil {
		where += fmt.Sprintf(" AND vehicle_id = $%d", nextArg)
		args = append(args, *vehicleID)
		nextArg++
	}
	if driverID != nil {
		where += fmt.Sprintf(" AND driver_id = $%d", nextArg)
		args = append(args, *driverID)
		nextArg++
	}

	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM fuel_logs "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, receipt_url, notes, created_at, created_by, updated_at, updated_by
		FROM fuel_logs
		%s
		ORDER BY transaction_date DESC
		LIMIT $%d OFFSET $%d
	`, where, nextArg, nextArg+1)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.FuelLog
	for rows.Next() {
		var fl models.FuelLog
		err := rows.Scan(
			&fl.ID, &fl.OrgID, &fl.VehicleID, &fl.DriverID, &fl.StationName, &fl.FuelType,
			&fl.Quantity, &fl.UnitPrice, &fl.TotalCost, &fl.Odometer, &fl.TransactionDate,
			&fl.Status, &fl.ReceiptURL, &fl.Notes, &fl.CreatedAt, &fl.CreatedBy, &fl.UpdatedAt, &fl.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, fl)
	}
	return logs, total, nil
}

func (r *FleetRepository) GetFuelStats(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID) (float64, float64, float64, float64, error) {
	where := "WHERE org_id = $1"
	args := []interface{}{orgID}
	if vehicleID != nil {
		where += " AND vehicle_id = $2"
		args = append(args, *vehicleID)
	}

	query := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(quantity), 0),
			COALESCE(SUM(total_cost), 0),
			COALESCE(AVG(quantity), 0),
			COALESCE(SUM(total_cost) / NULLIF(SUM(odometer), 0), 0)
		FROM fuel_logs
		%s
	`, where)

	var totalQty, totalCost, avgQty, costPerKm float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&totalQty, &totalCost, &avgQty, &costPerKm)
	return totalQty, totalCost, avgQty, costPerKm, err
}

func (r *FleetRepository) ReportTelemetry(ctx context.Context, t *models.Telemetry) error {
	query := `
		INSERT INTO vehicle_telemetry (
			time, vehicle_id, org_id, device_id, latitude, longitude,
			speed, fuel_level, battery_voltage, coolant_temp, odometer, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	metadataJSON, _ := json.Marshal(t.Metadata)
	_, err := r.db.ExecContext(ctx, query,
		t.Timestamp, t.VehicleID, t.OrgID, t.DeviceID, t.Latitude, t.Longitude,
		t.Speed, t.FuelLevel, t.BatteryVoltage, t.CoolantTemp, t.Odometer, metadataJSON,
	)
	return err
}

func (r *FleetRepository) GetLatestTelemetry(ctx context.Context, vehicleID uuid.UUID) (*models.Telemetry, error) {
	query := `
		SELECT 
			time, vehicle_id, org_id, device_id, latitude, longitude,
			speed, fuel_level, battery_voltage, coolant_temp, odometer, metadata
		FROM vehicle_telemetry
		WHERE vehicle_id = $1
		ORDER BY time DESC
		LIMIT 1
	`
	var t models.Telemetry
	var metadataJSON []byte
	err := r.db.QueryRowContext(ctx, query, vehicleID).Scan(
		&t.Timestamp, &t.VehicleID, &t.OrgID, &t.DeviceID, &t.Latitude, &t.Longitude,
		&t.Speed, &t.FuelLevel, &t.BatteryVoltage, &t.CoolantTemp, &t.Odometer, &metadataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	_ = json.Unmarshal(metadataJSON, &t.Metadata)
	return &t, nil
}
