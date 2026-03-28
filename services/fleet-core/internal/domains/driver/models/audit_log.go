package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverAuditLog represents an audit log entry for driver operations
type DriverAuditLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	DriverID  uuid.UUID              `json:"driver_id" db:"driver_id"`
	OrgID     uuid.UUID              `json:"org_id" db:"org_id"`
	Action    string                 `json:"action" db:"action"` // created, updated, deleted, status_changed, vehicle_assigned, document_uploaded, etc.
	OldValue  map[string]interface{} `json:"old_value,omitempty" db:"old_value"`
	NewValue  map[string]interface{} `json:"new_value,omitempty" db:"new_value"`
	UserID    uuid.UUID              `json:"user_id" db:"user_id"`
	UserType  string                 `json:"user_type" db:"user_type"` // admin, system, driver
	UserEmail string                 `json:"user_email" db:"user_email"`
	Reason    string                 `json:"reason,omitempty" db:"reason"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IPAddress string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent string                 `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// AuditAction defines valid audit log actions
type AuditAction string

const (
	ActionDriverCreated              AuditAction = "driver.created"
	ActionDriverUpdated              AuditAction = "driver.updated"
	ActionDriverDeleted              AuditAction = "driver.deleted"
	ActionDriverStatusChanged        AuditAction = "driver.status_changed"
	ActionDriverVehicleAssigned      AuditAction = "driver.vehicle_assigned"
	ActionDriverVehicleUnassigned    AuditAction = "driver.vehicle_unassigned"
	ActionDriverDocumentUploaded     AuditAction = "driver.document_uploaded"
	ActionDriverDocumentVerified     AuditAction = "driver.document_verified"
	ActionDriverDocumentExpired      AuditAction = "driver.document_expired"
	ActionDriverBehaviorScoreUpdated AuditAction = "driver.behavior_score_updated"
	ActionDriverIncidentCreated      AuditAction = "driver.incident_created"
	ActionDriverIncidentResolved     AuditAction = "driver.incident_resolved"
	ActionDriverDutyStatusChanged    AuditAction = "driver.duty_status_changed"
)

// AuditLogFilter defines filters for audit log queries
type AuditLogFilter struct {
	DriverID *uuid.UUID `json:"driver_id,omitempty"`
	OrgID    uuid.UUID  `json:"org_id"`
	Action   string     `json:"action,omitempty"`
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	UserType string     `json:"user_type,omitempty"`
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
	Page     int        `json:"page" validate:"gte=1"`
	PageSize int        `json:"page_size" validate:"gte=1,lte=100"`
}

// DriverChange represents a driver change for audit purposes
type DriverChange struct {
	Field     string      `json:"field"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	ChangedAt time.Time   `json:"changed_at"`
	ChangedBy uuid.UUID   `json:"changed_by"`
}

// DocumentAudit represents document audit information
type DocumentAudit struct {
	DocumentID   uuid.UUID  `json:"document_id"`
	DocumentType string     `json:"document_type"`
	Action       string     `json:"action"` // uploaded, verified, expired
	VerifiedBy   *uuid.UUID `json:"verified_by,omitempty"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	ExpiryDate   time.Time  `json:"expiry_date,omitempty"`
	AuditedAt    time.Time  `json:"audited_at"`
}

// BehaviorScoreAudit represents behavior score change audit
type BehaviorScoreAudit struct {
	DriverID    uuid.UUID `json:"driver_id"`
	OldScore    int       `json:"old_score"`
	NewScore    int       `json:"new_score"`
	ScoreChange int       `json:"score_change"`
	Reason      string    `json:"reason"`
	EventType   string    `json:"event_type"`
	ChangedBy   uuid.UUID `json:"changed_by"`
	ChangedAt   time.Time `json:"changed_at"`
}
