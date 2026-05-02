package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/india/models"
	"github.com/flyfleet/fleet-core/internal/domains/india/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *IndiaIntegrationRepositoryImpl) CreateIntegrationConfig(ctx context.Context, config *models.IntegrationConfig) error {
	query := `
		INSERT INTO integration_configs (
			id, tenant_id, service_type, api_endpoint, api_key, username, password,
			client_id, client_secret, environment, active, last_sync_at, next_sync_at,
			sync_interval, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	// Encrypt sensitive fields
	apiKey, _ := utils.Encrypt(config.APIKey)
	username, _ := utils.Encrypt(config.Username)
	password, _ := utils.Encrypt(config.Password)
	clientSecret, _ := utils.Encrypt(config.ClientSecret)

	_, err := r.db.Exec(ctx, query,
		config.ID,
		config.TenantID,
		config.ServiceType,
		config.APIEndpoint,
		apiKey,
		username,
		password,
		config.ClientID,
		clientSecret,
		config.Environment,
		config.Active,
		config.LastSyncAt,
		config.NextSyncAt,
		config.SyncInterval,
		config.CreatedAt,
		config.CreatedBy,
		config.UpdatedAt,
		config.UpdatedBy,
	)

	return err
}

// GetIntegrationConfig retrieves an integration configuration by ID
func (r *IndiaIntegrationRepositoryImpl) GetIntegrationConfig(ctx context.Context, id string) (*models.IntegrationConfig, error) {
	query := `
		SELECT id, tenant_id, service_type, api_endpoint, api_key, username, password,
		       client_id, client_secret, environment, active, last_sync_at, next_sync_at,
		       sync_interval, created_at, created_by, updated_at, updated_by
		FROM integration_configs
		WHERE id = $1`

	var config models.IntegrationConfig

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&config.ID,
		&config.TenantID,
		&config.ServiceType,
		&config.APIEndpoint,
		&config.APIKey,
		&config.Username,
		&config.Password,
		&config.ClientID,
		&config.ClientSecret,
		&config.Environment,
		&config.Active,
		&config.LastSyncAt,
		&config.NextSyncAt,
		&config.SyncInterval,
		&config.CreatedAt,
		&config.CreatedBy,
		&config.UpdatedAt,
		&config.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get integration config: %w", err)
	}

	// Decrypt sensitive fields
	config.APIKey, _ = utils.Decrypt(config.APIKey)
	config.Username, _ = utils.Decrypt(config.Username)
	config.Password, _ = utils.Decrypt(config.Password)
	config.ClientSecret, _ = utils.Decrypt(config.ClientSecret)

	return &config, nil
}

// GetIntegrationConfigByServiceType retrieves an integration configuration by service type and organization
func (r *IndiaIntegrationRepositoryImpl) GetIntegrationConfigByServiceType(ctx context.Context, tenantID uuid.UUID, serviceType string) (*models.IntegrationConfig, error) {
	query := `
		SELECT id, tenant_id, service_type, api_endpoint, api_key, username, password,
		       client_id, client_secret, environment, active, last_sync_at, next_sync_at,
		       sync_interval, created_at, created_by, updated_at, updated_by
		FROM integration_configs
		WHERE tenant_id = $1 AND service_type = $2`

	var config models.IntegrationConfig

	row := r.db.QueryRow(ctx, query, tenantID, serviceType)
	err := row.Scan(
		&config.ID,
		&config.TenantID,
		&config.ServiceType,
		&config.APIEndpoint,
		&config.APIKey,
		&config.Username,
		&config.Password,
		&config.ClientID,
		&config.ClientSecret,
		&config.Environment,
		&config.Active,
		&config.LastSyncAt,
		&config.NextSyncAt,
		&config.SyncInterval,
		&config.CreatedAt,
		&config.CreatedBy,
		&config.UpdatedAt,
		&config.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("integration config not found")
		}
		return nil, fmt.Errorf("failed to get integration config by service type: %w", err)
	}

	// Decrypt sensitive fields
	config.APIKey, _ = utils.Decrypt(config.APIKey)
	config.Username, _ = utils.Decrypt(config.Username)
	config.Password, _ = utils.Decrypt(config.Password)
	config.ClientSecret, _ = utils.Decrypt(config.ClientSecret)

	return &config, nil
}

// UpdateIntegrationConfig updates an existing integration configuration
func (r *IndiaIntegrationRepositoryImpl) UpdateIntegrationConfig(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		if key == "updated_at" {
			continue
		}

		// Encrypt sensitive fields if present in updates
		if key == "api_key" || key == "username" || key == "password" || key == "client_secret" {
			if strVal, ok := value.(string); ok {
				encryptedVal, _ := utils.Encrypt(strVal)
				value = encryptedVal
			}
		}

		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	if setClause != "" {
		setClause += ", "
	}
	setClause += fmt.Sprintf("updated_at = $%d", argIndex)
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE integration_configs SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteIntegrationConfig deletes an integration configuration
func (r *IndiaIntegrationRepositoryImpl) DeleteIntegrationConfig(ctx context.Context, id string) error {
	query := "DELETE FROM integration_configs WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete integration config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("integration config not found")
	}

	return nil
}

// ListIntegrationConfigs retrieves integration configurations with filtering
func (r *IndiaIntegrationRepositoryImpl) ListIntegrationConfigs(ctx context.Context, tenantID uuid.UUID, serviceType *string, active *bool, limit, offset int) ([]*models.IntegrationConfig, int, error) {
	countQuery := "SELECT COUNT(*) FROM integration_configs WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIndex := 2

	whereClause := "WHERE tenant_id = $1"
	if serviceType != nil {
		whereClause += fmt.Sprintf(" AND service_type = $%d", argIndex)
		args = append(args, serviceType)
		argIndex++
	}
	if active != nil {
		whereClause += fmt.Sprintf(" AND active = $%d", argIndex)
		args = append(args, active)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count integration configs: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, service_type, api_endpoint, api_key, username, password,
		       client_id, client_secret, environment, active, last_sync_at, next_sync_at,
		       sync_interval, created_at, created_by, updated_at, updated_by
		FROM integration_configs
		%s
		ORDER BY service_type, created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list integration configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.IntegrationConfig
	for rows.Next() {
		var config models.IntegrationConfig

		err := rows.Scan(
			&config.ID,
			&config.TenantID,
			&config.ServiceType,
			&config.APIEndpoint,
			&config.APIKey,
			&config.Username,
			&config.Password,
			&config.ClientID,
			&config.ClientSecret,
			&config.Environment,
			&config.Active,
			&config.LastSyncAt,
			&config.NextSyncAt,
			&config.SyncInterval,
			&config.CreatedAt,
			&config.CreatedBy,
			&config.UpdatedAt,
			&config.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan integration config: %w", err)
		}

		// Decrypt sensitive fields
		config.APIKey, _ = utils.Decrypt(config.APIKey)
		config.Username, _ = utils.Decrypt(config.Username)
		config.Password, _ = utils.Decrypt(config.Password)
		config.ClientSecret, _ = utils.Decrypt(config.ClientSecret)

		configs = append(configs, &config)
	}

	return configs, totalCount, nil
}

// HealthCheck performs a health check on the database
