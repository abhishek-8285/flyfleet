package models

import "time"

// UniversalRenderData is the single source of truth for ALL invoice templates.
// Any new template (Modern, Classic, etc.) will use this data structure.
type UniversalRenderData struct {
	// Branding & Company
	CompanyName    string `json:"company_name"`
	CompanyLogo    string `json:"company_logo"`
	CompanyAddress string `json:"company_address"`
	CompanyGSTIN   string `json:"company_gstin"`
	CompanyPAN     string `json:"company_pan"`
	CompanyEmail   string `json:"company_email"`
	CompanyPhone   string `json:"company_phone"`
	CompanyWebsite string `json:"company_website"`

	// Customer
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	CustomerGSTIN   string `json:"customer_gstin"`
	CustomerEmail   string `json:"customer_email"`

	// Invoice Metadata
	InvoiceNumber string `json:"invoice_number"`
	InvoiceDate   string `json:"invoice_date"`
	DueDate       string `json:"due_date"`
	Status        string `json:"status"`
	Currency      string `json:"currency"`
	PlaceOfSupply string `json:"place_of_supply"`

	// Line Items
	Items []RenderItem `json:"items"`

	// Financials
	Subtotal     float64 `json:"subtotal"`
	TaxableValue float64 `json:"taxable_value"`
	CGST         float64 `json:"cgst"`
	SGST         float64 `json:"sgst"`
	IGST         float64 `json:"igst"`
	TotalTax     float64 `json:"total_tax"`
	TotalAmount  float64 `json:"total_amount"`
	AmountPaid   float64 `json:"amount_paid"`
	BalanceDue   float64 `json:"balance_due"`

	// Payment & Terms
	BankName      string   `json:"bank_name"`
	AccountName   string   `json:"account_name"`
	AccountNumber string   `json:"account_number"`
	IFSC          string   `json:"ifsc"`
	UPI           string   `json:"upi"`
	QRCode        string   `json:"qr_code"`
	Notes         string   `json:"notes"`
	Terms         []string `json:"terms"`

	// System Info
	GeneratedAt time.Time `json:"generated_at"`
}

type RenderItem struct {
	Description string  `json:"description"`
	HSNCode     string  `json:"hsn_code"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TaxPercent  float64 `json:"tax_percent"`
	TaxAmount   float64 `json:"tax_amount"`
	Total       float64 `json:"total"`
}
