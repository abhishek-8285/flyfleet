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
	ID            uuid.UUID     `json:"id" db:"id"`
	OrgID         uuid.UUID     `json:"org_id" db:"org_id" validate:"required"`
	InvoiceNumber  string        `json:"invoice_number" db:"invoice_number"`
	CustomerName   string        `json:"customer_name" db:"customer_name" validate:"required"`
	CustomerEmail  string        `json:"customer_email" db:"customer_email"`
	BillingAddress string        `json:"billing_address" db:"billing_address"`
	CustomerID     *uuid.UUID    `json:"customer_id,omitempty" db:"customer_id"`
	ShipmentID    *uuid.UUID    `json:"shipment_id,omitempty" db:"shipment_id"`
	InvoiceDate   time.Time     `json:"invoice_date" db:"invoice_date" validate:"required"`
	DueDate       time.Time     `json:"due_date" db:"due_date" validate:"required"`
	Status        InvoiceStatus `json:"status" db:"status" validate:"required,oneof=PAID PENDING OVERDUE PARTIAL"`
	Currency      string        `json:"currency" db:"currency" validate:"required,len=3"`
	TotalAmount   float64       `json:"total_amount" db:"total_amount" validate:"gte=0"`
	AmountPaid    float64       `json:"amount_paid" db:"amount_paid" validate:"gte=0"`
	BalanceDue     float64       `json:"balance_due" db:"balance_due" validate:"gte=0"`
	Subtotal       float64       `json:"subtotal" db:"subtotal"`
	DiscountType   string        `json:"discount_type" db:"discount_type"`
	DiscountValue  float64       `json:"discount_value" db:"discount_value"`
	DiscountAmount float64       `json:"discount_amount" db:"discount_amount"`
	Notes          string        `json:"notes,omitempty" db:"notes"`
	Category       string        `json:"category" db:"category"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CustomFields  map[string]interface{} `json:"custom_fields,omitempty" db:"custom_fields"`
	Items         []InvoiceItem `json:"items,omitempty" validate:"dive"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	CreatedBy     uuid.UUID     `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	UpdatedBy     uuid.UUID     `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt     *time.Time    `json:"deleted_at,omitempty" db:"deleted_at"`
}

type InvoiceItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	InvoiceID   uuid.UUID `json:"invoice_id" db:"invoice_id" validate:"required"`
	Description string    `json:"description" db:"description" validate:"required"`
	Quantity    int       `json:"quantity" db:"quantity" validate:"required,gt=0"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price" validate:"required,gte=0"`
	TaxPercent  float64   `json:"tax_percent" db:"tax_percent" validate:"gte=0,lte=100"`
	Total       float64   `json:"total" db:"total" validate:"gte=0"`
	HSNCode     string    `json:"hsn_code" db:"hsn_code"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CustomFields  map[string]interface{} `json:"custom_fields,omitempty" db:"custom_fields"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type BillingSettings struct {
	OrgID                uuid.UUID `json:"org_id" db:"org_id"`
	DefaultTaxRate       float64   `json:"default_tax_rate" db:"default_tax_rate"`
	BaseCurrency         string    `json:"base_currency" db:"base_currency"`
	GracePeriodDays      int       `json:"grace_period_days" db:"grace_period_days"`
	AutoInvoicingEnabled bool      `json:"auto_invoicing_enabled" db:"auto_invoicing_enabled"`
	TaxCountry           string    `json:"tax_country" db:"tax_country"`
	TaxState             *string   `json:"tax_state,omitempty" db:"tax_state"`
	InvoicePrefix        string    `json:"invoice_prefix" db:"invoice_prefix"`
	NextInvoiceNumber    int       `json:"next_invoice_number" db:"next_invoice_number"`
	RenderConfig         []byte    `json:"render_config" db:"render_config"`
	
	// Dynamic Invoicing Details
	CompanyName          string    `json:"company_name" db:"company_name"`
	CompanyAddress       string    `json:"company_address" db:"company_address"`
	CompanyGSTIN         string    `json:"company_gstin" db:"company_gstin"`
	CompanyPAN           string    `json:"company_pan" db:"company_pan"`
	BankName             string    `json:"bank_name" db:"bank_name"`
	AccountName          string    `json:"account_name" db:"account_name"`
	AccountNumber        string    `json:"account_number" db:"account_number"`
	BankIFSC             string    `json:"bank_ifsc" db:"bank_ifsc"`
	Terms                string    `json:"terms" db:"terms"`
	Declaration          string    `json:"declaration" db:"declaration"`

	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	CreatedBy            uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy            uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

type BillingMetrics struct {
	TotalInvoicedMTD float64 `json:"total_invoiced_mtd"`
	PendingPayment   float64 `json:"pending_payment"`
	OverdueAmount    float64 `json:"overdue_amount"`
	PaidCount        int     `json:"paid_count"`
	OpenCount        int     `json:"open_count"`
}

type PaymentRecord struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	InvoiceID      uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	Amount         float64    `json:"amount" db:"amount"`
	PaymentMethod  string     `json:"payment_method" db:"payment_method"`
	PaymentReference string   `json:"payment_reference" db:"payment_reference"`
	PaymentDate    time.Time  `json:"payment_date" db:"payment_date"`
	Notes          string     `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	CreatedBy      uuid.UUID  `json:"created_by,omitempty" db:"created_by"`
}

type InvoiceTemplate struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Version     int       `json:"version" db:"version"`
	HTMLContent string    `json:"html_content" db:"html_content"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type InvoiceDocument struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	InvoiceID       uuid.UUID              `json:"invoice_id" db:"invoice_id"`
	TemplateID      uuid.UUID              `json:"template_id" db:"template_id"`
	TemplateVersion int                    `json:"template_version" db:"template_version"`
	Snapshot        map[string]interface{} `json:"snapshot" db:"snapshot"`
	FileURL         string                 `json:"file_url" db:"file_url"`
	StorageKey      string                 `json:"storage_key" db:"storage_key"`
	Checksum        string                 `json:"checksum" db:"checksum"`
	GeneratedAt     time.Time              `json:"generated_at" db:"generated_at"`
}
