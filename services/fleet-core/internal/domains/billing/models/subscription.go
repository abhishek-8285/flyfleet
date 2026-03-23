package models

import (
	"time"

	"github.com/google/uuid"
)

// PlanStatus represents the status of a billing plan
type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "ACTIVE"
	PlanStatusInactive PlanStatus = "INACTIVE"
)

// PricingComponentType represents the type of a pricing component
type PricingComponentType string

const (
	ComponentFlat    PricingComponentType = "FLAT"
	ComponentPerUnit PricingComponentType = "PER_UNIT"
	ComponentUsage   PricingComponentType = "USAGE"
)

// Frequency represents the billing frequency
type Frequency string

const (
	FrequencyMonthly Frequency = "MONTHLY"
	FrequencyYearly  Frequency = "YEARLY"
	FrequencyOneTime Frequency = "ONE_TIME"
)

// Plan represents a billing plan
type Plan struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	IsActive    bool        `json:"is_active" db:"is_active"`
	Components  []PricingComponent `json:"components,omitempty"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// PricingComponent represents a single charge within a plan
type PricingComponent struct {
	ID              uuid.UUID            `json:"id" db:"id"`
	PlanID          uuid.UUID            `json:"plan_id" db:"plan_id"`
	Name            string               `json:"name" db:"name"`
	Type            PricingComponentType `json:"type" db:"type"`
	UnitPrice       float64              `json:"unit_price" db:"unit_price"`
	Currency        string               `json:"currency" db:"currency"`
	MeterEventName  string               `json:"meter_event_name" db:"meter_event_name"`
	Frequency       Frequency            `json:"frequency" db:"frequency"`
	CreatedAt       time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" db:"updated_at"`
}

// SubscriptionStatus represents the status of an organization's subscription
type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "ACTIVE"
	SubscriptionTrial     SubscriptionStatus = "TRIAL"
	SubscriptionCancelled SubscriptionStatus = "CANCELLED"
	SubscriptionPastDue   SubscriptionStatus = "PAID_DUE"
)

// Subscription represents an organization's active plan
type Subscription struct {
	ID                  uuid.UUID          `json:"id" db:"id"`
	OrgID               uuid.UUID          `json:"org_id" db:"org_id"`
	PlanID              uuid.UUID          `json:"plan_id" db:"plan_id"`
	Status              SubscriptionStatus `json:"status" db:"status"`
	CurrentPeriodStart   time.Time          `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd     time.Time          `json:"current_period_end" db:"current_period_end"`
	TrialEnd            *time.Time         `json:"trial_end,omitempty" db:"trial_end"`
	CancelAtPeriodEnd   bool               `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	Metadata            map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt           time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at" db:"updated_at"`
}
