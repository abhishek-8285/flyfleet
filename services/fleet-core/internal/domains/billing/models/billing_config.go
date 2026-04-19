package models

import (
	"time"

	"github.com/google/uuid"
)

// BillingConfiguration represents per-organization billing customization
type BillingConfiguration struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrgID          uuid.UUID `json:"tenant_id" db:"tenant_id" validate:"required"`
	CompanyName    string    `json:"company_name" db:"company_name" validate:"required"`
	CompanyAddress string    `json:"company_address" db:"company_address"`
	CompanyGSTIN   string    `json:"company_gstin" db:"company_gstin"`
	CompanyPAN     string    `json:"company_pan" db:"company_pan"`
	CompanyEmail   string    `json:"company_email" db:"company_email" validate:"omitempty,email"`
	CompanyPhone   string    `json:"company_phone" db:"company_phone"`
	CompanyWebsite string    `json:"company_website" db:"company_website"`
	CompanyLogoURL string    `json:"company_logo_url" db:"company_logo_url"`

	// Branding
	PrimaryColor   string `json:"primary_color" db:"primary_color" validate:"hexcolor"`
	SecondaryColor string `json:"secondary_color" db:"secondary_color" validate:"hexcolor"`
	Theme          string `json:"theme" db:"theme" validate:"oneof=modern classic minimal"`

	// Payment Details
	BankName          string `json:"bank_name" db:"bank_name"`
	BankAccountName   string `json:"bank_account_name" db:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number" db:"bank_account_number"`
	BankIFSC          string `json:"bank_ifsc" db:"bank_ifsc"`
	BankBranch        string `json:"bank_branch" db:"bank_branch"`
	BankUPI           string `json:"bank_upi" db:"bank_upi"`

	// Invoice Settings
	InvoicePrefix   string   `json:"invoice_prefix" db:"invoice_prefix"`
	InvoiceTerms    []string `json:"invoice_terms" db:"invoice_terms"`
	InvoiceNotes    string   `json:"invoice_notes" db:"invoice_notes"`
	ShowHSN         bool     `json:"show_hsn" db:"show_hsn"`
	ShowQRCode      bool     `json:"show_qr_code" db:"show_qr_code"`
	ShowPaymentInfo bool     `json:"show_payment_info" db:"show_payment_info"`

	// Defaults
	DefaultCurrency         string `json:"default_currency" db:"default_currency" validate:"len=3"`
	DefaultTaxCalculation   string `json:"default_tax_calculation" db:"default_tax_calculation" validate:"oneof=inclusive exclusive"`
	DefaultPaymentTermsDays int    `json:"default_payment_terms_days" db:"default_payment_terms_days"`

	// Status
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

// BillingConfigRequest represents request to create/update billing config
type BillingConfigRequest struct {
	CompanyName    string `json:"company_name" validate:"required"`
	CompanyAddress string `json:"company_address"`
	CompanyGSTIN   string `json:"company_gstin"`
	CompanyPAN     string `json:"company_pan"`
	CompanyEmail   string `json:"company_email" validate:"omitempty,email"`
	CompanyPhone   string `json:"company_phone"`
	CompanyWebsite string `json:"company_website"`
	CompanyLogoURL string `json:"company_logo_url"`

	PrimaryColor   string `json:"primary_color" validate:"hexcolor"`
	SecondaryColor string `json:"secondary_color" validate:"hexcolor"`
	Theme          string `json:"theme" validate:"oneof=modern classic minimal"`

	BankName          string `json:"bank_name"`
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	BankIFSC          string `json:"bank_ifsc"`
	BankBranch        string `json:"bank_branch"`
	BankUPI           string `json:"bank_upi"`

	InvoicePrefix   string   `json:"invoice_prefix"`
	InvoiceTerms    []string `json:"invoice_terms"`
	InvoiceNotes    string   `json:"invoice_notes"`
	ShowHSN         *bool    `json:"show_hsn"`
	ShowQRCode      *bool    `json:"show_qr_code"`
	ShowPaymentInfo *bool    `json:"show_payment_info"`

	DefaultCurrency         string `json:"default_currency" validate:"len=3"`
	DefaultTaxCalculation   string `json:"default_tax_calculation" validate:"oneof=inclusive exclusive"`
	DefaultPaymentTermsDays int    `json:"default_payment_terms_days"`
}

// DefaultBillingConfig returns default billing configuration
func DefaultBillingConfig(orgID uuid.UUID) *BillingConfiguration {
	return &BillingConfiguration{
		ID:                      uuid.New(),
		OrgID:                   orgID,
		CompanyName:             "Organization",
		PrimaryColor:            "#3b82f6",
		SecondaryColor:          "#1e40af",
		Theme:                   "modern",
		InvoicePrefix:           "INV",
		ShowHSN:                 true,
		ShowQRCode:              true,
		ShowPaymentInfo:         true,
		DefaultCurrency:         "INR",
		DefaultTaxCalculation:   "exclusive",
		DefaultPaymentTermsDays: 30,
		IsActive:                true,
	}
}
