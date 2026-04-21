package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookSubscription represents a webhook subscription
type WebhookSubscription struct {
	ID                   uuid.UUID              `json:"id" db:"id"`
	TenantID                uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name                 string                 `json:"name" db:"name"`
	URL                  string                 `json:"url" db:"url"`
	Secret               string                 `json:"-" db:"secret"` // Never expose in JSON
	IsActive             bool                   `json:"is_active" db:"is_active"`
	Events               []string               `json:"events" db:"events"`
	CustomHeaders        map[string]string      `json:"custom_headers,omitempty" db:"custom_headers"`
	MaxRetries           int                    `json:"max_retries" db:"max_retries"`
	RetryDelaySeconds    int                    `json:"retry_delay_seconds" db:"retry_delay_seconds"`
	TimeoutSeconds       int                    `json:"timeout_seconds" db:"timeout_seconds"`
	RateLimitPerMinute   int                    `json:"rate_limit_per_minute" db:"rate_limit_per_minute"`
	TotalDeliveries      int                    `json:"total_deliveries" db:"total_deliveries"`
	SuccessfulDeliveries int                    `json:"successful_deliveries" db:"successful_deliveries"`
	FailedDeliveries     int                    `json:"failed_deliveries" db:"failed_deliveries"`
	LastDeliveryAt       *time.Time             `json:"last_delivery_at,omitempty" db:"last_delivery_at"`
	LastSuccessAt        *time.Time             `json:"last_success_at,omitempty" db:"last_success_at"`
	LastFailureAt        *time.Time             `json:"last_failure_at,omitempty" db:"last_failure_at"`
	Description          string                 `json:"description,omitempty" db:"description"`
	Metadata             map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	CreatedBy            uuid.UUID              `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`
	UpdatedBy            uuid.UUID              `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt            *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID              uuid.UUID         `json:"id" db:"id"`
	SubscriptionID  uuid.UUID         `json:"subscription_id" db:"subscription_id"`
	EventType       string            `json:"event_type" db:"event_type"`
	EventID         uuid.UUID         `json:"event_id" db:"event_id"`
	Payload         interface{}       `json:"payload" db:"payload"`
	Attempt         int               `json:"attempt" db:"attempt"`
	Status          string            `json:"status" db:"status"` // pending, success, failed, retrying
	HTTPStatus      *int              `json:"http_status,omitempty" db:"http_status"`
	ResponseBody    string            `json:"response_body,omitempty" db:"response_body"`
	ErrorMessage    string            `json:"error_message,omitempty" db:"error_message"`
	SentAt          time.Time         `json:"sent_at" db:"sent_at"`
	DeliveredAt     *time.Time        `json:"delivered_at,omitempty" db:"delivered_at"`
	NextRetryAt     *time.Time        `json:"next_retry_at,omitempty" db:"next_retry_at"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty" db:"request_headers"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" db:"response_headers"`
	DurationMs      *int              `json:"duration_ms,omitempty" db:"duration_ms"`
	Signature       string            `json:"signature" db:"signature"`
}

// WebhookEventType represents an available webhook event type
type WebhookEventType struct {
	EventType   string `json:"event_type" db:"event_type"`
	Description string `json:"description" db:"description"`
	Category    string `json:"category" db:"category"`
	IsActive    bool   `json:"is_active" db:"is_active"`
}

// WebhookStats represents webhook delivery statistics
type WebhookStats struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID                uuid.UUID  `json:"tenant_id"`
	Name                 string     `json:"name"`
	URL                  string     `json:"url"`
	IsActive             bool       `json:"is_active"`
	TotalDeliveries      int        `json:"total_deliveries"`
	SuccessfulDeliveries int        `json:"successful_deliveries"`
	FailedDeliveries     int        `json:"failed_deliveries"`
	SuccessRate          float64    `json:"success_rate"`
	LastDeliveryAt       *time.Time `json:"last_delivery_at,omitempty"`
	LastSuccessAt        *time.Time `json:"last_success_at,omitempty"`
	LastFailureAt        *time.Time `json:"last_failure_at,omitempty"`
}

// Standard webhook event types
const (
	EventInvoiceCreated    = "invoice.created"
	EventInvoiceUpdated    = "invoice.updated"
	EventInvoiceIssued     = "invoice.issued"
	EventInvoiceCancelled  = "invoice.cancelled"
	EventInvoicePaid       = "invoice.paid"
	EventInvoiceOverdue    = "invoice.overdue"
	EventPaymentRecorded   = "payment.recorded"
	EventPaymentRefunded   = "payment.refunded"
	EventCreditNoteCreated = "credit_note.created"
	EventCreditNoteIssued  = "credit_note.issued"
	EventCreditNoteApplied = "credit_note.applied"
	EventApprovalRequested = "approval.requested"
	EventApprovalApproved  = "approval.approved"
	EventApprovalRejected  = "approval.rejected"
)

// StandardEventTypes returns list of all standard webhook event types
func StandardEventTypes() []WebhookEventType {
	return []WebhookEventType{
		{EventType: EventInvoiceCreated, Description: "Triggered when a new invoice is created", Category: "invoice", IsActive: true},
		{EventType: EventInvoiceUpdated, Description: "Triggered when an invoice is updated", Category: "invoice", IsActive: true},
		{EventType: EventInvoiceIssued, Description: "Triggered when an invoice is issued", Category: "invoice", IsActive: true},
		{EventType: EventInvoiceCancelled, Description: "Triggered when an invoice is cancelled", Category: "invoice", IsActive: true},
		{EventType: EventInvoicePaid, Description: "Triggered when an invoice is fully paid", Category: "invoice", IsActive: true},
		{EventType: EventInvoiceOverdue, Description: "Triggered when an invoice becomes overdue", Category: "invoice", IsActive: true},
		{EventType: EventPaymentRecorded, Description: "Triggered when a payment is recorded", Category: "payment", IsActive: true},
		{EventType: EventPaymentRefunded, Description: "Triggered when a payment is refunded", Category: "payment", IsActive: true},
		{EventType: EventCreditNoteCreated, Description: "Triggered when a credit note is created", Category: "credit_note", IsActive: true},
		{EventType: EventCreditNoteIssued, Description: "Triggered when a credit note is issued", Category: "credit_note", IsActive: true},
		{EventType: EventCreditNoteApplied, Description: "Triggered when a credit note is applied", Category: "credit_note", IsActive: true},
		{EventType: EventApprovalRequested, Description: "Triggered when an invoice approval is requested", Category: "approval", IsActive: true},
		{EventType: EventApprovalApproved, Description: "Triggered when an invoice is approved", Category: "approval", IsActive: true},
		{EventType: EventApprovalRejected, Description: "Triggered when an invoice is rejected", Category: "approval", IsActive: true},
	}
}

// IsValidEventType checks if an event type is valid
func IsValidEventType(eventType string) bool {
	for _, et := range StandardEventTypes() {
		if et.EventType == eventType && et.IsActive {
			return true
		}
	}
	return false
}

// DefaultWebhookSubscription creates a default webhook subscription
func DefaultWebhookSubscription(tenantID uuid.UUID, name, url string) *WebhookSubscription {
	return &WebhookSubscription{
		ID:                 uuid.New(),
		TenantID:              tenantID,
		Name:               name,
		URL:                url,
		IsActive:           true,
		Events:             []string{EventInvoiceCreated, EventInvoicePaid},
		MaxRetries:         3,
		RetryDelaySeconds:  60,
		TimeoutSeconds:     30,
		RateLimitPerMinute: 60,
		CustomHeaders:      make(map[string]string),
		Metadata:           make(map[string]interface{}),
	}
}
