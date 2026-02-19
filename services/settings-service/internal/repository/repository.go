package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/settings-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("record not found")

type Repository interface {
	// Platform Settings
	GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error)
	UpdatePlatformSettings(ctx context.Context, s *models.PlatformSettings) error

	// Feature Flags
	ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error)
	GetFeatureFlag(ctx context.Context, id uuid.UUID) (*models.FeatureFlag, error)
	CreateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error
	UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error
	DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error

	// Localization
	GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error)
	UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error

	// Templates
	ListTemplates(ctx context.Context) ([]models.Template, error)
	GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error)
	CreateTemplate(ctx context.Context, t *models.Template) error
	UpdateTemplate(ctx context.Context, t *models.Template) error

	// Notification Channels
	ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error)
	UpdateNotificationChannel(ctx context.Context, c *models.NotificationChannel) error

	// System Limits
	GetSystemLimits(ctx context.Context) (*models.SystemLimit, error)
	UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error

	// Audit
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	ListAuditLogs(ctx context.Context, query models.AuditLogQuery) ([]models.AuditLog, error)

	// Transactions
	WithTx(ctx context.Context, fn func(Repository) error) error
}

type PostgresRepository struct {
	db     *pgxpool.Pool
	tx     pgx.Tx
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db, logger: logger}
}

func (r *PostgresRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if r.tx != nil {
		return r.tx.Query(ctx, sql, args...)
	}
	return r.query(ctx, sql, args...)
}

func (r *PostgresRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if r.tx != nil {
		return r.tx.QueryRow(ctx, sql, args...)
	}
	return r.queryRow(ctx, sql, args...)
}

func (r *PostgresRepository) exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if r.tx != nil {
		return r.tx.Exec(ctx, sql, args...)
	}
	return r.exec(ctx, sql, args...)
}

