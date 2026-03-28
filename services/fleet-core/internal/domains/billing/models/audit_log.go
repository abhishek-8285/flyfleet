package models

import (
	"time"

	"github.com/google/uuid"
)

// InvoiceAuditLog represents an audit log entry for invoice changes
type InvoiceAuditLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	InvoiceID      uuid.UUID `json:"invoice_id" db:"invoice_id"`
	Action         string    `json:"action" db:"action"` // created, updated, issued, cancelled, payment_recorded
	OldValue       []byte    `json:"old_value,omitempty" db:"old_value"`
	NewValue       []byte    `json:"new_value,omitempty" db:"new_value"`
	ChangedBy      uuid.UUID `json:"changed_by,omitempty" db:"changed_by"`
	ChangedByEmail string    `json:"changed_by_email,omitempty" db:"changed_by_email"`
	ChangedAt      time.Time `json:"changed_at" db:"changed_at"`
	IPAddress      string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      string    `json:"user_agent,omitempty" db:"user_agent"`
	Reason         string    `json:"reason,omitempty" db:"reason"`
}

// AuditLogAction defines valid audit log actions
type AuditLogAction string

const (
	AuditActionCreated          AuditLogAction = "created"
	AuditActionUpdated          AuditLogAction = "updated"
	AuditActionIssued           AuditLogAction = "issued"
	AuditActionCancelled        AuditLogAction = "cancelled"
	AuditActionPaymentRecorded  AuditLogAction = "payment_recorded"
	AuditActionCreditNoteIssued AuditLogAction = "credit_note_issued"
	AuditActionApplied          AuditLogAction = "applied"
	AuditActionSent             AuditLogAction = "sent"
	AuditActionViewed           AuditLogAction = "viewed"
	AuditActionDownloaded       AuditLogAction = "downloaded"
)

// AuditLogSummary represents a summary of audit logs for an invoice
type AuditLogSummary struct {
	InvoiceID        uuid.UUID `json:"invoice_id"`
	TotalChanges     int64     `json:"total_changes"`
	LastChangedAt    time.Time `json:"last_changed_at"`
	ActionsPerformed []string  `json:"actions_performed"`
}
