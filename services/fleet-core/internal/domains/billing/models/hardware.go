package models

import (
	"time"

	"github.com/google/uuid"
)

type HardwareCategory string

const (
	CategoryDashcam HardwareCategory = "DASHCAM"
	CategoryTracker HardwareCategory = "TRACKER"
	CategorySensor  HardwareCategory = "SENSOR"
)

type BillingType string

const (
	BillingOneTime BillingType = "ONE_TIME"
	BillingLease   BillingType = "LEASE"
)

type HardwareProduct struct {
	ID                uuid.UUID        `json:"id" db:"id"`
	Name              string           `json:"name" db:"name"`
	ModelCode         string           `json:"model_code" db:"model_code"`
	Category          HardwareCategory `json:"category" db:"category"`
	OneTimePrice      float64          `json:"one_time_price" db:"one_time_price"`
	LeaseMonthlyPrice *float64         `json:"lease_monthly_price,omitempty" db:"lease_monthly_price"`
	Currency          string           `json:"currency" db:"currency"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
}

type DeviceAssignment struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	OrgID       uuid.UUID   `json:"org_id" db:"org_id"`
	DeviceID    string      `json:"device_id" db:"device_id"`
	ProductID   uuid.UUID   `json:"product_id" db:"product_id"`
	BillingType BillingType `json:"billing_type" db:"billing_type"`
	Status      string      `json:"status" db:"status"` // PENDING, ACTIVE, TERMINATED
	ActivatedAt *time.Time  `json:"activated_at,omitempty" db:"activated_at"`
	InvoiceID   *uuid.UUID  `json:"invoice_id,omitempty" db:"invoice_id"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}
