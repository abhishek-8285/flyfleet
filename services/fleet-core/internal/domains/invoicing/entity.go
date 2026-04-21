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

type TemplateType string

const (
	TemplateTaxInvoice TemplateType = "TAX_INVOICE"
	TemplateTripNote   TemplateType = "TRIP_NOTE"
)

type Invoice struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	TenantID           uuid.UUID     `json:"tenant_id" db:"tenant_id"`
	InvoiceNumber   string        `json:"invoice_number" db:"invoice_number"`
	CustomerName    string        `json:"customer_name" db:"customer_name"`
	CustomerEmail   string        `json:"customer_email" db:"customer_email"`
	CustomerPhone   string        `json:"customer_phone" db:"customer_phone"`
	BillingAddress  string        `json:"billing_address" db:"billing_address"`
	DueDate         time.Time     `json:"due_date" db:"due_date"`
	InvoiceDate     time.Time     `json:"invoice_date" db:"invoice_date"`
	Status          InvoiceStatus `json:"status" db:"status"`
	Subtotal        float64       `json:"subtotal" db:"subtotal"`
	TaxAmount       float64       `json:"tax_amount" db:"tax_amount"`
	TotalAmount     float64       `json:"total_amount" db:"total_amount"`
	AmountPaid      float64       `json:"amount_paid" db:"amount_paid"`
	AmountRemaining float64       `json:"amount_remaining" db:"amount_remaining"`
	// Template control
	TemplateType TemplateType `json:"template_type" db:"template_type"`

	// Logistics / Trip info
	TripID         *uuid.UUID `json:"trip_id,omitempty" db:"trip_id"`
	VehicleNo      string     `json:"vehicle_no,omitempty" db:"vehicle_no"`
	DriverName     string     `json:"driver_name,omitempty" db:"driver_name"`
	DriverMobile   string     `json:"driver_mobile,omitempty" db:"driver_mobile"`
	PickupLocation string     `json:"pickup_location,omitempty" db:"pickup_location"`
	DropLocation   string     `json:"drop_location,omitempty" db:"drop_location"`
	PickupTime     *time.Time `json:"pickup_time,omitempty" db:"pickup_time"`
	DropTime       *time.Time `json:"drop_time,omitempty" db:"drop_time"`
	MapUrl         string     `json:"map_url,omitempty" db:"map_url"`
	BilledDistance float64    `json:"billed_distance,omitempty" db:"billed_distance"`

	// Compliance
	PlaceOfSupply string    `json:"place_of_supply,omitempty" db:"place_of_supply"`
	ReverseCharge bool      `json:"reverse_charge" db:"reverse_charge"`
	GstinSeller   string    `json:"gstin_seller,omitempty" db:"gstin_seller"`
	PanSeller     string    `json:"pan_seller,omitempty" db:"pan_seller"`
	EWayBillNo    string    `json:"eway_bill_no,omitempty" db:"eway_bill_no"`
	EWayBillDate  *time.Time `json:"eway_bill_date,omitempty" db:"eway_bill_date"`
	IRN           string    `json:"irn,omitempty" db:"irn"`
	AckNo         string    `json:"ack_no,omitempty" db:"ack_no"`
	AckDate       *time.Time `json:"ack_date,omitempty" db:"ack_date"`

	// Financials / Bank
	BankName     string `json:"bank_name,omitempty" db:"bank_name"`
	AccountNo    string `json:"account_no,omitempty" db:"account_no"`
	IFSC         string `json:"ifsc_code,omitempty" db:"ifsc_code"`
	Branch       string `json:"branch_name,omitempty" db:"branch_name"`
	Beneficiary  string `json:"beneficiary_name,omitempty" db:"beneficiary_name"`

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
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Description string    `json:"description" db:"description"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	TaxPercent  float64   `json:"tax_percent" db:"tax_percent"`
	Amount      float64   `json:"amount" db:"amount"`

	// New fields
	HSNSAC         string  `json:"hsn_sac,omitempty" db:"hsn_sac"`
	Unit           string  `json:"unit,omitempty" db:"unit"`
	ListPrice      float64 `json:"list_price,omitempty" db:"list_price"`
	DiscountAmount float64 `json:"discount_amount,omitempty" db:"discount_amount"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
