package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FinancialRepository struct {
	db *pgxpool.Pool
}

func NewFinancialRepository(db *pgxpool.Pool) *FinancialRepository {
	return &FinancialRepository{db: db}
}

func (r *FinancialRepository) Upsert(ctx context.Context, f *models.VehicleFinancials) error {
	query := `
		INSERT INTO vehicle_financials (
			vehicle_id, acquisition_cost, depreciation_method, depreciation_rate,
			residual_value, emi_amount, lease_end_date, insurance_premium,
			total_maintenance_cost, total_fuel_cost, total_toll_cost,
			tco_cached, tco_calculated_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		ON CONFLICT (vehicle_id) DO UPDATE SET
			acquisition_cost       = EXCLUDED.acquisition_cost,
			depreciation_method    = EXCLUDED.depreciation_method,
			depreciation_rate      = EXCLUDED.depreciation_rate,
			residual_value         = EXCLUDED.residual_value,
			emi_amount             = EXCLUDED.emi_amount,
			lease_end_date         = EXCLUDED.lease_end_date,
			insurance_premium      = EXCLUDED.insurance_premium,
			total_maintenance_cost = EXCLUDED.total_maintenance_cost,
			total_fuel_cost        = EXCLUDED.total_fuel_cost,
			total_toll_cost        = EXCLUDED.total_toll_cost,
			tco_cached             = EXCLUDED.tco_cached,
			tco_calculated_at      = EXCLUDED.tco_calculated_at,
			updated_at             = NOW()
		RETURNING created_at, updated_at`

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		f.VehicleID, f.AcquisitionCost, f.DepreciationMethod, f.DepreciationRate,
		f.ResidualValue, f.EMIAmount, f.LeaseEndDate, f.InsurancePremium,
		f.TotalMaintenanceCost, f.TotalFuelCost, f.TotalTollCost,
		f.TCOCached, f.TCOCalculatedAt,
	).Scan(&f.CreatedAt, &f.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to upsert vehicle financials: %w", err)
	}
	return nil
}

func (r *FinancialRepository) Get(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleFinancials, error) {
	f := &models.VehicleFinancials{}
	start := time.Now()
	err := r.db.QueryRow(ctx, `
		SELECT vehicle_id, acquisition_cost, depreciation_method, depreciation_rate,
		       residual_value, emi_amount, lease_end_date, insurance_premium,
		       total_maintenance_cost, total_fuel_cost, total_toll_cost,
		       tco_cached, tco_calculated_at, created_at, updated_at
		FROM vehicle_financials WHERE vehicle_id = $1`, vehicleID,
	).Scan(
		&f.VehicleID, &f.AcquisitionCost, &f.DepreciationMethod, &f.DepreciationRate,
		&f.ResidualValue, &f.EMIAmount, &f.LeaseEndDate, &f.InsurancePremium,
		&f.TotalMaintenanceCost, &f.TotalFuelCost, &f.TotalTollCost,
		&f.TCOCached, &f.TCOCalculatedAt, &f.CreatedAt, &f.UpdatedAt,
	)
	metrics.RecordDBQuery(`SELECT vehicle_id, acquisition_cost, depreciation_method, depreciation_rate,
		       residual_value, emi_amount, lease_end_date, insurance_premium,
		       total_maintenance_cost, total_fuel_cost, total_toll_cost,
		       tco_cached, tco_calculated_at, created_at, updated_at
		FROM vehicle_financials WHERE vehicle_id = $1`, time.Since(start))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return f, nil
}
