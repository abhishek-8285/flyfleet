package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/alerts/models"
	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateLog(ctx context.Context, log *models.NotificationLog) error {
	query := `
		INSERT INTO alerts_v2_logs (
			id, user_id, recipient, channel, alert_type, subject, body, status, 
			retry_count, max_retries, provider, metadata, scheduled_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	log.UpdatedAt = log.CreatedAt

	metadataJSON, _ := json.Marshal(log.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.Recipient, log.Channel, log.AlertType, log.Subject, log.Body, log.Status,
		log.RetryCount, log.MaxRetries, log.Provider, metadataJSON, log.ScheduledAt, log.CreatedAt, log.UpdatedAt,
	)
	return err
}

func (r *postgresRepository) UpdateLog(ctx context.Context, log *models.NotificationLog) error {
	query := `
		UPDATE alerts_v2_logs SET
			status = $1,
			retry_count = $2,
			provider = $3,
			provider_response = $4,
			error_message = $5,
			updated_at = $6
		WHERE id = $7
	`
	log.UpdatedAt = time.Now()
	responseJSON, _ := json.Marshal(log.ProviderResponse)

	_, err := r.db.ExecContext(ctx, query,
		log.Status, log.RetryCount, log.Provider, responseJSON, log.ErrorMessage, log.UpdatedAt, log.ID,
	)
	return err
}

func (r *postgresRepository) GetLog(ctx context.Context, id uuid.UUID) (*models.NotificationLog, error) {
	query := `
		SELECT id, user_id, recipient, channel, alert_type, subject, body, status, 
		       retry_count, max_retries, provider, provider_response, error_message, 
		       metadata, scheduled_at, created_at, updated_at
		FROM alerts_v2_logs WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	var log models.NotificationLog
	var metadataJSON, responseJSON []byte
	err := row.Scan(
		&log.ID, &log.UserID, &log.Recipient, &log.Channel, &log.AlertType, &log.Subject, &log.Body, &log.Status,
		&log.RetryCount, &log.MaxRetries, &log.Provider, &responseJSON, &log.ErrorMessage,
		&metadataJSON, &log.ScheduledAt, &log.CreatedAt, &log.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadataJSON, &log.Metadata)
	json.Unmarshal(responseJSON, &log.ProviderResponse)
	return &log, nil
}

func (r *postgresRepository) ListPendingLogs(ctx context.Context, limit int) ([]models.NotificationLog, error) {
	query := `
		SELECT id, user_id, recipient, channel, alert_type, subject, body, status, 
		       retry_count, max_retries, provider, provider_response, error_message, 
		       metadata, scheduled_at, created_at, updated_at
		FROM alerts_v2_logs 
		WHERE status = 'QUEUED' AND scheduled_at <= $1
		ORDER BY scheduled_at ASC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.NotificationLog
	for rows.Next() {
		var log models.NotificationLog
		var metadataJSON, responseJSON []byte
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Recipient, &log.Channel, &log.AlertType, &log.Subject, &log.Body, &log.Status,
			&log.RetryCount, &log.MaxRetries, &log.Provider, &responseJSON, &log.ErrorMessage,
			&metadataJSON, &log.ScheduledAt, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(metadataJSON, &log.Metadata)
		json.Unmarshal(responseJSON, &log.ProviderResponse)
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *postgresRepository) GetTemplate(ctx context.Context, name string, channel models.Channel, language string, tenantID uuid.UUID) (*models.Template, error) {
	query := `
		SELECT id, name, channel, language, subject, body, metadata, tenant_id, created_at, updated_at
		FROM alerts_v2_templates
		WHERE name = $1 AND channel = $2 AND language = $3 AND tenant_id = $4
	`
	row := r.db.QueryRowContext(ctx, query, name, channel, language, tenantID)
	var t models.Template
	var metadataJSON []byte
	err := row.Scan(&t.ID, &t.Name, &t.Channel, &t.Language, &t.Subject, &t.Body, &metadataJSON, &t.TenantID, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadataJSON, &t.Metadata)
	return &t, nil
}

func (r *postgresRepository) CreateTemplate(ctx context.Context, t *models.Template) error {
	query := `
		INSERT INTO alerts_v2_templates (id, name, channel, language, subject, body, metadata, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = t.CreatedAt
	metadataJSON, _ := json.Marshal(t.Metadata)

	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Channel, t.Language, t.Subject, t.Body, metadataJSON, t.TenantID, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *postgresRepository) GetUserPreferences(ctx context.Context, userID, tenantID uuid.UUID) ([]models.UserPreference, error) {
	query := `SELECT user_id, tenant_id, channel, enabled, updated_at FROM alerts_v2_user_preferences WHERE user_id = $1 AND tenant_id = $2`
	rows, err := r.db.QueryContext(ctx, query, userID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []models.UserPreference
	for rows.Next() {
		var p models.UserPreference
		if err := rows.Scan(&p.UserID, &p.TenantID, &p.Channel, &p.Enabled, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func (r *postgresRepository) UpsertUserPreference(ctx context.Context, pref *models.UserPreference) error {
	query := `
		INSERT INTO alerts_v2_user_preferences (user_id, tenant_id, channel, enabled, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, tenant_id, channel) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			updated_at = EXCLUDED.updated_at
	`
	pref.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, pref.UserID, pref.TenantID, pref.Channel, pref.Enabled, pref.UpdatedAt)
	return err
}

func (r *postgresRepository) AddToDLQ(ctx context.Context, entry *models.DLQEntry) error {
	query := `
		INSERT INTO alerts_v2_dlq (id, log_id, reason, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	payloadJSON, _ := json.Marshal(entry.Payload)

	_, err := r.db.ExecContext(ctx, query, entry.ID, entry.NotificationID, entry.Reason, payloadJSON, entry.CreatedAt)
	return err
}
