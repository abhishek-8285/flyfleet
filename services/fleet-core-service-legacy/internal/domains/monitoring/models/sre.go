package models

import (
	"time"

	"github.com/google/uuid"
)

// SyntheticCheck represents a single synthetic monitoring check result
type SyntheticCheck struct {
	ID             uuid.UUID `json:"id" db:"id"`
	CheckName      string    `json:"check_name" db:"check_name"`
	Status         string    `json:"status" db:"status"` // UP, DOWN, DEGRADED
	ResponseTimeMS int       `json:"response_time_ms" db:"response_time_ms"`
	Error          string    `json:"error,omitempty" db:"error"`
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
}

// ServiceSLO represents a Service Level Objective
type ServiceSLO struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	ServiceName        string    `json:"service_name" db:"service_name"`
	AvailabilityTarget float64   `json:"availability_target" db:"availability_target"`
	LatencyTargetMS    int       `json:"latency_target_ms" db:"latency_target_ms"`
	ErrorRateTarget    float64   `json:"error_rate_target" db:"error_rate_target"`
	PeriodDays         int       `json:"period_days" db:"period_days"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// ErrorBudget represents the error budget for a service
type ErrorBudget struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	ServiceName           string    `json:"service_name" db:"service_name"`
	AllowedDowntimeMinutes int       `json:"allowed_downtime_minutes" db:"allowed_downtime_minutes"`
	ConsumedMinutes        int       `json:"consumed_minutes" db:"consumed_minutes"`
	PeriodStart           time.Time `json:"period_start" db:"period_start"`
	PeriodEnd             time.Time `json:"period_end" db:"period_end"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// AlertGroup represents a grouped set of alerts
type AlertGroup struct {
	ID             uuid.UUID `json:"id" db:"id"`
	GroupKey       string    `json:"group_key" db:"group_key"` // component:metric:severity
	Component      string    `json:"component" db:"component"`
	MetricName     string    `json:"metric_name" db:"metric_name"`
	Severity       string    `json:"severity" db:"severity"`
	Status         string    `json:"status" db:"status"` // ACTIVE, SUPPRESSED, RESOLVED
	AlertCount     int       `json:"alert_count" db:"alert_count"`
	FirstTriggered time.Time `json:"first_triggered" db:"first_triggered"`
	LastTriggered  time.Time `json:"last_triggered" db:"last_triggered"`
}

// RemediationAction represents an automated recovery action
type RemediationAction struct {
	ID             uuid.UUID `json:"id" db:"id"`
	GroupKey       string    `json:"group_key" db:"group_key"`
	ActionType     string    `json:"action_type" db:"action_type"`
	Status         string    `json:"status" db:"status"` // PENDING, EXECUTED, FAILED, SUCCESS
	ResultMetadata map[string]interface{} `json:"result_metadata" db:"result_metadata"`
	TriggeredAt    time.Time `json:"triggered_at" db:"triggered_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// Incident represents an incident lifecycle record
type Incident struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	Severity       string    `json:"severity" db:"severity"`
	Status         string    `json:"status" db:"status"` // DETECTED, ACKNOWLEDGED, ESCALATED, RESOLVED
	GroupKey       string    `json:"group_key,omitempty" db:"group_key"`
	ServiceName    string    `json:"service_name" db:"service_name"`
	DetectedAt     time.Time `json:"detected_at" db:"detected_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ExternalRefID  string    `json:"external_ref_id,omitempty" db:"external_ref_id"`
}

// Response Models for SRE APIs

type CapacityForecast struct {
	MetricName    string    `json:"metric_name"`
	CurrentValue  float64   `json:"current_value"`
	ForecastValue float64   `json:"forecast_value"`
	HorizonDays   int       `json:"horizon_days"`
	Confidence    float64   `json:"confidence"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type BusinessKPIs struct {
	ActiveVehicles      int64     `json:"active_vehicles"`
	ActiveDrivers       int64     `json:"active_drivers"`
	TelemetryRatePerSec float64   `json:"telemetry_ingestion_rate_per_sec"`
	WorkOrdersPerHour   float64   `json:"work_orders_created_per_hour"`
	ActiveTenants       int       `json:"active_tenants"`
	Timestamp           time.Time `json:"timestamp"`
}

type RegionHealth struct {
	RegionName    string    `json:"region_name"`
	Status        string    `json:"status"` // HEALTHY, DEGRADED, DOWN
	LatencyMS     int       `json:"latency_ms"`
	FailoverReady bool      `json:"failover_ready"`
	LastChecked   time.Time `json:"last_checked"`
}
