package models

import (
	"time"
)

// FuelTransaction represents a fuel purchase event
type FuelTransaction struct {
	ID                    string    `json:"id" db:"id"`
	VehicleID             string    `json:"vehicle_id" db:"vehicle_id"`
	DriverID              *string   `json:"driver_id,omitempty" db:"driver_id"` // Nullable
	CardID                *string   `json:"card_id,omitempty" db:"card_id"`
	StationLocation       *string   `json:"station_location,omitempty" db:"station_location"`
	Liters                float64   `json:"liters" db:"liters"`
	Amount                float64   `json:"amount" db:"amount"`
	Currency              string    `json:"currency" db:"currency"`
	OdometerReading       *float64  `json:"odometer_reading,omitempty" db:"odometer_reading"`
	EfficiencyCalculation *float64  `json:"efficiency_calculation,omitempty" db:"efficiency_calculation"`
	AnomalyFlag           bool      `json:"anomaly_flag" db:"anomaly_flag"`
	TransactionDate       time.Time `json:"transaction_date" db:"transaction_date"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// ExpensePolicy defines spending limits and approval rules
type ExpensePolicy struct {
	ID               string    `json:"id" db:"id"`
	OrgID            string    `json:"org_id" db:"org_id"`
	Category         string    `json:"category" db:"category"` // MEALS, LODGING, FUEL, MAINTENANCE
	LimitAmount      *float64  `json:"limit_amount,omitempty" db:"limit_amount"`
	Currency         string    `json:"currency" db:"currency"`
	RequiresApproval bool      `json:"requires_approval" db:"requires_approval"`
	ReceiptRequired  bool      `json:"receipt_required" db:"receipt_required"`
	Description      *string   `json:"description,omitempty" db:"description"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ExpenseReport aggregates expenses for a period
type ExpenseReport struct {
	ID              string                 `json:"id" db:"id"`
	OrgID           string                 `json:"org_id" db:"org_id"`
	SubmitterID     string                 `json:"submitter_id" db:"submitter_id"`
	PeriodStart     *time.Time             `json:"period_start" db:"period_start"`
	PeriodEnd       *time.Time             `json:"period_end" db:"period_end"`
	TotalAmount     float64                `json:"total_amount" db:"total_amount"`
	Currency        string                 `json:"currency" db:"currency"`
	Status          string                 `json:"status" db:"status"` // DRAFT, SUBMITTED, APPROVED, REJECTED, PAID
	ApprovedBy      *string                `json:"approved_by,omitempty" db:"approved_by"`
	ApprovalDate    *time.Time             `json:"approval_date,omitempty" db:"approval_date"`
	RejectionReason *string                `json:"rejection_reason,omitempty" db:"rejection_reason"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

type ExpenseReportListResponse struct {
	Reports []ExpenseReport `json:"reports"`
	Total   int             `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
}

type RevenueForecast struct {
	RouteID         string  `json:"route_id"`
	PredictedDemand float64 `json:"predicted_demand"`
	CurrentDemand   float64 `json:"current_demand"`
	SuggestedPrice  float64 `json:"suggested_price"`
	BasePrice       float64 `json:"base_price"`
	ConfidencePct   float64 `json:"confidence_pct"`
}

type ABTestExperiment struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Variants  []string               `json:"variants"`
	Metrics   map[string]interface{} `json:"metrics"`
	CreatedAt time.Time              `json:"created_at"`
}

// ApprovalInstance tracks a multi-level approval process
type ApprovalInstance struct {
	ID                string    `json:"id" db:"id"`
	EntityID          string    `json:"entity_id" db:"entity_id"`
	EntityType        string    `json:"entity_type" db:"entity_type"`
	ConfigID          string    `json:"config_id" db:"config_id"`
	CurrentLevelIndex int       `json:"current_level_index" db:"current_level_index"`
	Status            string    `json:"status" db:"status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// ApprovalAction records a specific approver's decision
type ApprovalAction struct {
	ID         string    `json:"id" db:"id"`
	InstanceID string    `json:"instance_id" db:"instance_id"`
	ApproverID string    `json:"approver_id" db:"approver_id"`
	Action     string    `json:"action" db:"action"`
	LevelIndex int       `json:"level_index" db:"level_index"`
	Comments   string    `json:"comments" db:"comments"`
	ActionTime time.Time `json:"action_time" db:"action_time"`
}

// ApprovalConfig defines the workflow for approvals
type ApprovalConfig struct {
	ID         string   `json:"id" db:"id"`
	OrgID      string   `json:"org_id" db:"org_id"`
	EntityType string   `json:"entity_type" db:"entity_type"`
	MinAmount  float64  `json:"min_amount" db:"min_amount"`
	Levels     []string `json:"levels" db:"levels"` // List of role names or user IDs
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Budget tracks spending against a limit
type Budget struct {
	ID             string    `json:"id" db:"id"`
	OrgID          string    `json:"org_id" db:"org_id"`
	Category       string    `json:"category" db:"category"`
	Year           int       `json:"year" db:"year"`
	Month          int       `json:"month" db:"month"`
	LimitAmount    float64   `json:"limit_amount" db:"limit_amount"`
	SpentAmount    float64   `json:"spent_amount" db:"spent_amount"`
	Currency          string    `json:"currency" db:"currency"`
	AlertThresholdPct int       `json:"alert_threshold_pct" db:"alert_threshold_pct"`
	PeriodStart       time.Time `json:"period_start" db:"period_start"`
	PeriodEnd      time.Time `json:"period_end" db:"period_end"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
