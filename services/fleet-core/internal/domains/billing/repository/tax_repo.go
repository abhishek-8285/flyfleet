package repository

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaxRepository struct {
	db *pgxpool.Pool
}

func NewTaxRepository(db *pgxpool.Pool) *TaxRepository {
	return &TaxRepository{db: db}
}

// GetTaxRatesByRegion fetches all active tax rates for a specific region
func (r *TaxRepository) GetTaxRatesByRegion(ctx context.Context, regionID uuid.UUID) ([]models.TaxRate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, region_id, name, percentage, tax_type, is_default, created_at
		FROM billing_tax_rates
		WHERE region_id = $1 AND is_active = TRUE
	`, regionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.TaxRate
	for rows.Next() {
		var rate models.TaxRate
		if err := rows.Scan(&rate.ID, &rate.RegionID, &rate.Name, &rate.Percentage, &rate.TaxType, &rate.IsDefault, &rate.CreatedAt); err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, nil
}

// GetRegionByCountry fetches the tax region for a country/state
func (r *TaxRepository) GetRegionByCountry(ctx context.Context, countryCode string, stateCode *string) (*models.TaxRegion, error) {
	var region models.TaxRegion
	query := `SELECT id, name, country_code, state_code, is_active, created_at FROM billing_tax_regions WHERE country_code = $1`
	args := []interface{}{countryCode}

	if stateCode != nil {
		query += " AND state_code = $2"
		args = append(args, *stateCode)
	} else {
		query += " AND state_code IS NULL"
	}

	err := r.db.QueryRow(ctx, query, args...).Scan(&region.ID, &region.Name, &region.CountryCode, &region.StateCode, &region.IsActive, &region.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &region, err
}

type SequenceRepository struct {
	db *pgxpool.Pool
}

func NewSequenceRepository(db *pgxpool.Pool) *SequenceRepository {
	return &SequenceRepository{db: db}
}

// GetNextInvoiceNumber increments and returns the next gapless sequential number
func (r *SequenceRepository) GetNextInvoiceNumber(ctx context.Context, orgID uuid.UUID) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var prefix, suffix string
	var nextValue int64

	// Lock the sequence row for this organization
	err = tx.QueryRow(ctx, `
		SELECT prefix, next_value, suffix FROM billing_invoice_sequences 
		WHERE org_id = $1 FOR UPDATE
	`, orgID).Scan(&prefix, &nextValue, &suffix)

	if err == pgx.ErrNoRows {
		// Initialize sequence if not exists
		prefix = "INV-"
		nextValue = 1
		suffix = ""
		_, err = tx.Exec(ctx, `
			INSERT INTO billing_invoice_sequences (org_id, prefix, next_value, suffix)
			VALUES ($1, $2, $3, $4)
		`, orgID, prefix, nextValue+1, suffix)
	} else if err == nil {
		// Update existing sequence
		_, err = tx.Exec(ctx, `
			UPDATE billing_invoice_sequences SET next_value = next_value + 1, updated_at = CURRENT_TIMESTAMP
			WHERE org_id = $1
		`, orgID)
	}

	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%05d%s", prefix, nextValue, suffix), nil
}
