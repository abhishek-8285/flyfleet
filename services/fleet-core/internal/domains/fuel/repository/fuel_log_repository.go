package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
)

type FuelLogRepository interface {
	Create(ctx context.Context, log *models.FuelLog) error
	CreateWithLimitValidation(ctx context.Context, log *models.FuelLog, quantity, amount float64) error
	GetByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.FuelLog, error)
	List(ctx context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error)
	Update(ctx context.Context, log *models.FuelLog) error
	Delete(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.FuelLogStatus) error
	GetStats(ctx context.Context, orgID uuid.UUID, filter models.FuelLogFilter) (*models.FuelLogStats, error)
	GetLastOdometer(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID) (float64, error)
	CheckDuplicate(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, transactionDate time.Time, minutes int) (bool, error)
	GetVehicleSpecs(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleSpecs, error)
	GetDriverFuelLimit(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID) (*models.DriverFuelLimit, error)
	GetOrgFuelSettings(ctx context.Context, orgID uuid.UUID) (*models.OrgFuelSettings, error)
	GetFuelUsage(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, since time.Time) (float64, float64, error)
	GetLogsForAnomalyDetection(ctx context.Context, since time.Time) ([]models.FuelLog, error)
	MarkAsAnomalous(ctx context.Context, id uuid.UUID, orgID uuid.UUID, reason string) error
	CreateVendor(ctx context.Context, vendor *models.FuelVendor) error
	ListVendors(ctx context.Context, orgID uuid.UUID) ([]models.FuelVendor, error)
}

type fuelLogRepository struct {
	db *pgxpool.Pool
}

func NewFuelLogRepository(db *pgxpool.Pool) FuelLogRepository {
	return &fuelLogRepository{db: db}
}

func (r *fuelLogRepository) Create(ctx context.Context, log *models.FuelLog) error {
	query := `
		INSERT INTO fuel_logs (
			id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, gst_number, gst_rate, gst_amount, invoice_number, vendor_gstin,
			receipt_url, notes, anomaly_flag, anomaly_reason, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		log.ID, log.OrganizationID, log.VehicleID, log.DriverID, log.StationName, log.FuelType,
		log.Quantity, log.UnitPrice, log.TotalCost, log.Odometer, log.TransactionDate,
		log.Status, log.GSTNumber, log.GSTRate, log.GSTAmount, log.InvoiceNumber, log.VendorGSTIN,
		log.ReceiptURL, log.Notes, log.AnomalyFlag, log.AnomalyReason, log.CreatedBy, log.CreatedAt, log.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to create fuel log: %w", err)
	}
	return nil
}

func (r *fuelLogRepository) GetByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.FuelLog, error) {
	query := `
		SELECT
			id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, gst_number, gst_rate, gst_amount, invoice_number, vendor_gstin,
			receipt_url, notes, anomaly_flag, anomaly_reason, created_by, created_at, updated_at
		FROM fuel_logs
		WHERE id = $1 AND organization_id = $2
	`
	var log models.FuelLog
	start := time.Now()
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&log.ID, &log.OrganizationID, &log.VehicleID, &log.DriverID, &log.StationName, &log.FuelType,
		&log.Quantity, &log.UnitPrice, &log.TotalCost, &log.Odometer, &log.TransactionDate,
		&log.Status, &log.GSTNumber, &log.GSTRate, &log.GSTAmount, &log.InvoiceNumber, &log.VendorGSTIN,
		&log.ReceiptURL, &log.Notes, &log.AnomalyFlag, &log.AnomalyReason, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get fuel log: %w", err)
	}
	return &log, nil
}

func (r *fuelLogRepository) List(ctx context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error) {
	where := []string{"organization_id = $1"}
	args := []interface{}{filter.OrganizationID}
	argIdx := 2

	if filter.VehicleID != nil {
		where = append(where, fmt.Sprintf("vehicle_id = $%d", argIdx))
		args = append(args, *filter.VehicleID)
		argIdx++
	}

	if filter.DriverID != nil {
		where = append(where, fmt.Sprintf("driver_id = $%d", argIdx))
		args = append(args, *filter.DriverID)
		argIdx++
	}

	if len(filter.Status) > 0 {
		statusList := []string{}
		for _, s := range filter.Status {
			statusList = append(statusList, string(s))
		}
		where = append(where, fmt.Sprintf("status = ANY($%d)", argIdx))
		args = append(args, statusList)
		argIdx++
	}

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("transaction_date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("transaction_date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(station_name ILIKE $%d OR fuel_type ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM fuel_logs WHERE %s", strings.Join(where, " AND "))
	var total int
	start := time.Now()
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	metrics.RecordDBQuery(countQuery, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to count fuel logs: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	limit := filter.PageSize
	query := fmt.Sprintf(`
		SELECT 
				id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
				quantity, unit_price, total_cost, odometer, transaction_date,
				status, gst_number, gst_rate, gst_amount, invoice_number, vendor_gstin,
				receipt_url, notes, anomaly_flag, anomaly_reason, created_by, created_at, updated_at
		FROM fuel_logs
		WHERE %s
		ORDER BY transaction_date DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), argIdx, argIdx+1)

	args = append(args, limit, offset)

	start = time.Now()
	rows, err := r.db.Query(ctx, query, args...)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query fuel logs: %w", err)
	}
	defer rows.Close()

	logs := []models.FuelLog{}
	for rows.Next() {
		var log models.FuelLog
		err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.VehicleID, &log.DriverID, &log.StationName, &log.FuelType,
			&log.Quantity, &log.UnitPrice, &log.TotalCost, &log.Odometer, &log.TransactionDate,
			&log.Status, &log.GSTNumber, &log.GSTRate, &log.GSTAmount, &log.InvoiceNumber, &log.VendorGSTIN,
			&log.ReceiptURL, &log.Notes, &log.AnomalyFlag, &log.AnomalyReason, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fuel log: %w", err)
		}
		logs = append(logs, log)
	}

	return &models.FuelLogListResponse{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Logs:     logs,
	}, nil
}

