package models

import (
	"time"

	"github.com/google/uuid"
)

// --- GST E-Invoice ---

type GSTInvoiceStatus string

const (
	GSTInvoiceStatusPending   GSTInvoiceStatus = "PENDING"
	GSTInvoiceStatusGenerated GSTInvoiceStatus = "GENERATED"
	GSTInvoiceStatusCancelled GSTInvoiceStatus = "CANCELLED"
	GSTInvoiceStatusError     GSTInvoiceStatus = "ERROR"
)

type GSTInvoice struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	InvoiceNumber string           `json:"invoice_number" db:"invoice_number"`
	IRN           *string          `json:"irn,omitempty" db:"irn"`
	AckNumber     *string          `json:"ack_number,omitempty" db:"ack_number"`
	AckDate       *time.Time       `json:"ack_date,omitempty" db:"ack_date"`
	QRCode        *string          `json:"qr_code,omitempty" db:"qr_code"`
	SignedInvoice *string          `json:"signed_invoice,omitempty" db:"signed_invoice"`
	Status        GSTInvoiceStatus `json:"status" db:"status"`
	EntityType    string           `json:"entity_type" db:"entity_type"`
	EntityID      *uuid.UUID       `json:"entity_id,omitempty" db:"entity_id"`
	TenantID         uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	GSTIN         string           `json:"gstin" db:"gstin"`
	BuyerGSTIN    *string          `json:"buyer_gstin,omitempty" db:"buyer_gstin"`
	TotalAmount   float64          `json:"total_amount" db:"total_amount"`
	TaxableAmount float64          `json:"taxable_amount" db:"taxable_amount"`
	CGSTAmount    float64          `json:"cgst_amount" db:"cgst_amount"`
	SGSTAmount    float64          `json:"sgst_amount" db:"sgst_amount"`
	IGSTAmount    float64          `json:"igst_amount" db:"igst_amount"`
	ErrorMessage  *string          `json:"error_message,omitempty" db:"error_message"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
}

type GenerateEInvoiceRequest struct {
	EntityType    string    `json:"entity_type" validate:"required,oneof=fuel_transaction expense_report"`
	EntityID      uuid.UUID `json:"entity_id" validate:"required,uuid"`
	TenantID         uuid.UUID `json:"tenant_id" validate:"required,uuid"`
	SellerGSTIN   string    `json:"seller_gstin" validate:"required,len=15"`
	BuyerGSTIN    *string   `json:"buyer_gstin,omitempty" validate:"omitempty,len=15"`
	TotalAmount   float64   `json:"total_amount" validate:"required,gt=0"`
	TaxableAmount float64   `json:"taxable_amount" validate:"required,gte=0"`
}

// --- TDS Records ---

type TDSRecord struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	EntityType     string     `json:"entity_type" db:"entity_type"`
	EntityID       uuid.UUID  `json:"entity_id" db:"entity_id"`
	VendorID       *uuid.UUID `json:"vendor_id,omitempty" db:"vendor_id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	PANNumber      *string    `json:"pan_number,omitempty" db:"pan_number"`
	TDSSection     string     `json:"tds_section" db:"tds_section"`
	GrossAmount    float64    `json:"gross_amount" db:"gross_amount"`
	TDSRate        float64    `json:"tds_rate" db:"tds_rate"`
	TDSAmount      float64    `json:"tds_amount" db:"tds_amount"`
	NetAmount      float64    `json:"net_amount" db:"net_amount"`
	DeductedAt     time.Time  `json:"deducted_at" db:"deducted_at"`
	Deposited      bool       `json:"deposited" db:"deposited"`
	DepositChallan *string    `json:"deposit_challan,omitempty" db:"deposit_challan"`
	FinancialYear  string     `json:"financial_year" db:"financial_year"`
	Quarter        string     `json:"quarter" db:"quarter"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type DeductTDSRequest struct {
	EntityType  string     `json:"entity_type" validate:"required"`
	EntityID    uuid.UUID  `json:"entity_id" validate:"required,uuid"`
	VendorID    *uuid.UUID `json:"vendor_id,omitempty"`
	TenantID       uuid.UUID  `json:"tenant_id" validate:"required,uuid"`
	GrossAmount float64    `json:"gross_amount" validate:"required,gt=0"`
	TDSSection  string     `json:"tds_section" validate:"required"`
}

// --- Vendor Payments ---

type VendorPaymentStatus string

const (
	VendorPaymentStatusPending    VendorPaymentStatus = "PENDING"
	VendorPaymentStatusProcessing VendorPaymentStatus = "PROCESSING"
	VendorPaymentStatusPaid       VendorPaymentStatus = "PAID"
	VendorPaymentStatusFailed     VendorPaymentStatus = "FAILED"
	VendorPaymentStatusReversed   VendorPaymentStatus = "REVERSED"
)

type VendorPayment struct {
	ID               uuid.UUID           `json:"id" db:"id"`
	VendorID         uuid.UUID           `json:"vendor_id" db:"vendor_id"`
	TenantID            uuid.UUID           `json:"tenant_id" db:"tenant_id"`
	InvoiceNumber    string              `json:"invoice_number" db:"invoice_number"`
	Amount           float64             `json:"amount" db:"amount"`
	TDSDeducted      float64             `json:"tds_deducted" db:"tds_deducted"`
	NetPayable       float64             `json:"net_payable" db:"net_payable"`
	Currency         string              `json:"currency" db:"currency"`
	PaymentStatus    VendorPaymentStatus `json:"payment_status" db:"payment_status"`
	PaymentMode      *string             `json:"payment_mode,omitempty" db:"payment_mode"`
	PaymentReference *string             `json:"payment_reference,omitempty" db:"payment_reference"`
	BankAccount      *string             `json:"bank_account,omitempty" db:"bank_account"`
	IFSCCode         *string             `json:"ifsc_code,omitempty" db:"ifsc_code"`
	PaymentDate      *time.Time          `json:"payment_date,omitempty" db:"payment_date"`
	PaidAt           *time.Time          `json:"paid_at,omitempty" db:"paid_at"`
	GSTInvoiceID     *uuid.UUID          `json:"gst_invoice_id,omitempty" db:"gst_invoice_id"`
	TDSRecordID      *uuid.UUID          `json:"tds_record_id,omitempty" db:"tds_record_id"`
	Notes            *string             `json:"notes,omitempty" db:"notes"`
	CreatedBy        *uuid.UUID          `json:"created_by,omitempty" db:"created_by"`
	CreatedAt        time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at" db:"updated_at"`
}

