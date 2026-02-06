package schemas

import (
	"time"

	"github.com/google/uuid"
)

// Maintenance Events
type MaintenanceScheduleCreated struct {
	EventID     string    `json:"event_id"`
	ScheduleID  string    `json:"schedule_id"`
	OrgID       uuid.UUID `json:"org_id"`
	VehicleID   uuid.UUID `json:"vehicle_id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	DueDate     time.Time `json:"due_date"`
	Priority    string    `json:"priority"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Timestamp   time.Time `json:"timestamp"`
}

type MaintenanceScheduleUpdated struct {
	EventID    string                 `json:"event_id"`
	ScheduleID string                 `json:"schedule_id"`
	OrgID      uuid.UUID              `json:"org_id"`
	Updates    map[string]interface{} `json:"updates"`
	UpdatedBy  uuid.UUID              `json:"updated_by"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Timestamp  time.Time              `json:"timestamp"`
}

type MaintenanceLogCreated struct {
	EventID   string    `json:"event_id"`
	LogID     string    `json:"log_id"`
	OrgID     uuid.UUID `json:"org_id"`
	VehicleID uuid.UUID `json:"vehicle_id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Cost      float64   `json:"cost"`
	PerformedBy uuid.UUID `json:"performed_by"`
	CreatedAt time.Time `json:"created_at"`
	Timestamp time.Time `json:"timestamp"`
}

// Vehicle Events
type VehicleLocationUpdated struct {
	EventID   string    `json:"event_id"`
	VehicleID uuid.UUID `json:"vehicle_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}

type VehicleStatusChanged struct {
	EventID   string    `json:"event_id"`
	VehicleID uuid.UUID `json:"vehicle_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// Trip Events
type TripStarted struct {
	EventID   string    `json:"event_id"`
	TripID    string    `json:"trip_id"`
	VehicleID uuid.UUID `json:"vehicle_id"`
	DriverID  uuid.UUID `json:"driver_id"`
	StartTime time.Time `json:"start_time"`
	StartLocation GeoPoint `json:"start_location"`
	Timestamp time.Time `json:"timestamp"`
}

type TripCompleted struct {
	EventID     string    `json:"event_id"`
	TripID      string    `json:"trip_id"`
	VehicleID   uuid.UUID `json:"vehicle_id"`
	DriverID    uuid.UUID `json:"driver_id"`
	EndTime     time.Time `json:"end_time"`
	EndLocation GeoPoint  `json:"end_location"`
	Distance    float64   `json:"distance"`
	Duration    int       `json:"duration"`
	Timestamp   time.Time `json:"timestamp"`
}

// Alert Events
type MaintenanceAlertTriggered struct {
	EventID     string    `json:"event_id"`
	AlertID     string    `json:"alert_id"`
	ScheduleID  string    `json:"schedule_id"`
	VehicleID   uuid.UUID `json:"vehicle_id"`
	AlertType   string    `json:"alert_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

type GeofenceBreachDetected struct {
	EventID     string    `json:"event_id"`
	VehicleID   uuid.UUID `json:"vehicle_id"`
	GeofenceID  string    `json:"geofence_id"`
	BreachType  string    `json:"breach_type"` // enter, exit
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Timestamp   time.Time `json:"timestamp"`
}

// Common Types
type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Event Metadata
type EventMetadata struct {
	EventID     string            `json:"event_id"`
	EventType   string            `json:"event_type"`
	Source      string            `json:"source"`
	Version     string            `json:"version"`
	Timestamp   time.Time         `json:"timestamp"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Event Names (constants)
const (
	MaintenanceScheduleCreatedEvent  = "maintenance.schedule.created"
	MaintenanceScheduleUpdatedEvent  = "maintenance.schedule.updated"
	MaintenanceLogCreatedEvent       = "maintenance.log.created"
	VehicleLocationUpdatedEvent      = "vehicle.location.updated"
	VehicleStatusChangedEvent        = "vehicle.status.changed"
	TripStartedEvent                 = "trip.started"
	TripCompletedEvent               = "trip.completed"
	MaintenanceAlertTriggeredEvent   = "alert.maintenance.triggered"
	GeofenceBreachDetectedEvent      = "geofence.breach.detected"
)