func (r *PostgresRepository) WithTx(ctx context.Context, fn func(Repository) error) error {
	if r.tx != nil {
		return fn(r)
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	txRepo := &PostgresRepository{db: r.db, tx: tx, logger: r.logger}
	if err := fn(txRepo); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func defaultPlatformSettings() *models.PlatformSettings {
	return &models.PlatformSettings{
		ID:              models.PlatformSettingsSingletonID,
		SystemName:      "Enterprise Fleet Manager",
		SupportEmail:    "support@fleet-system.com",
		DefaultTimezone: "UTC",
		DefaultCurrency: "USD",
		DefaultLanguage: "en-US",
	}
}

// Platform Settings
func (r *PostgresRepository) GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error) {
	var s models.PlatformSettings
	// There should only be one row, but we use ORDER BY created_at DESC LIMIT 1 to get the latest if multiple exist (shouldn't happen with correct logic)
	query := `SELECT id, system_name, support_email, default_timezone, default_currency, default_language, white_label_enabled, smtp_config_id, created_at, updated_at 
	          FROM platform_settings ORDER BY created_at DESC LIMIT 1`
	err := r.queryRow(ctx, query).Scan(&s.ID, &s.SystemName, &s.SupportEmail, &s.DefaultTimezone, &s.DefaultCurrency, &s.DefaultLanguage, &s.WhiteLabel, &s.SMTPConfigID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return defaultPlatformSettings(), nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			r.logger.Warn("Settings schema unavailable; serving default platform settings", zap.String("code", pgErr.Code))
			return defaultPlatformSettings(), nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) UpdatePlatformSettings(ctx context.Context, s *models.PlatformSettings) error {
	// Upsert logic
	query := `INSERT INTO platform_settings (id, system_name, support_email, default_timezone, default_currency, default_language, white_label_enabled, smtp_config_id, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	          ON CONFLICT (id) DO UPDATE SET
	          system_name = EXCLUDED.system_name,
	          support_email = EXCLUDED.support_email,
	          default_timezone = EXCLUDED.default_timezone,
	          default_currency = EXCLUDED.default_currency,
	          default_language = EXCLUDED.default_language,
	          white_label_enabled = EXCLUDED.white_label_enabled,
	          smtp_config_id = EXCLUDED.smtp_config_id,
	          updated_at = NOW()`

	// Platform settings are a singleton record. Always write the canonical row ID
	// so updates cannot accidentally create additional rows.
	s.ID = models.PlatformSettingsSingletonID
	_, err := r.exec(ctx, query, s.ID, s.SystemName, s.SupportEmail, s.DefaultTimezone, s.DefaultCurrency, s.DefaultLanguage, s.WhiteLabel, s.SMTPConfigID)
	return err
}

// Feature Flags
func (r *PostgresRepository) ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	query := `SELECT id, key, description, scope, enabled, tags, created_by, updated_at FROM feature_flags ORDER BY key`
	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []models.FeatureFlag
	for rows.Next() {
		var f models.FeatureFlag
		if err := rows.Scan(&f.ID, &f.Key, &f.Description, &f.Scope, &f.Enabled, &f.Tags, &f.CreatedBy, &f.UpdatedAt); err != nil {
			return nil, err
		}
		flags = append(flags, f)
	}
	return flags, nil
}

func (r *PostgresRepository) GetFeatureFlag(ctx context.Context, id uuid.UUID) (*models.FeatureFlag, error) {
	var f models.FeatureFlag
	query := `SELECT id, key, description, scope, enabled, tags, created_by, updated_at FROM feature_flags WHERE id = $1`
	err := r.queryRow(ctx, query, id).Scan(&f.ID, &f.Key, &f.Description, &f.Scope, &f.Enabled, &f.Tags, &f.CreatedBy, &f.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

func (r *PostgresRepository) CreateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	f.ID = uuid.New()
	f.UpdatedAt = time.Now()
	query := `INSERT INTO feature_flags (id, key, description, scope, enabled, tags, created_by, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.exec(ctx, query, f.ID, f.Key, f.Description, f.Scope, f.Enabled, f.Tags, f.CreatedBy, f.UpdatedAt)
	return err
}

func (r *PostgresRepository) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	f.UpdatedAt = time.Now()
	query := `UPDATE feature_flags SET description = $2, scope = $3, enabled = $4, tags = $5, updated_at = $6 WHERE id = $1`
	res, err := r.exec(ctx, query, f.ID, f.Description, f.Scope, f.Enabled, f.Tags, f.UpdatedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM feature_flags WHERE id = $1`
	res, err := r.exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Localization
func (r *PostgresRepository) GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error) {
	query := `SELECT id, language_code, is_active, is_default FROM localization_settings ORDER BY language_code`
	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.LocalizationSettings
	for rows.Next() {
		var s models.LocalizationSettings
		if err := rows.Scan(&s.ID, &s.LanguageCode, &s.IsActive, &s.IsDefault); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *PostgresRepository) UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error {
	query := `INSERT INTO localization_settings (id, language_code, is_active, is_default)
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (language_code) DO UPDATE SET
	          is_active = EXCLUDED.is_active,
	          is_default = EXCLUDED.is_default`

	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	_, err := r.exec(ctx, query, l.ID, l.LanguageCode, l.IsActive, l.IsDefault)
	return err
}

// Templates
func (r *PostgresRepository) ListTemplates(ctx context.Context) ([]models.Template, error) {
	query := `SELECT id, template_key, channel, subject, body, variables, updated_by, updated_at FROM templates ORDER BY template_key`
	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.Template
	for rows.Next() {
		var t models.Template
		if err := rows.Scan(&t.ID, &t.Key, &t.Channel, &t.Subject, &t.Body, &t.Variables, &t.UpdatedBy, &t.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *PostgresRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	var t models.Template
	query := `SELECT id, template_key, channel, subject, body, variables, updated_by, updated_at FROM templates WHERE id = $1`
	err := r.queryRow(ctx, query, id).Scan(&t.ID, &t.Key, &t.Channel, &t.Subject, &t.Body, &t.Variables, &t.UpdatedBy, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRepository) CreateTemplate(ctx context.Context, t *models.Template) error {
	t.ID = uuid.New()
	t.UpdatedAt = time.Now()
	query := `INSERT INTO templates (id, template_key, channel, subject, body, variables, updated_by, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.exec(ctx, query, t.ID, t.Key, t.Channel, t.Subject, t.Body, t.Variables, t.UpdatedBy, t.UpdatedAt)
	return err
}

func (r *PostgresRepository) UpdateTemplate(ctx context.Context, t *models.Template) error {
	t.UpdatedAt = time.Now()
	query := `UPDATE templates SET subject = $2, body = $3, variables = $4, updated_by = $5, updated_at = $6 WHERE id = $1`
	res, err := r.exec(ctx, query, t.ID, t.Subject, t.Body, t.Variables, t.UpdatedBy, t.UpdatedAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Notification Channels
func (r *PostgresRepository) ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error) {
	query := `SELECT id, channel_type, provider, config, enabled, updated_at FROM notification_channels ORDER BY channel_type`
	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.NotificationChannel
	for rows.Next() {
		var c models.NotificationChannel
		if err := rows.Scan(&c.ID, &c.ChannelType, &c.Provider, &c.Config, &c.Enabled, &c.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, nil
}

func (r *PostgresRepository) UpdateNotificationChannel(ctx context.Context, c *models.NotificationChannel) error {
	query := `INSERT INTO notification_channels (id, channel_type, provider, config, enabled, updated_at)
	          VALUES ($1, $2, $3, $4, $5, NOW())
	          ON CONFLICT (channel_type) DO UPDATE SET
	          provider = EXCLUDED.provider,
	          config = EXCLUDED.config,
	          enabled = EXCLUDED.enabled,
	          updated_at = NOW()`

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	_, err := r.exec(ctx, query, c.ID, c.ChannelType, c.Provider, c.Config, c.Enabled)
	return err
}

// System Limits
func (r *PostgresRepository) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) {
	var l models.SystemLimit
	query := `SELECT id, api_rate_limit, burst_limit, audit_retention, activity_retention FROM system_limits LIMIT 1`
	err := r.queryRow(ctx, query).Scan(&l.ID, &l.APIRateLimit, &l.BurstLimit, &l.AuditRetention, &l.ActivityRetention)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return defaults if not found
			return &models.SystemLimit{
				APIRateLimit:      10000,
				BurstLimit:        500,
				AuditRetention:    "1y",
				ActivityRetention: "180d",
			}, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *PostgresRepository) UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error {
	query := `INSERT INTO system_limits (id, api_rate_limit, burst_limit, audit_retention, activity_retention)
	          VALUES ($1, $2, $3, $4, $5)
	          ON CONFLICT (id) DO UPDATE SET
	          api_rate_limit = EXCLUDED.api_rate_limit,
	          burst_limit = EXCLUDED.burst_limit,
	          audit_retention = EXCLUDED.audit_retention,
	          activity_retention = EXCLUDED.activity_retention`

	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	_, err := r.exec(ctx, query, l.ID, l.APIRateLimit, l.BurstLimit, l.AuditRetention, l.ActivityRetention)
	return err
}

// Audit
func (r *PostgresRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO settings_audit_logs (id, entity_type, entity_id, action, actor_id, actor_ip, changes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	_, err := r.exec(ctx, query,
		log.ID, log.EntityType, log.EntityID, log.Action, log.ActorID, log.ActorIP, log.Changes, log.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			r.logger.Warn("Audit schema unavailable; skipping audit write", zap.String("code", pgErr.Code))
			return nil
		}
	}
	return err
}

func (r *PostgresRepository) ListAuditLogs(ctx context.Context, query models.AuditLogQuery) ([]models.AuditLog, error) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 4)
	argPos := 1

	if query.EntityType != "" {
		clauses = append(clauses, fmt.Sprintf("entity_type = $%d", argPos))
		args = append(args, query.EntityType)
		argPos++
	}
	if query.Action != "" {
		clauses = append(clauses, fmt.Sprintf("action = $%d", argPos))
		args = append(args, query.Action)
		argPos++
	}

	sql := `SELECT id, entity_type, entity_id, action, actor_id, actor_ip, changes, created_at FROM settings_audit_logs`
	if len(clauses) > 0 {
		sql += " WHERE " + strings.Join(clauses, " AND ")
	}
	sql += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, query.Limit, query.Offset)

	rows, err := r.query(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			r.logger.Warn("Audit schema unavailable; returning empty audit log set", zap.String("code", pgErr.Code))
			return []models.AuditLog{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	logs := make([]models.AuditLog, 0, query.Limit)
	for rows.Next() {
		var entry models.AuditLog
		if err := rows.Scan(
			&entry.ID,
			&entry.EntityType,
			&entry.EntityID,
			&entry.Action,
			&entry.ActorID,
			&entry.ActorIP,
			&entry.Changes,
			&entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}

	return logs, nil
}
