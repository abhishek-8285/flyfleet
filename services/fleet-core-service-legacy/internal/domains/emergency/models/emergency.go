package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SOSSeverity string
type SOSStatus string
type IncidentType string
type PanicSource string

const (
	SeverityLow      SOSSeverity = "LOW"
	SeverityMedium   SOSSeverity = "MEDIUM"
	SeverityHigh     SOSSeverity = "HIGH"
	SeverityCritical SOSSeverity = "CRITICAL"

	StatusTriggered    SOSStatus = "TRIGGERED"
	StatusAcknowledged SOSStatus = "ACKNOWLEDGED"
	StatusResolved     SOSStatus = "RESOLVED"
	StatusFalseAlarm   SOSStatus = "FALSE_ALARM"

	TypePanic            IncidentType = "SOS_PANIC"
	TypeVehicleAccident  IncidentType = "VEHICLE_ACCIDENT"
	TypeVehicleBreakdown IncidentType = "VEHICLE_BREAKDOWN"
	TypeMedicalEmergency IncidentType = "MEDICAL_EMERGENCY"
	TypeSecurityThreat   IncidentType = "SECURITY_THREAT"
	TypeFireAlert        IncidentType = "FIRE_ALERT"
	TypeDriverAssault    IncidentType = "DRIVER_ASSAULT"
	TypeVehicleTheft     IncidentType = "VEHICLE_THEFT"

	SourceAIS140    PanicSource = "AIS140_DEVICE"
	SourceMobile     PanicSource = "DRIVER_MOBILE"
	SourceTelematics PanicSource = "VEHICLE_TELEMATICS"
	SourceDispatcher PanicSource = "DISPATCHER"
)

// SOSIncident represents an emergency event
type SOSIncident struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	VehicleID       uuid.UUID       `db:"vehicle_id" json:"vehicle_id"`
	DriverID        *uuid.UUID      `db:"driver_id" json:"driver_id,omitempty"`
	TripID          *uuid.UUID      `db:"trip_id" json:"trip_id,omitempty"`
	DeviceID        *uuid.UUID      `db:"device_id" json:"device_id,omitempty"`
	Timestamp       time.Time       `db:"timestamp" json:"timestamp"`
	Latitude        float64         `db:"latitude" json:"latitude"`
	Longitude       float64         `db:"longitude" json:"longitude"`
	LocationAddress *string         `db:"location_address" json:"location_address,omitempty"`
	Status          SOSStatus       `db:"status" json:"status"`
	Severity        SOSSeverity     `db:"severity" json:"severity"`
	IncidentType    IncidentType    `db:"incident_type" json:"incident_type"`
	PanicSource     PanicSource     `db:"panic_source" json:"panic_source"`
	Speed           int             `db:"speed" json:"speed"`
	Heading         int             `db:"heading" json:"heading"`
	Metadata        json.RawMessage `db:"metadata" json:"metadata"`
	AcknowledgedBy  *uuid.UUID      `db:"acknowledged_by" json:"acknowledged_by,omitempty"`
	ResolvedBy      *uuid.UUID      `db:"resolved_by" json:"resolved_by,omitempty"`
	OrgID           uuid.UUID       `db:"org_id" json:"org_id"`
	HighRiskLocation bool           `db:"high_risk_location" json:"high_risk_location"`
	SimulationMode   bool           `db:"simulation_mode" json:"simulation_mode"`
	DriverFeedback   *string        `db:"driver_feedback" json:"driver_feedback,omitempty"`
	FeedbackNotes    *string        `db:"feedback_notes" json:"feedback_notes,omitempty"`
	AcknowledgedAt   *time.Time     `db:"acknowledged_at" json:"acknowledged_at,omitempty"`
	ResolvedAt       *time.Time     `db:"resolved_at" json:"resolved_at,omitempty"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`

	// Join fields
	VehicleRegNumber *string `db:"vehicle_reg_number" json:"vehicle_reg_number,omitempty"`
	DriverName       *string `db:"driver_name" json:"driver_name,omitempty"`
}

