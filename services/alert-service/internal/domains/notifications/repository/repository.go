package repository

import (
	"context"
	"fmt"

	"github.com/flyfleet/alert-service/internal/domains/notifications/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListLogs(ctx context.Context, page, limit int, status, channel, search string) ([]models.NotificationLog, int, error)
	GetLog(ctx context.Context, id uuid.UUID) (*models.NotificationLog, error)
	GetSettings(ctx context.Context) ([]models.NotificationSettings, error)
	UpdateSettings(ctx context.Context, s *models.NotificationSettings) error
	CreateLog(ctx context.Context, l *models.NotificationLog) error
	AddEvent(ctx context.Context, e *models.NotificationEvent) error

	// Preferences
	ListPreferences(ctx context.Context, page, limit int, search string) ([]models.UserNotificationPreference, int, error)
	GetUserPreference(ctx context.Context, userId uuid.UUID) (*models.UserNotificationPreference, error)
	UpdateUserPreference(ctx context.Context, p *models.UserNotificationPreference) error
	AddPreferenceAudit(ctx context.Context, a *models.PreferenceAudit) error
	GetGlobalDefaults(ctx context.Context) (*models.GlobalNotificationDefault, error)
	UpdateGlobalDefaults(ctx context.Context, d *models.GlobalNotificationDefault) error
	GetPreferenceAudit(ctx context.Context, userId uuid.UUID) ([]models.PreferenceAudit, error)
}

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) ListLogs(ctx context.Context, page, limit int, status, channel, search string) ([]models.NotificationLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	
	query := `SELECT id, correlation_id, user_id, user_email, channel, type, subject, status, created_at, updated_at 
	          FROM notification_logs WHERE 1=1`
	args := []any{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}
	if channel != "" {
		query += fmt.Sprintf(" AND channel = $%d", argCount)
		args = append(args, channel)
		argCount++
	}
	if search != "" {
		query += fmt.Sprintf(" AND (user_email ILIKE $%d OR subject ILIKE $%d OR id::text ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS total", query)
	var total int
	err := r.db.QueryRow(ctx, totalQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.NotificationLog
	for rows.Next() {
		var l models.NotificationLog
		err := rows.Scan(&l.ID, &l.CorrelationID, &l.UserID, &l.UserEmail, &l.Channel, &l.Type, &l.Subject, &l.Status, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}

func (r *NotificationRepository) GetLog(ctx context.Context, id uuid.UUID) (*models.NotificationLog, error) {
	var l models.NotificationLog
	query := `SELECT id, correlation_id, user_id, user_email, channel, type, subject, body, status, provider, provider_response, retry_count, max_retries, metadata, created_at, updated_at 
	          FROM notification_logs WHERE id = $1`
	
	err := r.db.QueryRow(ctx, query, id).Scan(&l.ID, &l.CorrelationID, &l.UserID, &l.UserEmail, &l.Channel, &l.Type, &l.Subject, &l.Body, &l.Status, &l.Provider, &l.ProviderResponse, &l.RetryCount, &l.MaxRetries, &l.Metadata, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Fetch events/timeline
	eventRows, err := r.db.Query(ctx, "SELECT id, notification_id, event_type, message, created_at FROM notification_events WHERE notification_id = $1 ORDER BY created_at ASC", id)
	if err != nil {
		return nil, err
	}
	defer eventRows.Close()

	for eventRows.Next() {
		var e models.NotificationEvent
		err := eventRows.Scan(&e.ID, &e.NotificationID, &e.EventType, &e.Message, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		l.Items = append(l.Items, e)
	}

	return &l, nil
}

func (r *NotificationRepository) GetSettings(ctx context.Context) ([]models.NotificationSettings, error) {
	rows, err := r.db.Query(ctx, "SELECT id, channel, enabled, max_retries, priority, retention_months, updated_at FROM notification_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.NotificationSettings
	for rows.Next() {
		var s models.NotificationSettings
		err := rows.Scan(&s.ID, &s.Channel, &s.Enabled, &s.MaxRetries, &s.Priority, &s.RetentionMonths, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *NotificationRepository) UpdateSettings(ctx context.Context, s *models.NotificationSettings) error {
	query := `INSERT INTO notification_settings (channel, enabled, max_retries, priority, retention_months, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP) 
	          ON CONFLICT (channel) DO UPDATE SET 
	          enabled = EXCLUDED.enabled, 
	          max_retries = EXCLUDED.max_retries, 
	          priority = EXCLUDED.priority, 
	          retention_months = EXCLUDED.retention_months, 
	          updated_at = EXCLUDED.updated_at`
	
	_, err := r.db.Exec(ctx, query, s.Channel, s.Enabled, s.MaxRetries, s.Priority, s.RetentionMonths)
	return err
}

func (r *NotificationRepository) CreateLog(ctx context.Context, l *models.NotificationLog) error {
	l.ID = uuid.New()
	query := `INSERT INTO notification_logs (id, correlation_id, user_id, user_email, channel, type, subject, body, status, provider, max_retries, metadata, created_at, created_by, updated_at, updated_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP, $13, CURRENT_TIMESTAMP, $14) RETURNING created_at, updated_at`
	
	return r.db.QueryRow(ctx, query, l.ID, l.CorrelationID, l.UserID, l.UserEmail, l.Channel, l.Type, l.Subject, l.Body, l.Status, l.Provider, l.MaxRetries, l.Metadata, l.CreatedBy, l.UpdatedBy).Scan(&l.CreatedAt, &l.UpdatedAt)
}
func (r *NotificationRepository) AddEvent(ctx context.Context, e *models.NotificationEvent) error {
	e.ID = uuid.New()
	query := `INSERT INTO notification_events (id, notification_id, event_type, message) VALUES ($1, $2, $3, $4) RETURNING created_at`
	return r.db.QueryRow(ctx, query, e.ID, e.NotificationID, e.EventType, e.Message).Scan(&e.CreatedAt)
}

func (r *NotificationRepository) ListPreferences(ctx context.Context, page, limit int, search string) ([]models.UserNotificationPreference, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	query := `SELECT p.id, p.user_id, u.name as user_name, u.email as user_email, p.email_enabled, p.sms_enabled, p.push_enabled, p.in_app_enabled, p.last_modified_by, p.last_modified_at, p.created_at, p.updated_at 
	          FROM user_notification_preferences p
	          JOIN users u ON p.user_id = u.id
	          WHERE 1=1`
	args := []any{}
	argCount := 1

	if search != "" {
		query += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.name ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS total", query)
	var total int
	err := r.db.QueryRow(ctx, totalQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(" ORDER BY p.last_modified_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var prefs []models.UserNotificationPreference
	for rows.Next() {
		var p models.UserNotificationPreference
		err := rows.Scan(&p.ID, &p.UserID, &p.UserName, &p.UserEmail, &p.EmailEnabled, &p.SmsEnabled, &p.PushEnabled, &p.InAppEnabled, &p.LastModifiedBy, &p.LastModifiedAt, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		prefs = append(prefs, p)
	}

	return prefs, total, nil
}

func (r *NotificationRepository) GetUserPreference(ctx context.Context, userId uuid.UUID) (*models.UserNotificationPreference, error) {
	var p models.UserNotificationPreference
	query := `SELECT p.id, p.user_id, u.name as user_name, u.email as user_email, p.email_enabled, p.sms_enabled, p.push_enabled, p.in_app_enabled, p.last_modified_by, p.last_modified_at, p.created_at, p.updated_at 
	          FROM user_notification_preferences p
	          JOIN users u ON p.user_id = u.id
	          WHERE p.user_id = $1`
	
	err := r.db.QueryRow(ctx, query, userId).Scan(&p.ID, &p.UserID, &p.UserName, &p.UserEmail, &p.EmailEnabled, &p.SmsEnabled, &p.PushEnabled, &p.InAppEnabled, &p.LastModifiedBy, &p.LastModifiedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Try to fallback to user table if preference entry doesn't exist yet
			var name, email string
			errUser := r.db.QueryRow(ctx, "SELECT name, email FROM users WHERE id = $1", userId).Scan(&name, &email)
			if errUser != nil {
				if errUser == pgx.ErrNoRows {
					return nil, nil
				}
				return nil, errUser
			}
			// Return default settings
			return &models.UserNotificationPreference{
				UserID:       userId,
				UserName:     name,
				UserEmail:    email,
				EmailEnabled: true,
				SmsEnabled:   false,
				PushEnabled:  true,
				InAppEnabled: true,
			}, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *NotificationRepository) UpdateUserPreference(ctx context.Context, p *models.UserNotificationPreference) error {
	query := `INSERT INTO user_notification_preferences (user_id, email_enabled, sms_enabled, push_enabled, in_app_enabled, last_modified_by, last_modified_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
	          ON CONFLICT (user_id) DO UPDATE SET 
	          email_enabled = EXCLUDED.email_enabled, 
	          sms_enabled = EXCLUDED.sms_enabled, 
	          push_enabled = EXCLUDED.push_enabled, 
	          in_app_enabled = EXCLUDED.in_app_enabled, 
	          last_modified_by = EXCLUDED.last_modified_by, 
	          last_modified_at = EXCLUDED.last_modified_at, 
	          updated_at = EXCLUDED.updated_at`
	
	_, err := r.db.Exec(ctx, query, p.UserID, p.EmailEnabled, p.SmsEnabled, p.PushEnabled, p.InAppEnabled, p.LastModifiedBy)
	return err
}

func (r *NotificationRepository) AddPreferenceAudit(ctx context.Context, a *models.PreferenceAudit) error {
	a.ID = uuid.New()
	query := `INSERT INTO notification_preference_audit (id, user_id, changed_by, old_values, new_values, change_reason) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	return r.db.QueryRow(ctx, query, a.ID, a.UserID, a.ChangedBy, a.OldValues, a.NewValues, a.ChangeReason).Scan(&a.CreatedAt)
}

func (r *NotificationRepository) GetGlobalDefaults(ctx context.Context) (*models.GlobalNotificationDefault, error) {
	var d models.GlobalNotificationDefault
	query := `SELECT id, email_enabled, sms_enabled, push_enabled, in_app_enabled, updated_by, updated_at 
	          FROM global_notification_defaults ORDER BY updated_at DESC LIMIT 1`
	
	err := r.db.QueryRow(ctx, query).Scan(&d.ID, &d.EmailEnabled, &d.SmsEnabled, &d.PushEnabled, &d.InAppEnabled, &d.UpdatedBy, &d.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *NotificationRepository) UpdateGlobalDefaults(ctx context.Context, d *models.GlobalNotificationDefault) error {
	query := `UPDATE global_notification_defaults SET 
	          email_enabled = $1, sms_enabled = $2, push_enabled = $3, in_app_enabled = $4, 
	          updated_by = $5, updated_at = CURRENT_TIMESTAMP`
	
	_, err := r.db.Exec(ctx, query, d.EmailEnabled, d.SmsEnabled, d.PushEnabled, d.InAppEnabled, d.UpdatedBy)
	return err
}

func (r *NotificationRepository) GetPreferenceAudit(ctx context.Context, userId uuid.UUID) ([]models.PreferenceAudit, error) {
	query := `SELECT id, user_id, changed_by, old_values, new_values, change_reason, created_at 
	          FROM notification_preference_audit WHERE user_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []models.PreferenceAudit
	for rows.Next() {
		var a models.PreferenceAudit
		err := rows.Scan(&a.ID, &a.UserID, &a.ChangedBy, &a.OldValues, &a.NewValues, &a.ChangeReason, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		audits = append(audits, a)
	}
	return audits, nil
}
