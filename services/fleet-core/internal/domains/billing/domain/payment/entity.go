package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusSucceeded PaymentStatus = "succeeded"
	StatusFailed    PaymentStatus = "failed"
	StatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID            uuid.UUID     `json:"id"`
	OrgID         uuid.UUID     `json:"tenant_id"`
	InvoiceID     *uuid.UUID    `json:"invoice_id,omitempty"`
	Amount        float64       `json:"amount" validate:"required,gt=0"`
	Status        PaymentStatus `json:"status"`
	PaymentMethod string        `json:"payment_method"`
	TransactionID string        `json:"transaction_id"`
	Notes         string        `json:"notes"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	IsDeleted     bool          `json:"is_deleted"`
	DeletedAt     *time.Time    `json:"deleted_at,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, p *Payment) (*Payment, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*Payment, error)
	ListByInvoice(ctx context.Context, orgID, invoiceID uuid.UUID) ([]Payment, error)
}
