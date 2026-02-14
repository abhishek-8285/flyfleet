package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationChannel string

const (
	ChannelEmail  NotificationChannel = "EMAIL"
	ChannelSMS    NotificationChannel = "SMS"
	ChannelPush   NotificationChannel = "PUSH"
	ChannelInApp  NotificationChannel = "IN_APP"
)

type NotificationStatus string

const (
	StatusQueued NotificationStatus = "QUEUED"
	StatusSent   NotificationStatus = "SENT"
	StatusFailed NotificationStatus = "FAILED"
	StatusRead   NotificationStatus = "READ"
)

type NotificationLog struct {
	ID               uuid.UUID           `json:"id"`
	CorrelationID    *string             `json:"correlation_id"`
	UserID           *uuid.UUID          `json:"user_id"`
	UserEmail        string              `json:"user_email"`
	Channel          NotificationChannel `json:"channel"`
	Type             string              `json:"type"`
	Subject          string              `json:"subject"`
	Body             string              `json:"body"`
	Status           NotificationStatus  `json:"status"`
	Provider         *string             `json:"provider"`
	ProviderResponse *string             `json:"provider_response"`
	RetryCount       int                 `json:"retry_count"`
	MaxRetries       int                 `json:"max_retries"`
	Metadata         map[string]any      `json:"metadata"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Items            []NotificationEvent `json:"items,omitempty"` // For detail view
}

type NotificationEvent struct {
	ID             uuid.UUID `json:"id"`
	NotificationID uuid.UUID `json:"notification_id"`
	EventType      string    `json:"event_type"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}

type NotificationSettings struct {
	ID              uuid.UUID           `json:"id"`
	Channel         NotificationChannel `json:"channel" binding:"required,oneof=EMAIL SMS PUSH IN_APP"`
	Enabled         bool                `json:"enabled"`
	MaxRetries      int                 `json:"max_retries" binding:"min=0,max=10"`
	Priority        string              `json:"priority" binding:"oneof=HIGH STANDARD LOW"`
	RetentionMonths int                 `json:"retention_months" binding:"min=1,max=24"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type NotificationTemplate struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Channel   NotificationChannel `json:"channel"`
	Language  string    `json:"language"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	UpdatedAt time.Time `json:"updated_at"`
}
type UserNotificationPreference struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	UserName       string    `json:"user_name,omitempty"`  // Joined field
	UserEmail      string    `json:"user_email,omitempty"` // Joined field
	EmailEnabled   bool      `json:"email_enabled"`
	SmsEnabled     bool      `json:"sms_enabled"`
	PushEnabled    bool      `json:"push_enabled"`
	InAppEnabled   bool      `json:"in_app_enabled"`
	LastModifiedBy uuid.UUID `json:"last_modified_by"`
	LastModifiedAt time.Time `json:"last_modified_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PreferenceAudit struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ChangedBy    uuid.UUID `json:"changed_by"`
	OldValues    any       `json:"old_values"`
	NewValues    any       `json:"new_values"`
	ChangeReason string    `json:"change_reason"`
	CreatedAt    time.Time `json:"created_at"`
}

type GlobalNotificationDefault struct {
	ID           int       `json:"id"`
	EmailEnabled bool      `json:"email_enabled"`
	SmsEnabled   bool      `json:"sms_enabled"`
	PushEnabled  bool      `json:"push_enabled"`
	InAppEnabled bool      `json:"in_app_enabled"`
	UpdatedBy    uuid.UUID `json:"updated_by"`
	UpdatedAt    time.Time `json:"updated_at"`
}
