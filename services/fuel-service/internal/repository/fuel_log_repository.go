package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fuel-service/internal/models"
)

type FuelLogRepository interface {
	Create(ctx context.Context, log *models.FuelLog) error
	GetByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.FuelLog, error)
	List(ctx context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error)
	Update(ctx context.Context, log *models.FuelLog) error
	Delete(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.FuelLogStatus) error
	GetStats(ctx context.Context, orgID uuid.UUID, filter models.FuelLogFilter) (*models.FuelLogStats, error)
	GetLastOdometer(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID) (float64, error)
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
			status, receipt_url, notes, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.OrganizationID, log.VehicleID, log.DriverID, log.StationName, log.FuelType,
		log.Quantity, log.UnitPrice, log.TotalCost, log.Odometer, log.TransactionDate,
		log.Status, log.ReceiptURL, log.Notes, log.CreatedBy, log.CreatedAt, log.UpdatedAt,
	)
	return err
}

func (r *fuelLogRepository) GetByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.FuelLog, error) {
	query := `
		SELECT 
			id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, receipt_url, notes, created_by, created_at, updated_at
		FROM fuel_logs
		WHERE id = $1 AND organization_id = $2
	`
	var log models.FuelLog
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&log.ID, &log.OrganizationID, &log.VehicleID, &log.DriverID, &log.StationName, &log.FuelType,
		&log.Quantity, &log.UnitPrice, &log.TotalCost, &log.Odometer, &log.TransactionDate,
		&log.Status, &log.ReceiptURL, &log.Notes, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
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
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	limit := filter.PageSize
	query := fmt.Sprintf(`
		SELECT 
			id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
			quantity, unit_price, total_cost, odometer, transaction_date,
			status, receipt_url, notes, created_by, created_at, updated_at
		FROM fuel_logs
		WHERE %s
		ORDER BY transaction_date DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []models.FuelLog{}
	for rows.Next() {
		var log models.FuelLog
		err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.VehicleID, &log.DriverID, &log.StationName, &log.FuelType,
			&log.Quantity, &log.UnitPrice, &log.TotalCost, &log.Odometer, &log.TransactionDate,
			&log.Status, &log.ReceiptURL, &log.Notes, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
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
			receipt_url = $8, notes = $9, updated_at = NOW()
		WHERE id = $10 AND organization_id = $11
	`
	_, err := r.db.Exec(ctx, query,
		log.StationName, log.FuelType, log.Quantity, log.UnitPrice,
		log.TotalCost, log.Odometer, log.TransactionDate,
		log.ReceiptURL, log.Notes, log.ID, log.OrganizationID,
	)
	return err
}

func (r *fuelLogRepository) Delete(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	query := "DELETE FROM fuel_logs WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, orgID)
	return err
}

func (r *fuelLogRepository) UpdateStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.FuelLogStatus) error {
	query := "UPDATE fuel_logs SET status = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3"
	_, err := r.db.Exec(ctx, query, status, id, orgID)
	return err
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
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&stats.TotalConsumption, &stats.TotalCost, &stats.PendingApprovals,
	)
	if err != nil {
		return nil, err
	}

	// Stats by status
	statusQuery := `
		SELECT status, COUNT(*)
		FROM fuel_logs
		WHERE organization_id = $1
		GROUP BY status
	`
	rows, err := r.db.Query(ctx, statusQuery, orgID)
	if err != nil {
		return nil, err
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
	rows2, err := r.db.Query(ctx, fuelQuery, orgID)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var fuelType string
		var qty float64
		if err := rows2.Scan(&fuelType, &qty); err == nil {
			stats.ByFuelType[fuelType] = qty
		}
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
	err := r.db.QueryRow(ctx, query, vehicleID, orgID).Scan(&odometer)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return odometer, nil
}
