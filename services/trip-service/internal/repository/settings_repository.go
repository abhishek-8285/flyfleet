package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsRepository struct {
	db *pgxpool.Pool
}

func NewSettingsRepository(db *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetSettings(ctx context.Context, orgID string) (*models.TripSettings, *models.SettingsMeta, error) {
	var settingsJSON []byte
	var meta models.SettingsMeta
	var modifiedByID *string

	query := `SELECT settings, version, last_modified, modified_by FROM trip_settings WHERE org_id = $1`
	err := r.db.QueryRow(ctx, query, orgID).Scan(&settingsJSON, &meta.Version, &meta.LastModified, &modifiedByID)
	
	if err == pgx.ErrNoRows {
		return nil, nil, nil // Not found, service will handle defaults
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get trip settings: %w", err)
	}

	var settings models.TripSettings
	if err := json.Unmarshal(settingsJSON, &settings); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	meta.ModifiedBy = modifiedByID
	return &settings, &meta, nil
}

func (r *SettingsRepository) UpdateSettings(ctx context.Context, orgID string, settings models.TripSettings, userID string, version string) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO trip_settings (org_id, settings, version, last_modified, modified_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (org_id) DO UPDATE
		SET settings = EXCLUDED.settings,
		    version = EXCLUDED.version,
		    last_modified = EXCLUDED.last_modified,
		    modified_by = EXCLUDED.modified_by,
		    updated_at = NOW()`
	
	_, err = r.db.Exec(ctx, query, orgID, settingsJSON, version, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update trip settings: %w", err)
	}

	return nil
}