// IncidentMedia represents an evidence file attached to an incident
type IncidentMedia struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	IncidentID uuid.UUID  `db:"incident_id" json:"incident_id"`
	MediaType  string     `db:"media_type" json:"media_type"` // IMAGE, VIDEO, AUDIO
	FileURL    string     `db:"file_url" json:"file_url"`
	UploadedBy *uuid.UUID `db:"uploaded_by" json:"uploaded_by,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

// IncidentVehicleSnapshot represents the telemetry state at the time of incident
type IncidentVehicleSnapshot struct {
	IncidentID           uuid.UUID `db:"incident_id" json:"incident_id"`
	Speed                int       `db:"speed" json:"speed"`
	Heading              int       `db:"heading" json:"heading"`
	EngineStatus         bool      `db:"engine_status" json:"engine_status"`
	FuelLevel            int       `db:"fuel_level" json:"fuel_level"`
	BatteryVoltage       float64   `db:"battery_voltage" json:"battery_voltage"`
	GPSAccuracy          float64   `db:"gps_accuracy" json:"gps_accuracy"`
	DeviceSignalStrength int       `db:"device_signal_strength" json:"device_signal_strength"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
}

// EmergencyEscalation tracks the notification chain
type EmergencyEscalation struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	IncidentID          uuid.UUID  `db:"incident_id" json:"incident_id"`
	ContactID           uuid.UUID  `db:"contact_id" json:"contact_id"`
	NotificationChannel string     `db:"notification_channel" json:"notification_channel"`
	Status              string     `db:"status" json:"status"`
	AttemptedAt         time.Time  `db:"attempted_at" json:"attempted_at"`
	RespondedAt         *time.Time `db:"responded_at" json:"responded_at,omitempty"`
}

// EmergencyResponseUnit for Indian emergency services
type EmergencyResponseUnit struct {
	ID           uuid.UUID `db:"id" json:"id"`
	OrgID        uuid.UUID `db:"org_id" json:"org_id"`
	UnitType     string    `db:"unit_type" json:"unit_type"`
	Name         string    `db:"name" json:"name"`
	Phone        string    `db:"phone" json:"phone"`
	CoverageCity string    `db:"coverage_city" json:"coverage_city"`
	CoverageState string    `db:"coverage_state" json:"coverage_state"`
	ContactNumber string   `db:"contact_number" json:"contact_number"` // Expected by service
	Type         string    `db:"type" json:"type"`           // Expected by service
	IsActive     bool      `db:"is_active" json:"is_active"`
}

// EmergencyVehicleState for persistence
type EmergencyVehicleState struct {
	VehicleID uuid.UUID `db:"vehicle_id" json:"vehicle_id"`
	OrgID     uuid.UUID `db:"org_id" json:"org_id"`
	Halted    bool      `db:"halted" json:"halted"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreateSOSRequest updated for AIS-140
type CreateSOSRequest struct {
	VehicleID       string          `json:"vehicle_id" validate:"required,uuid"`
	OrgID           uuid.UUID       `json:"org_id"` // Standardized
	DeviceID        string          `json:"device_id" validate:"omitempty,uuid"`
	DriverID        string          `json:"driver_id" validate:"omitempty,uuid"`
	TripID          string          `json:"trip_id" validate:"omitempty,uuid"`
	Latitude        float64         `json:"latitude" validate:"required"`
	Longitude       float64         `json:"longitude" validate:"required"`
	LocationAddress string          `json:"location_address"`
	IncidentType    IncidentType    `json:"incident_type" validate:"omitempty,oneof=SOS_PANIC VEHICLE_ACCIDENT VEHICLE_BREAKDOWN MEDICAL_EMERGENCY SECURITY_THREAT FIRE_ALERT DRIVER_ASSAULT VEHICLE_THEFT"`
	PanicSource     PanicSource     `json:"panic_source"`
	Severity        SOSSeverity     `json:"severity"`
	Speed           int             `json:"speed"`
	Heading         int             `json:"heading"`
	Metadata        json.RawMessage `json:"metadata"`
	SimulationMode   bool            `json:"simulation_mode"`
}

// UpdateIncidentRequest payload for acknowledging/resolving
type UpdateIncidentRequest struct {
	Status string `json:"status" validate:"required,oneof=ACKNOWLEDGED RESOLVED FALSE_ALARM"`
	Notes  string `json:"notes"`
}

// EmergencyContact represents a point of contact for escalation
type EmergencyContact struct {
	ID        uuid.UUID `db:"id" json:"id"`
	OrgID     uuid.UUID `db:"org_id" json:"org_id"`
	Name      string    `db:"name" json:"name"`
	Role      string    `db:"role" json:"role"`
	Phone     string    `db:"phone" json:"phone"`
	Email     string    `db:"email" json:"email"`
	Priority  int       `db:"priority" json:"priority"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// IncidentLog represents an audit trail entry
type IncidentLog struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	IncidentID uuid.UUID  `db:"incident_id" json:"incident_id"`
	UserID     *uuid.UUID `db:"user_id" json:"user_id,omitempty"` // Nullable for system actions
	Action     string     `db:"action" json:"action"`
	Notes      string     `db:"notes" json:"notes"`
	Timestamp  time.Time  `db:"timestamp" json:"timestamp"`
}