type CreateVendorPaymentRequest struct {
	VendorID      uuid.UUID `json:"vendor_id" validate:"required,uuid"`
	InvoiceNumber string    `json:"invoice_number" validate:"required"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	PaymentMode   string    `json:"payment_mode" validate:"required,oneof=NEFT RTGS IMPS CHEQUE CASH"`
	BankAccount   *string   `json:"bank_account,omitempty"`
	IFSCCode      *string   `json:"ifsc_code,omitempty"`
	Notes         *string   `json:"notes,omitempty"`
}

// --- Expense Items ---

type ExpenseItemCategory string

const (
	ExpenseItemCategoryFuel        ExpenseItemCategory = "FUEL"
	ExpenseItemCategoryMeals       ExpenseItemCategory = "MEALS"
	ExpenseItemCategoryLodging     ExpenseItemCategory = "LODGING"
	ExpenseItemCategoryToll        ExpenseItemCategory = "TOLL"
	ExpenseItemCategoryMaintenance ExpenseItemCategory = "MAINTENANCE"
	ExpenseItemCategoryOther       ExpenseItemCategory = "OTHER"
)

type ExpenseItem struct {
	ID              uuid.UUID           `json:"id" db:"id"`
	ReportID        uuid.UUID           `json:"report_id" db:"report_id"`
	InvoiceNumber   *string             `json:"invoice_number,omitempty" db:"invoice_number"`
	Category        ExpenseItemCategory `json:"category" db:"category"`
	Amount          float64             `json:"amount" db:"amount"`
	GSTAmount       float64             `json:"gst_amount" db:"gst_amount"`
	GSTRate         float64             `json:"gst_rate" db:"gst_rate"`
	ReceiptURL      *string             `json:"receipt_url,omitempty" db:"receipt_url"`
	Description     *string             `json:"description,omitempty" db:"description"`
	VendorName      *string             `json:"vendor_name,omitempty" db:"vendor_name"`
	VendorGSTIN     *string             `json:"vendor_gstin,omitempty" db:"vendor_gstin"`
	TransactionDate time.Time           `json:"transaction_date" db:"transaction_date"`
	OCRProcessed    bool                `json:"ocr_processed" db:"ocr_processed"`
	OCRConfidence   *float64            `json:"ocr_confidence,omitempty" db:"ocr_confidence"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
}

type CreateExpenseItemRequest struct {
	ReportID        uuid.UUID           `json:"report_id" validate:"required,uuid"`
	InvoiceNumber   *string             `json:"invoice_number,omitempty"`
	Category        ExpenseItemCategory `json:"category" validate:"required,oneof=FUEL MEALS LODGING TOLL MAINTENANCE OTHER"`
	Amount          float64             `json:"amount" validate:"required,gt=0"`
	GSTRate         float64             `json:"gst_rate" validate:"gte=0,lte=28"`
	ReceiptURL      *string             `json:"receipt_url,omitempty"`
	Description     *string             `json:"description,omitempty"`
	VendorName      *string             `json:"vendor_name,omitempty"`
	VendorGSTIN     *string             `json:"vendor_gstin,omitempty"`
	TransactionDate time.Time           `json:"transaction_date" validate:"required"`
}

// --- Fraud Incidents ---

type FraudRiskLevel string

const (
	FraudRiskLow      FraudRiskLevel = "LOW"
	FraudRiskMedium   FraudRiskLevel = "MEDIUM"
	FraudRiskHigh     FraudRiskLevel = "HIGH"
	FraudRiskCritical FraudRiskLevel = "CRITICAL"
)

