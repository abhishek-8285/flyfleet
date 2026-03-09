package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ComplianceRepository struct {
	db *pgxpool.Pool
}

func NewComplianceRepository(db *pgxpool.Pool) *ComplianceRepository {
	return &ComplianceRepository{db: db}
}

func (r *ComplianceRepository) Upsert(ctx context.Context, c *models.VehicleCompliance) error {
	c.UpdatedAt = time.Now()
	query := `
		INSERT INTO vehicle_compliance (
			vehicle_id, rc_expiry_date, insurance_expiry_date, puc_expiry_date,
			fitness_expiry_date, permit_expiry_date, permit_type, insurance_policy_no,
			is_blacklisted, legal_hold, last_verified_at, verified_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		ON CONFLICT (vehicle_id) DO UPDATE SET
			rc_expiry_date        = EXCLUDED.rc_expiry_date,
			insurance_expiry_date = EXCLUDED.insurance_expiry_date,
			puc_expiry_date       = EXCLUDED.puc_expiry_date,
			fitness_expiry_date   = EXCLUDED.fitness_expiry_date,
			permit_expiry_date    = EXCLUDED.permit_expiry_date,
			permit_type           = EXCLUDED.permit_type,
			insurance_policy_no   = EXCLUDED.insurance_policy_no,
			is_blacklisted        = EXCLUDED.is_blacklisted,
			legal_hold            = EXCLUDED.legal_hold,
			last_verified_at      = EXCLUDED.last_verified_at,
			verified_by           = EXCLUDED.verified_by,
			updated_at            = NOW()
		RETURNING created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		c.VehicleID, c.RCExpiryDate, c.InsuranceExpiryDate, c.PUCExpiryDate,
		c.FitnessExpiryDate, c.PermitExpiryDate, c.PermitType, c.InsurancePolicyNo,
		c.IsBlacklisted, c.LegalHold, c.LastVerifiedAt, c.VerifiedBy,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (r *ComplianceRepository) Get(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleCompliance, error) {
	c := &models.VehicleCompliance{}
	err := r.db.QueryRow(ctx, `
		SELECT vehicle_id, rc_expiry_date, insurance_expiry_date, puc_expiry_date,
		       fitness_expiry_date, permit_expiry_date, permit_type, insurance_policy_no,
		       is_blacklisted, legal_hold, last_verified_at, verified_by, created_at, updated_at
		FROM vehicle_compliance WHERE vehicle_id = $1`, vehicleID,
	).Scan(
		&c.VehicleID, &c.RCExpiryDate, &c.InsuranceExpiryDate, &c.PUCExpiryDate,
		&c.FitnessExpiryDate, &c.PermitExpiryDate, &c.PermitType, &c.InsurancePolicyNo,
		&c.IsBlacklisted, &c.LegalHold, &c.LastVerifiedAt, &c.VerifiedBy,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return c, nil
}

func (r *ComplianceRepository) GetExpiringVehicles(ctx context.Context, tenantID uuid.UUID, withinDays int) ([]models.ComplianceExpiryAlert, error) {
	query := `
		SELECT vehicle_id, tenant_id, registration_number, status,
		       TO_CHAR(earliest_expiry, 'YYYY-MM-DD') AS earliest_expiry,
		       days_to_expiry::integer
		FROM vehicles_due_for_compliance
		WHERE tenant_id = $1 AND days_to_expiry <= $2
		ORDER BY days_to_expiry ASC`

	rows, err := r.db.Query(ctx, query, tenantID, withinDays)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var alerts []models.ComplianceExpiryAlert
	for rows.Next() {
		a := models.ComplianceExpiryAlert{}
		if err := rows.Scan(&a.VehicleID, &a.TenantID, &a.RegistrationNumber, &a.Status, &a.EarliestExpiry, &a.DaysToExpiry); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}
