package models

import (
	"time"

	"github.com/google/uuid"
)

// UsageMeter defines a billable metric
type UsageMeter struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Unit        string    `json:"unit" db:"unit"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	Currency    string    `json:"currency" db:"currency"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ConsumptionEvent represents an atomic unit of usage
type ConsumptionEvent struct {
	Time               time.Time              `json:"time" db:"time"`
	OrgID              uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	VehicleID          *uuid.UUID             `json:"vehicle_id,omitempty" db:"vehicle_id"`
	MeterCode          string                 `json:"meter_code" db:"meter_code"`
	Amount             float64                `json:"amount" db:"amount"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	ProcessedInvoiceID *uuid.UUID             `json:"processed_invoice_id,omitempty" db:"processed_invoice_id"`
}

// UsageSummary represents aggregated usage for an invoice
type UsageSummary struct {
	MeterCode string  `json:"meter_code"`
	Name      string  `json:"name"`
	Unit      string  `json:"unit"`
	Total     float64 `json:"total"`
	Amount    float64 `json:"amount"` // Total * UnitPrice
}