type FraudIncident struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	EntityType     string         `json:"entity_type" db:"entity_type"`
	EntityID       uuid.UUID      `json:"entity_id" db:"entity_id"`
	TenantID          uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	VehicleID      *uuid.UUID     `json:"vehicle_id,omitempty" db:"vehicle_id"`
	DriverID       *uuid.UUID     `json:"driver_id,omitempty" db:"driver_id"`
	FraudScore     float64        `json:"fraud_score" db:"fraud_score"`
	RiskLevel      FraudRiskLevel `json:"risk_level" db:"risk_level"`
	Signals        []string       `json:"signals" db:"signals"`
	Status         string         `json:"status" db:"status"`
	InvestigatedBy *uuid.UUID     `json:"investigated_by,omitempty" db:"investigated_by"`
	ResolvedAt     *time.Time     `json:"resolved_at,omitempty" db:"resolved_at"`
	Notes          *string        `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

// --- Trip Profitability ---

type TripProfitability struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TripID          uuid.UUID  `json:"trip_id" db:"trip_id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	VehicleID       *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	DriverID        *uuid.UUID `json:"driver_id,omitempty" db:"driver_id"`
	TripRevenue     float64    `json:"trip_revenue" db:"trip_revenue"`
	FuelCost        float64    `json:"fuel_cost" db:"fuel_cost"`
	TollCost        float64    `json:"toll_cost" db:"toll_cost"`
	DriverCost      float64    `json:"driver_cost" db:"driver_cost"`
	MaintenanceCost float64    `json:"maintenance_cost" db:"maintenance_cost"`
	OtherCost       float64    `json:"other_cost" db:"other_cost"`
	TotalCost       float64    `json:"total_cost" db:"total_cost"`
	Profit          float64    `json:"profit" db:"profit"`
	ProfitMargin    float64    `json:"profit_margin" db:"profit_margin"`
	DistanceKM      float64    `json:"distance_km" db:"distance_km"`
	CostPerKM       float64    `json:"cost_per_km" db:"cost_per_km"`
	CalculatedAt    time.Time  `json:"calculated_at" db:"calculated_at"`
}

// --- Dashboard Metrics ---

type FinanceDashboardMetrics struct {
	TenantID         uuid.UUID `json:"tenant_id"`
	MetricMonth   time.Time `json:"metric_month"`
	TotalRevenue  float64   `json:"total_revenue"`
	TotalFuelCost float64   `json:"total_fuel_cost"`
	TotalExpenses float64   `json:"total_expenses"`
	ProfitMargin  float64   `json:"profit_margin"`
	AvgCostPerKm  float64   `json:"avg_cost_per_km"`
}

// --- Fuel Card Reconciliation ---

type ReconciliationStatus string

const (
	ReconciliationStatusPending   ReconciliationStatus = "PENDING"
	ReconciliationStatusMatched   ReconciliationStatus = "MATCHED"
	ReconciliationStatusUnmatched ReconciliationStatus = "UNMATCHED"
	ReconciliationStatusDisputed  ReconciliationStatus = "DISPUTED"
)

type FuelCardReconciliation struct {
	ID                    uuid.UUID            `json:"id" db:"id"`
	CardTransactionID     *uuid.UUID           `json:"card_transaction_id,omitempty" db:"card_transaction_id"`
	InternalTransactionID *uuid.UUID           `json:"internal_transaction_id,omitempty" db:"internal_transaction_id"`
	TenantID                 uuid.UUID            `json:"tenant_id" db:"tenant_id"`
	Status                ReconciliationStatus `json:"status" db:"status"`
	Variance              float64              `json:"variance" db:"variance"`
	CardAmount            *float64             `json:"card_amount,omitempty" db:"card_amount"`
	InternalAmount        *float64             `json:"internal_amount,omitempty" db:"internal_amount"`
	ReconciledBy          *uuid.UUID           `json:"reconciled_by,omitempty" db:"reconciled_by"`
	ReconciledAt          *time.Time           `json:"reconciled_at,omitempty" db:"reconciled_at"`
	Notes                 *string              `json:"notes,omitempty" db:"notes"`
	CreatedAt             time.Time            `json:"created_at" db:"created_at"`
}

type FuelCardTransaction struct {
	ID                    uuid.UUID      `json:"id" db:"id"`
	TenantID                 uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	VehicleID             *uuid.UUID     `json:"vehicle_id,omitempty" db:"vehicle_id"`
	CardID                string         `json:"card_id" db:"card_id"`
	CardNumber            string         `json:"card_number" db:"card_number"`
	Provider              string         `json:"provider" db:"provider"`
	Vendor                string         `json:"vendor" db:"vendor"`
	Amount                float64        `json:"amount" db:"amount"`
	Currency              string         `json:"currency" db:"currency"`
	TransactionTime       time.Time      `json:"transaction_time" db:"transaction_time"`
	Location              string         `json:"location" db:"location"`
	Status                string         `json:"status" db:"status"`
	ExternalTransactionID string         `json:"external_transaction_id" db:"external_transaction_id"`
	Metadata              map[string]any `json:"metadata" db:"metadata"`
	CreatedAt             time.Time      `json:"created_at" db:"created_at"`
}
