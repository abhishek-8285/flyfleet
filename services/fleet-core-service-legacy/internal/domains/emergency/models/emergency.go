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
	ID            uuid.UUID `db:"id" json:"id"`
	OrgID         uuid.UUID `db:"org_id" json:"org_id"`
	UnitType      string    `db:"unit_type" json:"unit_type"`
	Name          string    `db:"name" json:"name"`
	Phone         string    `db:"phone" json:"phone"`
	CoverageCity  string    `db:"coverage_city" json:"coverage_city"`
	CoverageState string    `db:"coverage_state" json:"coverage_state"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	// New Phase 3/4 Fields
	Type          string    `db:"authority_type" json:"authority_type"`
	ContactNumber string    `db:"contact_number" json:"contact_number"`
	APIEndpoint   string    `db:"api_endpoint" json:"api_endpoint"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
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

// EmergencyContactResponse tracks individual contact interaction with an incident
type EmergencyContactResponse struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	IncidentID     uuid.UUID  `db:"incident_id" json:"incident_id"`
	ContactID      uuid.UUID  `db:"contact_id" json:"contact_id"`
	ResponseStatus string     `db:"response_status" json:"response_status"` // NOTIFIED, ACKNOWLEDGED, DECLINED, FAILED
	RespondedAt    *time.Time `db:"responded_at" json:"responded_at,omitempty"`
	Notes          string     `db:"notes" json:"notes"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

// IncidentAnalyticsResponse for API (updated for enterprise requirements)
type IncidentAnalyticsResponse struct {
	TotalIncidents          int                 `json:"total_incidents"`
	AvgResponseTimeSeconds  float64             `json:"avg_response_time_seconds"`
	IncidentsLast24h        int                 `json:"incidents_last_24h"`
	TopLocations            []string            `json:"top_locations"`
	IncidentsByType         []IncidentTypeCount `json:"incidents_by_type"`
	IncidentsToday          int                 `json:"incidents_today"`           // Backward compatibility
	IncidentsThisWeek       int                 `json:"incidents_this_week"`      // Backward compatibility
	AverageResponseTime     float64             `json:"average_response_time_min"` // Backward compatibility
	AverageResolutionTime   float64             `json:"average_resolution_time_min"`
}

// VehicleSafetyScore based on incident history
type VehicleSafetyScore struct {
	VehicleID        uuid.UUID `json:"vehicle_id"`
	VehicleRegNumber string    `json:"vehicle_reg_number"`
	IncidentCount    int       `json:"incident_count"`
	SafetyScore      float64   `json:"safety_score"` // 0-100
}

// DeviceEventQueueItem represents an SOS event buffered when the device is offline.
type DeviceEventQueueItem struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	VehicleID   uuid.UUID       `db:"vehicle_id" json:"vehicle_id"`
	EventType   string          `db:"event_type" json:"event_type"`
	Payload     json.RawMessage `db:"payload" json:"payload"`
	RetryCount  int             `db:"retry_count" json:"retry_count"`
	Status      string          `db:"status" json:"status"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	NextRetryAt time.Time       `db:"next_retry_at" json:"next_retry_at"`
}

// EmergencyDrillRequest is the request body for the drill simulation endpoint.
type EmergencyDrillRequest struct {
	OrgID       string  `json:"org_id" validate:"required,uuid"`
	VehicleID   string  `json:"vehicle_id" validate:"required,uuid"`
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	IncidentType string  `json:"incident_type,omitempty"`
}

// AuthorityDeliveryLog tracks attempts to notify emergency authorities.
type AuthorityDeliveryLog struct {
	ID             uuid.UUID `db:"id" json:"id"`
	IncidentID     uuid.UUID `db:"incident_id" json:"incident_id"`
	AuthorityType  string    `db:"authority_type" json:"authority_type"`
	DeliveryStatus string    `db:"delivery_status" json:"delivery_status"`
	ResponseCode   int       `db:"response_code" json:"response_code"`
	ResponseBody   string    `db:"response_body" json:"response_body"`
	RetryCount     int       `db:"retry_count" json:"retry_count"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// TimelineEvent represents a fine-grained audit event for an incident.
type TimelineEvent struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	IncidentID uuid.UUID       `db:"incident_id" json:"incident_id"`
	EventType  string          `db:"event_type" json:"event_type"`
	EventData  json.RawMessage `db:"event_data" json:"event_data"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
}
// IncidentSnapshot represents a backup record of a critical incident.
type IncidentSnapshot struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	IncidentID   uuid.UUID       `db:"incident_id" json:"incident_id"`
	VehicleID    uuid.UUID       `db:"vehicle_id" json:"vehicle_id"`
	OrgID        uuid.UUID       `db:"org_id" json:"org_id"`
	IncidentType IncidentType    `db:"incident_type" json:"incident_type"`
	Severity     SOSSeverity     `db:"severity" json:"severity"`
	Status       SOSStatus       `db:"status" json:"status"`
	Latitude     float64         `db:"latitude" json:"latitude"`
	Longitude    float64         `db:"longitude" json:"longitude"`
	SnapshotData json.RawMessage `db:"snapshot_data" json:"snapshot_data"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
}
