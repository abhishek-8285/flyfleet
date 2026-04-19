package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
)

// SettingsRepository handles maintenance settings operations
type SettingsRepository struct {
	db *pgxpool.Pool
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retrieves settings for an organization
func (r *SettingsRepository) Get(ctx context.Context, orgID uuid.UUID) (*models.MaintenanceSettings, error) {
	query := `
		SELECT tenant_id, sla_p1_hours, sla_p2_hours, sla_p3_hours,
			   low_stock_threshold, approval_threshold,
			   auto_assign_enabled, notification_enabled,
			   created_at, updated_at
		FROM maintenance_settings WHERE tenant_id = $1
	`
	settings := &models.MaintenanceSettings{}
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&settings.OrgID, &settings.SLAP1Hours, &settings.SLAP2Hours, &settings.SLAP3Hours,
		&settings.LowStockThreshold, &settings.ApprovalThreshold,
		&settings.AutoAssignEnabled, &settings.NotificationEnabled,
		&settings.CreatedAt, &settings.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return settings, err
}

// Upsert creates or updates settings for an organization
func (r *SettingsRepository) Upsert(ctx context.Context, settings *models.MaintenanceSettings) error {
	query := `
		INSERT INTO maintenance_settings (
			tenant_id, sla_p1_hours, sla_p2_hours, sla_p3_hours,
			low_stock_threshold, approval_threshold,
			auto_assign_enabled, notification_enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id) DO UPDATE SET
			sla_p1_hours = EXCLUDED.sla_p1_hours,
			sla_p2_hours = EXCLUDED.sla_p2_hours,
			sla_p3_hours = EXCLUDED.sla_p3_hours,
			low_stock_threshold = EXCLUDED.low_stock_threshold,
			approval_threshold = EXCLUDED.approval_threshold,
			auto_assign_enabled = EXCLUDED.auto_assign_enabled,
			notification_enabled = EXCLUDED.notification_enabled
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		settings.OrgID, settings.SLAP1Hours, settings.SLAP2Hours, settings.SLAP3Hours,
		settings.LowStockThreshold, settings.ApprovalThreshold,
		settings.AutoAssignEnabled, settings.NotificationEnabled,
	).Scan(&settings.CreatedAt, &settings.UpdatedAt)
}

// GetOrDefault retrieves settings or returns defaults if not found
func (r *SettingsRepository) GetOrDefault(ctx context.Context, orgID uuid.UUID) (*models.MaintenanceSettings, error) {
	settings, err := r.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return &models.MaintenanceSettings{
			OrgID:               orgID,
			SLAP1Hours:          4,
			SLAP2Hours:          24,
			SLAP3Hours:          72,
			LowStockThreshold:   5,
			ApprovalThreshold:   500.00,
			AutoAssignEnabled:   false,
			NotificationEnabled: true,
		}, nil
	}
	return settings, nil
}
