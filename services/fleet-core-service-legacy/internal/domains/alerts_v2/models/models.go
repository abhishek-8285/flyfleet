package models

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelEmail   Channel = "EMAIL"
	ChannelSMS     Channel = "SMS"
	ChannelPush    Channel = "PUSH"
	ChannelInApp   Channel = "IN_APP"
	ChannelWebhook Channel = "WEBHOOK"
)

type NotificationStatus string

const (
	StatusQueued   NotificationStatus = "QUEUED"
	StatusSending  NotificationStatus = "SENDING"
	StatusSent     NotificationStatus = "SENT"
	StatusFailed   NotificationStatus = "FAILED"
	StatusDLQ      NotificationStatus = "DLQ"
)

type AlertPriority string

const (
	PriorityCritical AlertPriority = "CRITICAL"
	PriorityHigh     AlertPriority = "HIGH"
	PriorityNormal   AlertPriority = "NORMAL"
	PriorityLow      AlertPriority = "LOW"
)

type Template struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Channel   Channel        `json:"channel"`
	Language  string         `json:"language"`
	Subject   string         `json:"subject,omitempty"`
	Body      string         `json:"body"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type UserPreference struct {
	UserID    uuid.UUID `json:"user_id"`
	Channel   Channel   `json:"channel"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationLog struct {
	ID               uuid.UUID           `json:"id"`
	UserID           *uuid.UUID          `json:"user_id,omitempty"`
	Recipient        string              `json:"recipient"`
	Channel          Channel             `json:"channel"`
	AlertType        string              `json:"alert_type"`
	Subject          string              `json:"subject,omitempty"`
	Body             string              `json:"body"`
	Status           NotificationStatus  `json:"status"`
	RetryCount       int                 `json:"retry_count"`
	MaxRetries       int                 `json:"max_retries"`
	Provider         string              `json:"provider,omitempty"`
	ProviderResponse map[string]any      `json:"provider_response,omitempty"`
	ErrorMessage     string              `json:"error_message,omitempty"`
	Metadata         map[string]any      `json:"metadata,omitempty"`
	ScheduledAt      time.Time           `json:"scheduled_at"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type DLQEntry struct {
	ID             uuid.UUID      `json:"id"`
	NotificationID uuid.UUID      `json:"notification_id"`
	Reason         string         `json:"reason"`
	Payload        map[string]any `json:"payload"`
	CreatedAt      time.Time      `json:"created_at"`
}

type AlertRequest struct {
	UserID    *uuid.UUID          `json:"user_id,omitempty"`
	Recipient string              `json:"recipient,omitempty"` // For non-user alerts
	AlertType string              `json:"alert_type"`
	Priority  AlertPriority       `json:"priority"`
	Data      map[string]any      `json:"data"`
	Metadata  map[string]any      `json:"metadata,omitempty"`
}
