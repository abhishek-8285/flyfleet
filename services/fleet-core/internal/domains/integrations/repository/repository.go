package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/integrations/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type Repository interface {
	// Integrations
	CreateIntegration(ctx context.Context, integration *models.Integration) error
	GetIntegration(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Integration, error)
	ListIntegrations(ctx context.Context, orgID uuid.UUID) ([]*models.Integration, error)
	UpdateIntegration(ctx context.Context, integration *models.Integration) error
	DeleteIntegration(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error

	// API Clients
	CreateAPIClient(ctx context.Context, client *models.APIClient) error
	GetAPIClient(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.APIClient, error)
	GetAPIClientByClientID(ctx context.Context, clientID string) (*models.APIClient, error)
	ListAPIClients(ctx context.Context, orgID uuid.UUID) ([]*models.APIClient, error)
	UpdateAPIClient(ctx context.Context, client *models.APIClient) error
	DeleteAPIClient(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error

	// Webhooks
	CreateWebhook(ctx context.Context, webhook *models.Webhook) error
	GetWebhook(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Webhook, error)
	ListWebhooks(ctx context.Context, orgID uuid.UUID) ([]*models.Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *models.Webhook) error
	DeleteWebhook(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error

	// Logs
	CreateIntegrationLog(ctx context.Context, log *models.IntegrationLog) error
	ListIntegrationLogs(ctx context.Context, orgID uuid.UUID, integrationID *uuid.UUID, limit int, offset int) ([]*models.IntegrationLog, error)
}

type postgresRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &postgresRepository{db: db, logger: logger}
}

// Integration Implementations

func (r *postgresRepository) CreateIntegration(ctx context.Context, i *models.Integration) error {
	configJSON, _ := json.Marshal(i.Config)
	query := `INSERT INTO integrations (id, org_id, name, type, provider, status, config, created_by, updated_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, i.ID, i.OrgID, i.Name, i.Type, i.Provider, i.Status, configJSON, i.CreatedBy).Scan(&i.CreatedAt, &i.UpdatedAt)
}

func (r *postgresRepository) GetIntegration(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Integration, error) {
	query := `SELECT id, org_id, name, type, provider, status, config, last_sync, created_at, created_by, updated_at, updated_by 
	          FROM integrations WHERE id = $1 AND org_id = $2`
	i := &models.Integration{}
	var configJSON []byte
	err := r.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&i.ID, &i.OrgID, &i.Name, &i.Type, &i.Provider, &i.Status, &configJSON, &i.LastSync, &i.CreatedAt, &i.CreatedBy, &i.UpdatedAt, &i.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(configJSON, &i.Config)
	return i, nil
}

func (r *postgresRepository) ListIntegrations(ctx context.Context, orgID uuid.UUID) ([]*models.Integration, error) {
	query := `SELECT id, org_id, name, type, provider, status, config, last_sync, created_at, created_by, updated_at, updated_by 
	          FROM integrations WHERE org_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []*models.Integration
	for rows.Next() {
		i := &models.Integration{}
		var configJSON []byte
		if err := rows.Scan(&i.ID, &i.OrgID, &i.Name, &i.Type, &i.Provider, &i.Status, &configJSON, &i.LastSync, &i.CreatedAt, &i.CreatedBy, &i.UpdatedAt, &i.UpdatedBy); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(configJSON, &i.Config)
		integrations = append(integrations, i)
	}
	return integrations, nil
}

func (r *postgresRepository) UpdateIntegration(ctx context.Context, i *models.Integration) error {
	configJSON, _ := json.Marshal(i.Config)
	query := `UPDATE integrations SET name = $1, type = $2, provider = $3, status = $4, config = $5, updated_at = CURRENT_TIMESTAMP, updated_by = $6 
	          WHERE id = $7 AND org_id = $8`
	_, err := r.db.ExecContext(ctx, query, i.Name, i.Type, i.Provider, i.Status, configJSON, i.UpdatedBy, i.ID, i.OrgID)
	return err
}

func (r *postgresRepository) DeleteIntegration(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	query := `DELETE FROM integrations WHERE id = $1 AND org_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, orgID)
	return err
}

// API Client Implementations

func (r *postgresRepository) CreateAPIClient(ctx context.Context, c *models.APIClient) error {
	query := `INSERT INTO api_clients (id, org_id, client_id, client_secret_hash, name, scopes, expiry, status, created_by, updated_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, c.ID, c.OrgID, c.ClientID, c.ClientSecretHash, c.Name, pq.Array(c.Scopes), c.Expiry, c.Status, c.CreatedBy).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (r *postgresRepository) GetAPIClient(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.APIClient, error) {
	query := `SELECT id, org_id, client_id, client_secret_hash, name, scopes, expiry, status, created_at, created_by, updated_at, updated_by 
	          FROM api_clients WHERE id = $1 AND org_id = $2`
	c := &models.APIClient{}
	err := r.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&c.ID, &c.OrgID, &c.ClientID, &c.ClientSecretHash, &c.Name, pq.Array(&c.Scopes), &c.Expiry, &c.Status, &c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy,
	)
	return c, err
}

func (r *postgresRepository) GetAPIClientByClientID(ctx context.Context, clientID string) (*models.APIClient, error) {
	query := `SELECT id, org_id, client_id, client_secret_hash, name, scopes, expiry, status, created_at, created_by, updated_at, updated_by 
	          FROM api_clients WHERE client_id = $1`
	c := &models.APIClient{}
	err := r.db.QueryRowContext(ctx, query, clientID).Scan(
		&c.ID, &c.OrgID, &c.ClientID, &c.ClientSecretHash, &c.Name, pq.Array(&c.Scopes), &c.Expiry, &c.Status, &c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy,
	)
	return c, err
}

func (r *postgresRepository) ListAPIClients(ctx context.Context, orgID uuid.UUID) ([]*models.APIClient, error) {
	query := `SELECT id, org_id, client_id, client_secret_hash, name, scopes, expiry, status, created_at, created_by, updated_at, updated_by 
	          FROM api_clients WHERE org_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*models.APIClient
	for rows.Next() {
		c := &models.APIClient{}
		if err := rows.Scan(&c.ID, &c.OrgID, &c.ClientID, &c.ClientSecretHash, &c.Name, pq.Array(&c.Scopes), &c.Expiry, &c.Status, &c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, nil
}

func (r *postgresRepository) UpdateAPIClient(ctx context.Context, c *models.APIClient) error {
	query := `UPDATE api_clients SET name = $1, scopes = $2, expiry = $3, status = $4, updated_at = CURRENT_TIMESTAMP, updated_by = $5 
	          WHERE id = $6 AND org_id = $7`
	_, err := r.db.ExecContext(ctx, query, c.Name, pq.Array(c.Scopes), c.Expiry, c.Status, c.UpdatedBy, c.ID, c.OrgID)
	return err
}

func (r *postgresRepository) DeleteAPIClient(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	query := `DELETE FROM api_clients WHERE id = $1 AND org_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, orgID)
	return err
}

// Webhook Implementations

func (r *postgresRepository) CreateWebhook(ctx context.Context, w *models.Webhook) error {
	query := `INSERT INTO webhooks (id, org_id, url, events, secret, status, created_by, updated_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $7) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, w.ID, w.OrgID, w.URL, pq.Array(w.Events), w.Secret, w.Status, w.CreatedBy).Scan(&w.CreatedAt, &w.UpdatedAt)
}

func (r *postgresRepository) GetWebhook(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Webhook, error) {
	query := `SELECT id, org_id, url, events, secret, status, created_at, created_by, updated_at, updated_by 
	          FROM webhooks WHERE id = $1 AND org_id = $2`
	w := &models.Webhook{}
	err := r.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&w.ID, &w.OrgID, &w.URL, pq.Array(&w.Events), &w.Secret, &w.Status, &w.CreatedAt, &w.CreatedBy, &w.UpdatedAt, &w.UpdatedBy,
	)
	return w, err
}

func (r *postgresRepository) ListWebhooks(ctx context.Context, orgID uuid.UUID) ([]*models.Webhook, error) {
	query := `SELECT id, org_id, url, events, secret, status, created_at, created_by, updated_at, updated_by 
	          FROM webhooks WHERE org_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		w := &models.Webhook{}
		if err := rows.Scan(&w.ID, &w.OrgID, &w.URL, pq.Array(&w.Events), &w.Secret, &w.Status, &w.CreatedAt, &w.CreatedBy, &w.UpdatedAt, &w.UpdatedBy); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (r *postgresRepository) UpdateWebhook(ctx context.Context, w *models.Webhook) error {
	query := `UPDATE webhooks SET url = $1, events = $2, secret = $3, status = $4, updated_at = CURRENT_TIMESTAMP, updated_by = $5 
	          WHERE id = $6 AND org_id = $7`
	_, err := r.db.ExecContext(ctx, query, w.URL, pq.Array(w.Events), w.Secret, w.Status, w.UpdatedBy, w.ID, w.OrgID)
	return err
}

func (r *postgresRepository) DeleteWebhook(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1 AND org_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, orgID)
	return err
}

// Log Implementations

func (r *postgresRepository) CreateIntegrationLog(ctx context.Context, l *models.IntegrationLog) error {
	reqJSON, _ := json.Marshal(l.RequestBody)
	respJSON, _ := json.Marshal(l.ResponseBody)
	query := `INSERT INTO integration_logs (id, org_id, integration_id, method, endpoint, request_body, response_body, status_code, status, duration_ms) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.OrgID, l.IntegrationID, l.Method, l.Endpoint, reqJSON, respJSON, l.StatusCode, l.Status, l.DurationMS)
	return err
}

func (r *postgresRepository) ListIntegrationLogs(ctx context.Context, orgID uuid.UUID, integrationID *uuid.UUID, limit int, offset int) ([]*models.IntegrationLog, error) {
	query := `SELECT id, org_id, integration_id, method, endpoint, request_body, response_body, status_code, status, duration_ms, timestamp 
	          FROM integration_logs WHERE org_id = $1`
	args := []interface{}{orgID}

	if integrationID != nil {
		query += ` AND integration_id = $2`
		args = append(args, *integrationID)
	}

	query += fmt.Sprintf(` ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.IntegrationLog
	for rows.Next() {
		l := &models.IntegrationLog{}
		var reqJSON, respJSON []byte
		if err := rows.Scan(&l.ID, &l.OrgID, &l.IntegrationID, &l.Method, &l.Endpoint, &reqJSON, &respJSON, &l.StatusCode, &l.Status, &l.DurationMS, &l.Timestamp); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(reqJSON, &l.RequestBody)
		_ = json.Unmarshal(respJSON, &l.ResponseBody)
		logs = append(logs, l)
	}
	return logs, nil
}
