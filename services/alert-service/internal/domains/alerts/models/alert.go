package models

import (
	"time"

	"github.com/google/uuid"
)

// AlertRule represents a rule for generating alerts
type AlertRule struct {
	ID          string            `json:"id" db:"id"`
	OrgID       uuid.UUID         `json:"org_id" db:"org_id"`
	Name        string            `json:"name" db:"name"`
	Type        string            `json:"type" db:"type"` // overspeed, geofence_entry, geofence_exit, harsh_acceleration, etc.
	Description string            `json:"description" db:"description"`
	Conditions  map[string]interface{} `json:"conditions" db:"conditions"` // Rule-specific conditions
	Severity    string            `json:"severity" db:"severity"` // info, warning, critical
	Actions     []AlertAction     `json:"actions" db:"actions"`
	Vehicles    []uuid.UUID       `json:"vehicles" db:"vehicles"` // Specific vehicles or empty for all
	DriverIDs   []uuid.UUID       `json:"driver_ids" db:"driver_ids"` // Specific drivers or empty for all
	Enabled     bool              `json:"enabled" db:"enabled"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	CreatedBy   uuid.UUID         `json:"created_by" db:"created_by"`
}

// AlertAction represents an action to take when an alert is triggered
type AlertAction struct {
	Type     string `json:"type" db:"type"`     // notification, email, sms, immobilize, webhook
	Target   string `json:"target" db:"target"` // recipient email/phone or webhook URL
	Template string `json:"template" db:"template"` // template ID for notification
	Delay    int    `json:"delay" db:"delay"`   // delay in seconds before executing
}

// Alert represents an individual alert instance
type Alert struct {
	ID              string            `json:"id" db:"id"`
	OrgID           uuid.UUID         `json:"org_id" db:"org_id"`
	RuleID          string            `json:"rule_id" db:"rule_id"`
	Type            string            `json:"type" db:"type"`
	Severity        string            `json:"severity" db:"severity"`
	VehicleID       *uuid.UUID        `json:"vehicle_id" db:"vehicle_id"`
	DriverID        *uuid.UUID        `json:"driver_id" db:"driver_id"`
	Location        *Location         `json:"location" db:"location"`
	Timestamp       time.Time         `json:"timestamp" db:"timestamp"`
	Message         string            `json:"message" db:"message"`
	Details         map[string]interface{} `json:"details" db:"details"` // Rule-specific data
	Status          string            `json:"status" db:"status"` // new, acknowledged, resolved, escalated
	DedupKey        string            `json:"dedup_key" db:"dedup_key"`
	EscalationLevel int               `json:"escalation_level" db:"escalation_level"`
	LastEscalatedAt *time.Time        `json:"last_escalated_at" db:"last_escalated_at"`
	SLADeadline     *time.Time        `json:"sla_deadline" db:"sla_deadline"`
	Acknowledged    bool              `json:"acknowledged" db:"-"` // Calculated field for frontend
	AcknowledgedBy  *uuid.UUID        `json:"acknowledged_by" db:"acknowledged_by"`
	AcknowledgedAt  *time.Time        `json:"acknowledged_at" db:"acknowledged_at"`
	ResolvedBy      *uuid.UUID        `json:"resolved_by" db:"resolved_by"`
	ResolvedAt      *time.Time        `json:"resolved_at" db:"resolved_at"`
	EscalatedTo     []uuid.UUID       `json:"escalated_to" db:"escalated_to"`
	Notes           string            `json:"notes" db:"notes"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
	Address   string  `json:"address" db:"address"`
}

