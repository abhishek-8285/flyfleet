package models

import (
	"time"

	"github.com/google/uuid"
)

// TaxRegion represents a geographical area for tax calculation
type TaxRegion struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	CountryCode string    `json:"country_code" db:"country_code"`
	StateCode   *string   `json:"state_code,omitempty" db:"state_code"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TaxRate represents a specific tax percentage
type TaxRate struct {
	ID         uuid.UUID `json:"id" db:"id"`
	RegionID   uuid.UUID `json:"region_id" db:"region_id"`
	Name       string    `json:"name" db:"name"`
	Percentage float64   `json:"percentage" db:"percentage"`
	TaxType    string    `json:"tax_type" db:"tax_type"` // CGST, SGST, IGST, VAT
	IsDefault  bool      `json:"is_default" db:"is_default"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// InvoiceTaxLine represents a tax entry on an invoice
type InvoiceTaxLine struct {
	ID         uuid.UUID `json:"id" db:"id"`
	InvoiceID  uuid.UUID `json:"invoice_id" db:"invoice_id"`
	TaxRateID  uuid.UUID `json:"tax_rate_id" db:"tax_rate_id"`
	Name       string    `json:"name" db:"name"`
	Amount     float64   `json:"amount" db:"amount"`
	Percentage float64   `json:"percentage" db:"percentage"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// InvoiceSequence tracks the gapless sequential numbering per organization
type InvoiceSequence struct {
	OrgID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Prefix    string    `json:"prefix" db:"prefix"`
	NextValue int64     `json:"next_value" db:"next_value"`
	Suffix    string    `json:"suffix" db:"suffix"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
