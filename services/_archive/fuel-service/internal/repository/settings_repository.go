package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fuel-service/internal/models"
)

type SettingsRepository interface {
	Get(ctx context.Context, orgID uuid.UUID) (*models.FuelSettings, error)
	Upsert(ctx context.Context, settings *models.FuelSettings) error
	GetOrDefault(ctx context.Context, orgID uuid.UUID) (*models.FuelSettings, error)
}

type settingsRepository struct {
	db *pgxpool.Pool
}

func NewSettingsRepository(db *pgxpool.Pool) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) Get(ctx context.Context, orgID uuid.UUID) (*models.FuelSettings, error) {
	query := `
		SELECT 
			organization_id, default_currency, allowed_fuel_types,
			validation_rules, auto_sync_enabled, created_at, updated_at
		FROM fuel_settings
		WHERE organization_id = $1
	`
	var s models.FuelSettings
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&s.OrganizationID, &s.DefaultCurrency, &s.AllowedFuelTypes,
		&s.ValidationRules, &s.AutoSyncEnabled, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *settingsRepository) Upsert(ctx context.Context, s *models.FuelSettings) error {
	query := `
		INSERT INTO fuel_settings (
			organization_id, default_currency, allowed_fuel_types,
			validation_rules, auto_sync_enabled, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (organization_id) DO UPDATE SET
			default_currency = EXCLUDED.default_currency,
			allowed_fuel_types = EXCLUDED.allowed_fuel_types,
			validation_rules = EXCLUDED.validation_rules,
			auto_sync_enabled = EXCLUDED.auto_sync_enabled,
			updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query,
		s.OrganizationID, s.DefaultCurrency, s.AllowedFuelTypes,
		s.ValidationRules, s.AutoSyncEnabled,
	)
	return err
}

func (r *settingsRepository) GetOrDefault(ctx context.Context, orgID uuid.UUID) (*models.FuelSettings, error) {
	s, err := r.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		// Return defaults
		return &models.FuelSettings{
			OrganizationID:   orgID,
			DefaultCurrency:  "USD",
			AllowedFuelTypes: []byte(`["diesel", "petrol-95", "adblue", "ev"]`),
			ValidationRules:  []byte(`{"odometer_check": true, "cost_calculation_verify": true}`),
			AutoSyncEnabled:  true,
		}, nil
	}
	return s, nil
}