// Geofence represents a geographic boundary for monitoring
type Geofence struct {
	ID             string      `json:"id" db:"id"`
	OrgID          uuid.UUID   `json:"org_id" db:"org_id"`
	Name           string      `json:"name" db:"name"`
	Type           string      `json:"type" db:"type"` // circle, polygon, route_corridor
	Center         *Location   `json:"center" db:"center"`         // For circle geofences
	RadiusMeters   *int        `json:"radius" db:"radius_meters"` // For circle geofences
	Coordinates    []Location  `json:"coordinates" db:"coordinates"`     // For polygon geofences
	RoutePolyline  *string     `json:"route_polyline" db:"route_polyline"` // For route corridors
	BufferMeters   *int        `json:"buffer_meters" db:"buffer_meters"`   // For route corridors
	EntryAlert     bool        `json:"entry_alert" db:"entry_alert"`
	ExitAlert      bool        `json:"exit_alert" db:"exit_alert"`
	SpeedLimit     *int        `json:"speed_limit" db:"speed_limit"`
	Vehicles       []uuid.UUID `json:"vehicles" db:"vehicles"` // Specific vehicles or empty for all
	Enabled        bool        `json:"active" db:"enabled"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
	CreatedBy      uuid.UUID   `json:"created_by" db:"created_by"`
}

// NotificationTemplate represents a template for alert notifications
type NotificationTemplate struct {
	ID          string    `json:"id" db:"id"`
	OrgID       uuid.UUID `json:"org_id" db:"org_id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"` // sms, email, push
	Subject     string    `json:"subject" db:"subject"` // For email
	Body        string    `json:"body" db:"body"`
	Variables   []string  `json:"variables" db:"variables"` // Available template variables
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
}

// AlertRuleType represents different types of alert rules
const (
	AlertTypeOverspeed          = "overspeed"
	AlertTypeGeofenceEntry      = "geofence_entry"
	AlertTypeGeofenceExit       = "geofence_exit"
	AlertTypeHarshAcceleration  = "harsh_acceleration"
	AlertTypeHarshBraking       = "harsh_braking"
	AlertTypeHarshCornering     = "harsh_cornering"
	AlertTypeIdleTime           = "idle_time"
	AlertTypeFuelTheft          = "fuel_theft"
	AlertTypeLowBattery         = "low_battery"
	AlertTypeTemperatureExcursion = "temperature_excursion"
	AlertTypeAccidentDetected   = "accident_detected"
	AlertTypeMaintenanceDue     = "maintenance_due"
	AlertTypeDocumentExpiry     = "document_expiry"
	AlertTypeDriverFatigue      = "driver_fatigue"
	AlertTypeADASEvent          = "adas_event"
	AlertTypeVehicleOffline     = "vehicle_offline"
)

// AlertSeverity represents severity levels
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// AlertStatus represents alert statuses
const (
	AlertStatusNew         = "new"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusResolved    = "resolved"
	AlertStatusEscalated   = "escalated"
)

// ActionType represents different action types
const (
	ActionTypeNotification = "notification"
	ActionTypeEmail        = "email"
	ActionTypeSMS          = "sms"
	ActionTypeImmobilize   = "immobilize"
	ActionTypeWebhook      = "webhook"
	ActionTypePush         = "push"
)

// GeofenceType represents different geofence types
const (
	GeofenceTypeCircle        = "circle"
	GeofenceTypePolygon       = "polygon"
	GeofenceTypeRouteCorridor = "route_corridor"
)

// CreateRuleRequest represents the request body for creating an alert rule
type CreateRuleRequest struct {
	Name        string                 `json:"name" validate:"required,max=100"`
	Type        string                 `json:"type" validate:"required,oneof=overspeed geofence_entry geofence_exit harsh_acceleration harsh_braking harsh_cornering idle_time fuel_theft low_battery temperature_excursion accident_detected maintenance_due document_expiry driver_fatigue adas_event vehicle_offline"`
	Description string                 `json:"description" validate:"max=500"`
	Conditions  map[string]interface{} `json:"conditions" validate:"required"`
	Severity    string                 `json:"severity" validate:"required,oneof=info warning critical"`
	Actions     []AlertAction          `json:"actions" validate:"required,min=1"`
	Vehicles    []uuid.UUID            `json:"vehicles"`
	DriverIDs   []uuid.UUID            `json:"driver_ids"`
}

// UpdateRuleRequest represents the request body for updating an alert rule
type UpdateRuleRequest struct {
	Name        *string                 `json:"name,omitempty" validate:"omitempty,max=100"`
	Description *string                 `json:"description,omitempty" validate:"omitempty,max=500"`
	Conditions  *map[string]interface{} `json:"conditions,omitempty"`
	Severity    *string                 `json:"severity,omitempty" validate:"omitempty,oneof=info warning critical"`
	Actions     *[]AlertAction          `json:"actions,omitempty" validate:"omitempty,min=1"`
	Vehicles    *[]uuid.UUID            `json:"vehicles,omitempty"`
	DriverIDs   *[]uuid.UUID            `json:"driver_ids,omitempty"`
	Enabled     *bool                   `json:"enabled,omitempty"`
}


// AlertListFilter represents filter criteria for listing alerts
type AlertListFilter struct {
        OrgID     uuid.UUID  `json:"-"`
        VehicleID *uuid.UUID `json:"vehicle_id,omitempty"`
        DriverID  *uuid.UUID `json:"driver_id,omitempty"`
        Type      []string   `json:"type,omitempty"`
        Status    []string   `json:"status,omitempty"`
        Severity  []string   `json:"severity,omitempty"`
        FromDate  time.Time  `json:"from_date,omitempty"`
        ToDate    time.Time  `json:"to_date,omitempty"`
        Page      int        `json:"page,omitempty"`
        PageSize  int        `json:"page_size,omitempty"`
        OrderBy   string     `json:"order_by,omitempty"`
        OrderDir  string     `json:"order_dir,omitempty"`
}

// AlertSummary represents alert statistics
type AlertSummary struct {
        TotalAlerts   int            `json:"total_alerts"`
        NewAlerts     int            `json:"new_alerts"`
        Acknowledged  int            `json:"acknowledged"`
        Resolved      int            `json:"resolved"`
        Escalated     int            `json:"escalated"`
        ByType        map[string]int `json:"by_type"`
        BySeverity    map[string]int `json:"by_severity"`
        ByStatus      map[string]int `json:"by_status"`
        CriticalAlerts int           `json:"critical_alerts"`
        WarningAlerts  int           `json:"warning_alerts"`
        InfoAlerts     int           `json:"info_alerts"`
}

// AlertTrend represents alert trend data over time
type AlertTrend struct {
        Date          time.Time `json:"date"`
        TotalAlerts   int       `json:"total_alerts"`
        NewAlerts     int       `json:"new_alerts"`
        Resolved      int       `json:"resolved"`
        Critical      int       `json:"critical"`
        Warning       int       `json:"warning"`
        Info          int       `json:"info"`
}

type AlertStats struct {
	TotalAlerts int `json:"total_alerts"`
	NewAlerts   int `json:"new_alerts"`
}
