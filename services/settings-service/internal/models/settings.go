package models

import (
	"time"

	"github.com/google/uuid"
)

// PlatformSettings represents global platform configuration
type PlatformSettings struct {
	ID              uuid.UUID `json:"id" db:"id"`
	SystemName      string    `json:"system_name" db:"system_name"`
	SupportEmail    string    `json:"support_email" db:"support_email"`
	DefaultTimezone string    `json:"default_timezone" db:"default_timezone"`
	DefaultCurrency string    `json:"default_currency" db:"default_currency"`
	DefaultLanguage string    `json:"default_language" db:"default_language"`
	WhiteLabel      bool      `json:"white_label_enabled" db:"white_label_enabled"`
	SMTPConfigID    uuid.UUID `json:"smtp_config_id,omitempty" db:"smtp_config_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// FeatureFlag represents a feature toggle
type FeatureFlag struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Description string    `json:"description" db:"description"`
	Scope       string    `json:"scope" db:"scope"` // GLOBAL, ORG
	Enabled     bool      `json:"enabled" db:"enabled"`
	Tags        []string  `json:"tags" db:"tags"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// LocalizationSettings represents language configuration
type LocalizationSettings struct {
	ID           uuid.UUID `json:"id" db:"id"`
	LanguageCode string    `json:"language_code" db:"language_code"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	IsDefault    bool      `json:"is_default" db:"is_default"`
}

// Template represents communication templates (Email/SMS)
type Template struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"template_key"`
	Channel     string    `json:"channel" db:"channel"` // EMAIL, SMS
	Subject     string    `json:"subject" db:"subject"`
	Body        string    `json:"body" db:"body"`
	Variables   []string  `json:"variables" db:"variables"`
	UpdatedBy   uuid.UUID `json:"updated_by" db:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationChannel represents external provider config
type NotificationChannel struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	ChannelType string         `json:"channel_type" db:"channel_type"` // EMAIL, SMS
	Provider    string         `json:"provider" db:"provider"`         // SendGrid, Twilio
	Config      map[string]any `json:"config" db:"config"`             // Encrypted in DB
	Enabled     bool           `json:"enabled" db:"enabled"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// SystemLimit represents global rate limits and retention policies
type SystemLimit struct {
	ID                uuid.UUID `json:"id" db:"id"`
	APIRateLimit      int       `json:"api_rate_limit" db:"api_rate_limit"`
	BurstLimit        int       `json:"burst_limit" db:"burst_limit"`
	AuditRetention    string    `json:"audit_retention" db:"audit_retention"`       // e.g. "1y"
	ActivityRetention string    `json:"activity_retention" db:"activity_retention"` // e.g. "180d"
}

// AuditLog represents a change record
type AuditLog struct {
	ID         uuid.UUID      `json:"id" db:"id"`
	EntityType string         `json:"entity_type" db:"entity_type"`
	EntityID   *uuid.UUID     `json:"entity_id" db:"entity_id"`
	Action     string         `json:"action" db:"action"`
	ActorID    uuid.UUID      `json:"actor_id" db:"actor_id"`
	ActorIP    string         `json:"actor_ip" db:"actor_ip"`
	Changes    map[string]any `json:"changes" db:"changes"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}
