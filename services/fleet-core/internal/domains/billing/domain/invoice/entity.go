package invoice

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	StatusDraft         InvoiceStatus = "draft"
	StatusOpen          InvoiceStatus = "open"
	StatusPaid          InvoiceStatus = "paid"
	StatusVoid          InvoiceStatus = "void"
	StatusUncollectible InvoiceStatus = "uncollectible"
)

type Invoice struct {
	ID              uuid.UUID     `json:"id"`
	OrgID           uuid.UUID     `json:"org_id"`
	InvoiceNumber   string        `json:"invoice_number"`
	CustomerName    string        `json:"customer_name" validate:"required"`
	CustomerEmail   string        `json:"customer_email" validate:"email"`
	BillingAddress  string        `json:"billing_address"`
	DueDate         time.Time     `json:"due_date" validate:"required"`
	Status          InvoiceStatus `json:"status"`
	Subtotal        float64       `json:"subtotal"`
	TaxAmount       float64       `json:"tax_amount"`
	TotalAmount     float64       `json:"total_amount"`
	AmountPaid      float64       `json:"amount_paid"`
	AmountRemaining float64       `json:"amount_remaining"`
	Currency        string        `json:"currency"`
	Notes           string        `json:"notes"`
	Items           []InvoiceItem `json:"items,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	CreatedBy       uuid.UUID     `json:"created_by"`
	UpdatedAt       time.Time     `json:"updated_at"`
	UpdatedBy       uuid.UUID     `json:"updated_by"`
	IsDeleted       bool          `json:"is_deleted"`
	DeletedAt       *time.Time    `json:"deleted_at,omitempty"`
}

type InvoiceItem struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	InvoiceID   uuid.UUID `json:"invoice_id"`
	Description string    `json:"description" validate:"required"`
	Quantity    int       `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64   `json:"unit_price" validate:"required,gte=0"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

type Repository interface {
	Create(ctx context.Context, inv *Invoice) (*Invoice, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*Invoice, error)
	List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Invoice, int64, error)
	UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status InvoiceStatus, updatedBy uuid.UUID) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type Service interface {
	CreateInvoice(ctx context.Context, orgID uuid.UUID, inv *Invoice) (*Invoice, error)
	GetInvoice(ctx context.Context, orgID, id uuid.UUID) (*Invoice, error)
	ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Invoice, int64, error)
	VoidInvoice(ctx context.Context, orgID, id uuid.UUID) error
}
