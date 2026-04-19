package invoicing

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	StatusDraft     InvoiceStatus = "DRAFT"
	StatusPending   InvoiceStatus = "PENDING"
	StatusPaid      InvoiceStatus = "PAID"
	StatusOverdue   InvoiceStatus = "OVERDUE"
	StatusVoid      InvoiceStatus = "VOID"
	StatusCancelled InvoiceStatus = "CANCELLED"
)

type Invoice struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	OrgID           uuid.UUID     `json:"org_id" db:"tenant_id"`
	InvoiceNumber   string        `json:"invoice_number" db:"invoice_number"`
	CustomerName    string        `json:"customer_name" db:"customer_name"`
	CustomerEmail   string        `json:"customer_email" db:"customer_email"`
	CustomerPhone   string        `json:"customer_phone" db:"customer_phone"`
	BillingAddress  string        `json:"billing_address" db:"billing_address"`
	DueDate         time.Time     `json:"due_date" db:"due_date"`
	Status          InvoiceStatus `json:"status" db:"status"`
	Subtotal        float64       `json:"subtotal" db:"subtotal"`
	TaxAmount       float64       `json:"tax_amount" db:"tax_amount"`
	TotalAmount     float64       `json:"total_amount" db:"total_amount"`
	AmountPaid      float64       `json:"amount_paid" db:"amount_paid"`
	AmountRemaining float64       `json:"amount_remaining" db:"amount_remaining"`
	Currency        string        `json:"currency" db:"currency"`
	Notes           string        `json:"notes" db:"notes"`
	Items           []InvoiceItem `json:"items"`
	CreatedBy       uuid.UUID     `json:"created_by" db:"created_by"`
	UpdatedBy       uuid.UUID     `json:"updated_by" db:"updated_by"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

type InvoiceItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	InvoiceID   uuid.UUID `json:"invoice_id" db:"invoice_id"`
	OrgID       uuid.UUID `json:"org_id" db:"tenant_id"`
	Description string    `json:"description" db:"description"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	Amount      float64   `json:"amount" db:"amount"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