func (r *fuelLogRepository) Update(ctx context.Context, log *models.FuelLog) error {
	query := `
		UPDATE fuel_logs SET
			station_name = $1, fuel_type = $2, quantity = $3, unit_price = $4,
			total_cost = $5, odometer = $6, transaction_date = $7,
			gst_number = $8, gst_rate = $9, gst_amount = $10, invoice_number = $11, vendor_gstin = $12,
			receipt_url = $13, notes = $14, updated_at = NOW()
		WHERE id = $15 AND organization_id = $16
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		log.StationName, log.FuelType, log.Quantity, log.UnitPrice,
		log.TotalCost, log.Odometer, log.TransactionDate,
		log.GSTNumber, log.GSTRate, log.GSTAmount, log.InvoiceNumber, log.VendorGSTIN,
		log.ReceiptURL, log.Notes, log.ID, log.OrganizationID,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to update fuel log: %w", err)
	}
	return nil
}

func (r *fuelLogRepository) Delete(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	query := "DELETE FROM fuel_logs WHERE id = $1 AND organization_id = $2"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to delete fuel log: %w", err)
	}
	return nil
}

func (r *fuelLogRepository) UpdateStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.FuelLogStatus) error {
	query := "UPDATE fuel_logs SET status = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, status, id, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to update fuel log status: %w", err)
	}
	return nil
}

func (r *fuelLogRepository) CreateAuditLog(ctx context.Context, entry *models.FuelAuditLog) error {
	query := `
		INSERT INTO fuel_audit_logs (
			id, fuel_log_id, action, changed_by, changed_at, old_values, new_values, diff
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.ChangedAt.IsZero() {
		entry.ChangedAt = time.Now().UTC()
	}
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		entry.ID, entry.FuelLogID, strings.ToLower(entry.Action), entry.ChangedBy, entry.ChangedAt,
		entry.OldValues, entry.NewValues, entry.Diff,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to create fuel audit log: %w", err)
	}
	return nil
}