// SLAViolation represents an SLA breach record
type SLAViolation struct {
	ID         uuid.UUID   `db:"id" json:"id"`
	IncidentID uuid.UUID   `db:"incident_id" json:"incident_id"`
	Severity   SOSSeverity `db:"severity" json:"severity"`
	BreachType string      `db:"breach_type" json:"breach_type"`
	BreachTime time.Time   `db:"breach_time" json:"breach_time"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
}

// IncidentTimeline represents an entry in the incident history
type IncidentTimeline struct {
	ID         uuid.UUID       `json:"id"`
	IncidentID uuid.UUID       `json:"incident_id"`
	EventType  string          `json:"event_type"` // e.g., "TRIGGERED", "ACKNOWLEDGED", "ESCALATED", "AUTHORITY_NOTIFIED"
	Actor      string          `json:"actor"`      // e.g., "SYSTEM", "USER_ID", "ERU_POLICE"
	Metadata   json.RawMessage `json:"metadata"`
	CreatedAt  time.Time       `json:"created_at"`
}

// EscalationState represents the state of an escalation attempt
type EscalationState string

const (
	EscalationPending   EscalationState = "PENDING"
	EscalationNotified  EscalationState = "NOTIFIED"
	EscalationResponded EscalationState = "RESPONDED"
	EscalationFailed    EscalationState = "FAILED"
	EscalationEscalated EscalationState = "ESCALATED"
)

// IncidentAuditHistory represents complete audit history for an incident
type IncidentAuditHistory struct {
	IncidentID uuid.UUID          `json:"incident_id"`
	Logs       []IncidentLog      `json:"logs"`
	Timeline   []IncidentTimeline `json:"timeline"`
}

// DailyIncidentCount for analytics
type DailyIncidentCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// IncidentTypeCount for analytics
type IncidentTypeCount struct {
	Type  IncidentType `json:"type"`
	Count int          `json:"count"`
}

// IncidentAnalyticsResponse for API
type IncidentAnalyticsResponse struct {
	IncidentsToday       int                 `json:"incidents_today"`
	IncidentsThisWeek    int                 `json:"incidents_this_week"`
	IncidentsByType      []IncidentTypeCount `json:"incidents_by_type"`
	AverageResponseTime  float64             `json:"average_response_time"`
	AverageResolutionTime float64            `json:"average_resolution_time"`
}

// VehicleSafetyScore based on incident history
type VehicleSafetyScore struct {
	VehicleID        uuid.UUID `json:"vehicle_id"`
	VehicleRegNumber string    `json:"vehicle_reg_number"`
	IncidentCount    int       `json:"incident_count"`
	SafetyScore      float64   `json:"safety_score"` // 0-100
}
