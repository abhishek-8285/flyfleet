package models

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	StatusPaid    InvoiceStatus = "PAID"
	StatusPending InvoiceStatus = "PENDING"
	StatusOverdue InvoiceStatus = "OVERDUE"
	StatusPartial InvoiceStatus = "PARTIAL"
)

type Invoice struct {
	ID            uuid.UUID     `json:"id"`
	OrgID         uuid.UUID     `json:"org_id"`
	CustomerName  string        `json:"customer_name"`
	InvoiceDate   time.Time     `json:"invoice_date"`
	DueDate       time.Time     `json:"due_date"`
	Status        InvoiceStatus `json:"status"`
	Currency      string        `json:"currency"`
	TotalAmount   float64       `json:"total_amount"`
	AmountPaid    float64       `json:"amount_paid"`
	BalanceDue    float64       `json:"balance_due"`
	Items         []InvoiceItem `json:"items,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	CreatedBy     uuid.UUID     `json:"created_by"`
	UpdatedAt     time.Time     `json:"updated_at"`
	UpdatedBy     uuid.UUID     `json:"updated_by"`
}

type InvoiceItem struct {
	ID          uuid.UUID `json:"id"`
	InvoiceID   uuid.UUID `json:"invoice_id"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TaxPercent  float64   `json:"tax_percent"`
	Total       float64   `json:"total"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BillingSettings struct {
	OrgID                 uuid.UUID `json:"org_id"`
	DefaultTaxRate        float64   `json:"default_tax_rate"`
	BaseCurrency          string    `json:"base_currency"`
	GracePeriodDays       int       `json:"grace_period_days"`
	AutoInvoicingEnabled  bool      `json:"auto_invoicing_enabled"`
	CreatedAt             time.Time `json:"created_at"`
	CreatedBy             uuid.UUID `json:"created_by"`
	UpdatedAt             time.Time `json:"updated_at"`
	UpdatedBy             uuid.UUID `json:"updated_by"`
}

type BillingMetrics struct {
	TotalInvoicedMTD float64 `json:"total_invoiced_mtd"`
	PendingPayment   float64 `json:"pending_payment"`
	OverdueAmount    float64 `json:"overdue_amount"`
	PaidCount        int     `json:"paid_count"`
	OpenCount        int     `json:"open_count"`
}