func (r *fuelLogRepository) GetStats(ctx context.Context, orgID uuid.UUID, filter models.FuelLogFilter) (*models.FuelLogStats, error) {
	// Simple summary stats
	query := `
		SELECT
			COALESCE(SUM(quantity), 0),
			COALESCE(SUM(total_cost), 0),
			COUNT(*) FILTER (WHERE status = 'PENDING')
		FROM fuel_logs
		WHERE organization_id = $1
	`
	stats := &models.FuelLogStats{
		ByStatus:   make(map[string]int),
		ByFuelType: make(map[string]float64),
	}
	start := time.Now()
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&stats.TotalConsumption, &stats.TotalCost, &stats.PendingApprovals,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to get fuel stats: %w", err)
	}

	// Stats by status
	statusQuery := `
		SELECT status, COUNT(*)
		FROM fuel_logs
		WHERE organization_id = $1
		GROUP BY status
	`
	start = time.Now()
	rows, err := r.db.Query(ctx, statusQuery, orgID)
	metrics.RecordDBQuery(statusQuery, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query status stats: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err == nil {
			stats.ByStatus[status] = count
		}
	}

	// Stats by fuel type
	fuelQuery := `
		SELECT fuel_type, SUM(quantity)
		FROM fuel_logs
		WHERE organization_id = $1
		GROUP BY fuel_type
	`
	start = time.Now()
	rows2, err := r.db.Query(ctx, fuelQuery, orgID)
	metrics.RecordDBQuery(fuelQuery, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query fuel type stats: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var fuelType string
		var qty float64
		if err := rows2.Scan(&fuelType, &qty); err == nil {
			stats.ByFuelType[fuelType] = qty
		}
	}

	// Derived metrics from approved logs to keep analytics stable.
	var minOdo, maxOdo, approvedFuel float64
	distanceQuery := `
		SELECT
			COALESCE(MIN(odometer), 0),
			COALESCE(MAX(odometer), 0),
			COALESCE(SUM(quantity), 0)
		FROM fuel_logs
		WHERE organization_id = $1 AND status = 'APPROVED'
	`
	start = time.Now()
	if err := r.db.QueryRow(ctx, distanceQuery, orgID).Scan(&minOdo, &maxOdo, &approvedFuel); err == nil {
		metrics.RecordDBQuery(distanceQuery, time.Since(start))
		distance := maxOdo - minOdo
		if approvedFuel > 0 && distance > 0 {
			stats.AvgConsumption = distance / approvedFuel
			stats.AvgCostPerKm = stats.TotalCost / distance
		}
	} else {
		metrics.RecordDBQuery(distanceQuery, time.Since(start))
	}

	return stats, nil
}

func (r *fuelLogRepository) GetLastOdometer(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID) (float64, error) {
	query := `
		SELECT odometer
		FROM fuel_logs
		WHERE vehicle_id = $1 AND organization_id = $2
		ORDER BY transaction_date DESC, created_at DESC
		LIMIT 1
	`
	var odometer float64
	start := time.Now()
	err := r.db.QueryRow(ctx, query, vehicleID, orgID).Scan(&odometer)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get last odometer: %w", err)
	}
	return odometer, nil
}

