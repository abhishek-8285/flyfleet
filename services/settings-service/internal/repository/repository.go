package repository

import (
	"context"
	"errors"
	"time"

	"github.com/flyfleet/settings-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
}

type PostgresRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db, logger: logger}
}

// Platform Settings
func (r *PostgresRepository) GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error) {
	var s models.PlatformSettings
	// There should only be one row, but we use ORDER BY created_at DESC LIMIT 1 to get the latest if multiple exist (shouldn't happen with correct logic)
	query := `SELECT id, system_name, support_email, default_timezone, default_currency, default_language, white_label_enabled, smtp_config_id, created_at, updated_at 
	          FROM platform_settings ORDER BY created_at DESC LIMIT 1`
	err := r.db.QueryRow(ctx, query).Scan(&s.ID, &s.SystemName, &s.SupportEmail, &s.DefaultTimezone, &s.DefaultCurrency, &s.DefaultLanguage, &s.WhiteLabel, &s.SMTPConfigID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default if no settings exist
			return &models.PlatformSettings{
				SystemName:      "Enterprise Fleet Manager",
				SupportEmail:    "support@fleet-system.com",
				DefaultTimezone: "UTC",
				DefaultCurrency: "USD",
				DefaultLanguage: "en-US",
			}, nil
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
	
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, query, s.ID, s.SystemName, s.SupportEmail, s.DefaultTimezone, s.DefaultCurrency, s.DefaultLanguage, s.WhiteLabel, s.SMTPConfigID)
	return err
}

// Feature Flags
func (r *PostgresRepository) ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	query := `SELECT id, key, description, scope, enabled, tags, created_by, updated_at FROM feature_flags ORDER BY key`
	rows, err := r.db.Query(ctx, query)
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
	err := r.db.QueryRow(ctx, query, id).Scan(&f.ID, &f.Key, &f.Description, &f.Scope, &f.Enabled, &f.Tags, &f.CreatedBy, &f.UpdatedAt)
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
	_, err := r.db.Exec(ctx, query, f.ID, f.Key, f.Description, f.Scope, f.Enabled, f.Tags, f.CreatedBy, f.UpdatedAt)
	return err
}

func (r *PostgresRepository) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	f.UpdatedAt = time.Now()
	query := `UPDATE feature_flags SET description = $2, scope = $3, enabled = $4, tags = $5, updated_at = $6 WHERE id = $1`
	res, err := r.db.Exec(ctx, query, f.ID, f.Description, f.Scope, f.Enabled, f.Tags, f.UpdatedAt)
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
	res, err := r.db.Exec(ctx, query, id)
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
	rows, err := r.db.Query(ctx, query)
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
	_, err := r.db.Exec(ctx, query, l.ID, l.LanguageCode, l.IsActive, l.IsDefault)
	return err
}

// Templates
func (r *PostgresRepository) ListTemplates(ctx context.Context) ([]models.Template, error) {
	query := `SELECT id, template_key, channel, subject, body, variables, updated_by, updated_at FROM templates ORDER BY template_key`
	rows, err := r.db.Query(ctx, query)
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
	err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.Key, &t.Channel, &t.Subject, &t.Body, &t.Variables, &t.UpdatedBy, &t.UpdatedAt)
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
	_, err := r.db.Exec(ctx, query, t.ID, t.Key, t.Channel, t.Subject, t.Body, t.Variables, t.UpdatedBy, t.UpdatedAt)
	return err
}

func (r *PostgresRepository) UpdateTemplate(ctx context.Context, t *models.Template) error {
	t.UpdatedAt = time.Now()
	query := `UPDATE templates SET subject = $2, body = $3, variables = $4, updated_by = $5, updated_at = $6 WHERE id = $1`
	res, err := r.db.Exec(ctx, query, t.ID, t.Subject, t.Body, t.Variables, t.UpdatedBy, t.UpdatedAt)
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
	rows, err := r.db.Query(ctx, query)
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
	_, err := r.db.Exec(ctx, query, c.ID, c.ChannelType, c.Provider, c.Config, c.Enabled)
	return err
}

// System Limits
func (r *PostgresRepository) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) {
	var l models.SystemLimit
	query := `SELECT id, api_rate_limit, burst_limit, audit_retention, activity_retention FROM system_limits LIMIT 1`
	err := r.db.QueryRow(ctx, query).Scan(&l.ID, &l.APIRateLimit, &l.BurstLimit, &l.AuditRetention, &l.ActivityRetention)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return defaults if not found
			return &models.SystemLimit{
				APIRateLimit: 10000,
				BurstLimit: 500,
				AuditRetention: "1y",
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
	_, err := r.db.Exec(ctx, query, l.ID, l.APIRateLimit, l.BurstLimit, l.AuditRetention, l.ActivityRetention)
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

	_, err := r.db.Exec(ctx, query,
		log.ID, log.EntityType, log.EntityID, log.Action, log.ActorID, log.ActorIP, log.Changes, log.CreatedAt,
	)
	return err
}
