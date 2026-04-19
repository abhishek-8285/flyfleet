package models

import (
	"time"

	"github.com/google/uuid"
)

type MFAEvent struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	EventType string    `json:"event_type"`
	Method    string    `json:"method,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type MFAPolicy struct {
	ID                 uuid.UUID `json:"id"`
	OrgID              uuid.UUID `json:"tenant_id"`
	RequireMFA         bool      `json:"require_mfa"`
	AllowedMethods     []string  `json:"allowed_methods"`
	RememberDeviceDays int       `json:"remember_device_days"`
	EnforcedForRoles   []string  `json:"enforced_for_roles"`
	UpdatedAt          time.Time `json:"updated_at"`
}