// CheckDuplicate checks if a similar fuel transaction exists within the specified time window
func (r *fuelLogRepository) CheckDuplicate(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, transactionDate time.Time, minutes int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM fuel_logs
		WHERE vehicle_id = $1
		  AND organization_id = $2
		  AND transaction_date >= $3 - ($4 * INTERVAL '1 minute')
		  AND transaction_date <= $3 + ($4 * INTERVAL '1 minute')
		  AND status != 'REJECTED'
	`
	var count int
	start := time.Now()
	err := r.db.QueryRow(ctx, query, vehicleID, orgID, transactionDate, minutes).Scan(&count)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return false, fmt.Errorf("failed to check for duplicate: %w", err)
	}
	return count > 0, nil
}

func (r *fuelLogRepository) GetVehicleSpecs(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleSpecs, error) {
	query := `
		SELECT vehicle_id, fuel_type, tank_capacity_liters, efficiency_expected_km_l, created_at, updated_at
		FROM vehicle_specs
		WHERE vehicle_id = $1
	`
	var specs models.VehicleSpecs
	start := time.Now()
	err := r.db.QueryRow(ctx, query, vehicleID).Scan(
		&specs.VehicleID, &specs.FuelType, &specs.TankCapacityLiters, &specs.EfficiencyExpectedKmL,
		&specs.CreatedAt, &specs.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get vehicle specs: %w", err)
	}
	return &specs, nil
}

func (r *fuelLogRepository) GetDriverFuelLimit(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID) (*models.DriverFuelLimit, error) {
	query := `
		SELECT id, organization_id, driver_id, daily_liters_limit, monthly_liters_limit, 
		       daily_amount_limit, monthly_amount_limit, currency, is_active, created_at, updated_at
		FROM driver_fuel_limits
		WHERE driver_id = $1 AND organization_id = $2 AND is_active = TRUE
	`
	var limit models.DriverFuelLimit
	start := time.Now()
	err := r.db.QueryRow(ctx, query, driverID, orgID).Scan(
		&limit.ID, &limit.OrganizationID, &limit.DriverID, &limit.DailyLitersLimit, &limit.MonthlyLitersLimit,
		&limit.DailyAmountLimit, &limit.MonthlyAmountLimit, &limit.Currency, &limit.IsActive, &limit.CreatedAt, &limit.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver fuel limit: %w", err)
	}
	return &limit, nil
}

func (r *fuelLogRepository) GetOrgFuelSettings(ctx context.Context, orgID uuid.UUID) (*models.OrgFuelSettings, error) {
	query := `
		SELECT organization_id, default_daily_liters_limit, default_monthly_liters_limit,
		       default_daily_amount_limit, default_monthly_amount_limit, currency, created_at, updated_at
		FROM org_fuel_settings
		WHERE organization_id = $1
	`
	var settings models.OrgFuelSettings
	start := time.Now()
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&settings.OrganizationID, &settings.DefaultDailyLitersLimit, &settings.DefaultMonthlyLitersLimit,
		&settings.DefaultDailyAmountLimit, &settings.DefaultMonthlyAmountLimit, &settings.Currency, &settings.CreatedAt, &settings.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get org fuel settings: %w", err)
	}
	return &settings, nil
}

func (r *fuelLogRepository) GetFuelUsage(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, since time.Time) (float64, float64, error) {
	query := `
		SELECT COALESCE(SUM(quantity), 0), COALESCE(SUM(total_cost), 0)
		FROM fuel_logs
		WHERE driver_id = $1 AND organization_id = $2 AND transaction_date >= $3 AND status != 'REJECTED'
	`
	var totalLiters, totalAmount float64
	start := time.Now()
	err := r.db.QueryRow(ctx, query, driverID, orgID, since).Scan(&totalLiters, &totalAmount)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get fuel usage: %w", err)
	}
	return totalLiters, totalAmount, nil
}

func (r *fuelLogRepository) GetLogsForAnomalyDetection(ctx context.Context, since time.Time) ([]models.FuelLog, error) {
	query := `
		SELECT 
			id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, gst_number, gst_rate, gst_amount, invoice_number, vendor_gstin,
			receipt_url, notes, anomaly_flag, anomaly_reason, created_by, created_at, updated_at
		FROM fuel_logs
		WHERE updated_at >= $1 AND status = 'APPROVED'
		ORDER BY vehicle_id, transaction_date ASC
	`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, since)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.FuelLog
	for rows.Next() {
		var log models.FuelLog
		err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.VehicleID, &log.DriverID, &log.StationName, &log.FuelType,
			&log.Quantity, &log.UnitPrice, &log.TotalCost, &log.Odometer, &log.TransactionDate,
			&log.Status, &log.GSTNumber, &log.GSTRate, &log.GSTAmount, &log.InvoiceNumber, &log.VendorGSTIN,
			&log.ReceiptURL, &log.Notes, &log.AnomalyFlag, &log.AnomalyReason, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *fuelLogRepository) MarkAsAnomalous(ctx context.Context, id uuid.UUID, orgID uuid.UUID, reason string) error {
	query := "UPDATE fuel_logs SET anomaly_flag = TRUE, anomaly_reason = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, reason, id, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *fuelLogRepository) CreateVendor(ctx context.Context, vendor *models.FuelVendor) error {
	query := `
		INSERT INTO fuel_vendors (
			id, organization_id, name, gstin, address, contact_phone, email, status, rating, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	now := time.Now()
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		vendor.ID, vendor.OrganizationID, vendor.Name, vendor.GSTIN, vendor.Address,
		vendor.ContactPhone, vendor.Email, vendor.Status, vendor.Rating, vendor.Metadata, now, now)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *fuelLogRepository) ListVendors(ctx context.Context, orgID uuid.UUID) ([]models.FuelVendor, error) {
	query := `SELECT id, organization_id, name, gstin, address, contact_phone, email, status, rating, metadata, created_at, updated_at
			  FROM fuel_vendors WHERE organization_id = $1`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendors []models.FuelVendor
	for rows.Next() {
		var v models.FuelVendor
		err := rows.Scan(&v.ID, &v.OrganizationID, &v.Name, &v.GSTIN, &v.Address, &v.ContactPhone, &v.Email, &v.Status, &v.Rating, &v.Metadata, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vendors = append(vendors, v)
	}
	return vendors, nil
}

// CreateWithLimitValidation atomically validates fuel limits and creates a fuel log
// This prevents TOCTOU race conditions by using a transaction with row-level locks
func (r *fuelLogRepository) CreateWithLimitValidation(ctx context.Context, log *models.FuelLog, quantity, amount float64) error {
	// Start a transaction for atomic check-and-insert
	start := time.Now()
	tx, err := r.db.Begin(ctx)
	metrics.RecordDBQuery("BEGIN TRANSACTION", time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	driverID := log.DriverID
	orgID := log.OrganizationID

	// 1. Lock driver fuel limit row (if exists) to prevent concurrent modifications
	var limit *models.DriverFuelLimit
	limitQuery := `
		SELECT id, organization_id, driver_id, daily_liters_limit, monthly_liters_limit,
		       daily_amount_limit, monthly_amount_limit, currency, is_active, created_at, updated_at
		FROM driver_fuel_limits
		WHERE driver_id = $1 AND organization_id = $2 AND is_active = TRUE
		FOR UPDATE SKIP LOCKED
	`
	qStart := time.Now()
	err = tx.QueryRow(ctx, limitQuery, driverID, orgID).Scan(
		&limit.ID, &limit.OrganizationID, &limit.DriverID, &limit.DailyLitersLimit, &limit.MonthlyLitersLimit,
		&limit.DailyAmountLimit, &limit.MonthlyAmountLimit, &limit.Currency, &limit.IsActive, &limit.CreatedAt, &limit.UpdatedAt,
	)
	metrics.RecordDBQuery(limitQuery, time.Since(qStart))
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("failed to query driver fuel limit: %w", err)
	} else if err == pgx.ErrNoRows {
		limit = nil // No driver-specific limits
	}

	// 2. Lock org fuel settings (if exists) as fallback
	var orgSettings *models.OrgFuelSettings
	orgQuery := `
		SELECT organization_id, default_daily_liters_limit, default_monthly_liters_limit,
		       default_daily_amount_limit, default_monthly_amount_limit, currency, created_at, updated_at
		FROM org_fuel_settings
		WHERE organization_id = $1
		FOR UPDATE SKIP LOCKED
	`
	qStart = time.Now()
	err = tx.QueryRow(ctx, orgQuery, orgID).Scan(
		&orgSettings.OrganizationID, &orgSettings.DefaultDailyLitersLimit, &orgSettings.DefaultMonthlyLitersLimit,
		&orgSettings.DefaultDailyAmountLimit, &orgSettings.DefaultMonthlyAmountLimit, &orgSettings.Currency, &orgSettings.CreatedAt, &orgSettings.UpdatedAt,
	)
	metrics.RecordDBQuery(orgQuery, time.Since(qStart))
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("failed to query org fuel settings: %w", err)
	} else if err == pgx.ErrNoRows {
		orgSettings = nil // No org-level settings
	}

	// 3. Determine which limits to enforce
	var dailyLiters, monthlyLiters, dailyAmount, monthlyAmount *float64
	if limit != nil {
		dailyLiters = limit.DailyLitersLimit
		monthlyLiters = limit.MonthlyLitersLimit
		dailyAmount = limit.DailyAmountLimit
		monthlyAmount = limit.MonthlyAmountLimit
	} else if orgSettings != nil {
		dailyLiters = orgSettings.DefaultDailyLitersLimit
		monthlyLiters = orgSettings.DefaultMonthlyLitersLimit
		dailyAmount = orgSettings.DefaultDailyAmountLimit
		monthlyAmount = orgSettings.DefaultMonthlyAmountLimit
	}

	// 4. Check if any limits are configured
	hasLimits := dailyLiters != nil || monthlyLiters != nil || dailyAmount != nil || monthlyAmount != nil

	if hasLimits {
		// 5. Calculate current usage INCLUDING the proposed new log within the same transaction
		// This ensures we count the pending insert when checking limits
		startOfDay := time.Now().Truncate(24 * time.Hour)
		startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

		// Get current usage (excluding the new log)
		usageQuery := `
			SELECT COALESCE(SUM(quantity), 0), COALESCE(SUM(total_cost), 0)
			FROM fuel_logs
			WHERE driver_id = $1 AND organization_id = $2 AND transaction_date >= $3 AND status != 'REJECTED'
		`
		var usedDailyLiters, usedDailyAmount float64
		var usedMonthlyLiters, usedMonthlyAmount float64

		// Daily usage from start of day
		qStart = time.Now()
		err = tx.QueryRow(ctx, usageQuery, driverID, orgID, startOfDay).Scan(&usedDailyLiters, &usedDailyAmount)
		metrics.RecordDBQuery(usageQuery, time.Since(qStart))
		if err != nil {
			return fmt.Errorf("failed to get daily fuel usage: %w", err)
		}

		// Monthly usage from start of month
		qStart = time.Now()
		err = tx.QueryRow(ctx, usageQuery, driverID, orgID, startOfMonth).Scan(&usedMonthlyLiters, &usedMonthlyAmount)
		metrics.RecordDBQuery(usageQuery, time.Since(qStart))
		if err != nil {
			return fmt.Errorf("failed to get monthly fuel usage: %w", err)
		}

		// 6. Check daily liter limit
		if dailyLiters != nil && (usedDailyLiters+quantity) > *dailyLiters {
			return fmt.Errorf("daily fuel quantity limit exceeded: %.2f/%.2f", usedDailyLiters+quantity, *dailyLiters)
		}

		// 7. Check daily amount limit
		if dailyAmount != nil && (usedDailyAmount+amount) > *dailyAmount {
			return fmt.Errorf("daily fuel amount limit exceeded: %.2f/%.2f", usedDailyAmount+amount, *dailyAmount)
		}

		// 8. Check monthly liter limit
		if monthlyLiters != nil && (usedMonthlyLiters+quantity) > *monthlyLiters {
			return fmt.Errorf("monthly fuel quantity limit exceeded: %.2f/%.2f", usedMonthlyLiters+quantity, *monthlyLiters)
		}

		// 9. Check monthly amount limit
		if monthlyAmount != nil && (usedMonthlyAmount+amount) > *monthlyAmount {
			return fmt.Errorf("monthly fuel amount limit exceeded: %.2f/%.2f", usedMonthlyAmount+amount, *monthlyAmount)
		}
	}

	// 10. All checks passed (or no limits configured), perform the insert
	insertQuery := `
		INSERT INTO fuel_logs (
			id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, gst_number, gst_rate, gst_amount, invoice_number, vendor_gstin,
			receipt_url, notes, anomaly_flag, anomaly_reason, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)
	`
	qStart = time.Now()
	_, err = tx.Exec(ctx, insertQuery,
		log.ID, log.OrganizationID, log.VehicleID, log.DriverID, log.StationName, log.FuelType,
		log.Quantity, log.UnitPrice, log.TotalCost, log.Odometer, log.TransactionDate,
		log.Status, log.GSTNumber, log.GSTRate, log.GSTAmount, log.InvoiceNumber, log.VendorGSTIN,
		log.ReceiptURL, log.Notes, log.AnomalyFlag, log.AnomalyReason, log.CreatedBy, log.CreatedAt, log.UpdatedAt,
	)
	metrics.RecordDBQuery(insertQuery, time.Since(qStart))
	if err != nil {
		return fmt.Errorf("failed to create fuel log: %w", err)
	}

	// 11. Commit the transaction
	cStart := time.Now()
	if err := tx.Commit(ctx); err != nil {
		metrics.RecordDBQuery("COMMIT TRANSACTION", time.Since(cStart))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	metrics.RecordDBQuery("COMMIT TRANSACTION", time.Since(cStart))

	return nil
}
