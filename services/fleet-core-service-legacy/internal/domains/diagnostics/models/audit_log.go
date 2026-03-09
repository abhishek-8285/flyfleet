package models

import (
	"time"

	"github.com/google/uuid"
)

// DiagnosticsAuditLog represents an audit log entry
type DiagnosticsAuditLog struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	OrgID      uuid.UUID              `json:"org_id" db:"org_id"`
	EntityType string                 `json:"entity_type" db:"entity_type"` // device, command, firmware, rollout
	EntityID   uuid.UUID              `json:"entity_id" db:"entity_id"`
	Action     string                 `json:"action" db:"action"` // created, updated, deleted, status_changed, command_sent, command_executed
	OldValue   map[string]interface{} `json:"old_value,omitempty" db:"old_value"`
	NewValue   map[string]interface{} `json:"new_value,omitempty" db:"new_value"`
	UserID     *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	UserType   string                 `json:"user_type" db:"user_type"` // admin, system, device
	UserEmail  *string                `json:"user_email,omitempty" db:"user_email"`
	IPAddress  *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  *string                `json:"user_agent,omitempty" db:"user_agent"`
	Reason     *string                `json:"reason,omitempty" db:"reason"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// AuditLogAction defines valid audit log actions
type AuditLogAction string

const (
	ActionCreated        AuditLogAction = "created"
	ActionUpdated        AuditLogAction = "updated"
	ActionDeleted        AuditLogAction = "deleted"
	ActionStatusChanged  AuditLogAction = "status_changed"
	ActionCommandSent    AuditLogAction = "command_sent"
	ActionCommandExecuted AuditLogAction = "command_executed"
)

// EntityType defines auditable entity types
type EntityType string

const (
	EntityDevice   EntityType = "device"
	EntityCommand  EntityType = "command"
	EntityFirmware EntityType = "firmware"
	EntityRollout  EntityType = "rollout"
)

// AuditLogFilter defines filters for audit log queries
type AuditLogFilter struct {
	OrgID      uuid.UUID
	EntityType string
	EntityID   *uuid.UUID
	Action     string
	UserID     *uuid.UUID
	UserType   string
	FromDate   *time.Time
	ToDate     *time.Time
	Page       int
	PageSize   int
}
