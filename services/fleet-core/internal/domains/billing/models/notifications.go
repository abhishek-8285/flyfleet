package models

import (
	"time"

	"github.com/google/uuid"
)

// PaymentAttemptStatus represents the status of a single payment retry/attempt
type PaymentAttemptStatus string

const (
	AttemptStatusSuccess    PaymentAttemptStatus = "SUCCESS"
	AttemptStatusFailed     PaymentAttemptStatus = "FAILED"
	AttemptStatusProcessing PaymentAttemptStatus = "PROCESSING"
)

// PaymentAttempt tracks a single dunning retry
type PaymentAttempt struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	OrgID             uuid.UUID              `json:"org_id" db:"org_id"`
	InvoiceID         uuid.UUID              `json:"invoice_id" db:"invoice_id"`
	AttemptNumber     int                    `json:"attempt_number" db:"attempt_number"`
	Status            PaymentAttemptStatus   `json:"status" db:"status"`
	FailureReason     string                 `json:"failure_reason,omitempty" db:"failure_reason"`
	ProcessorResponse map[string]interface{} `json:"processor_response,omitempty" db:"processor_response"`
	Amount            float64                `json:"amount" db:"amount"`
	Currency          string                 `json:"currency" db:"currency"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
}

// NotificationEvent defines the payload for billing-related NATS events
type NotificationEvent struct {
	Type      string                 `json:"type"` // invoice_created, payment_failed, payment_success, subscription_cancelled
	OrgID     uuid.UUID              `json:"org_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
