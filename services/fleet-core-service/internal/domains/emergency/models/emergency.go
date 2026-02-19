package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SOSSeverity string
type SOSStatus string

const (
	SeverityLow      SOSSeverity = "LOW"
	SeverityMedium   SOSSeverity = "MEDIUM"
	SeverityCritical SOSSeverity = "CRITICAL"

	StatusTriggered      SOSStatus = "TRIGGERED"
	StatusAcknowledged   SOSStatus = "ACKNOWLEDGED"
	StatusResolved       SOSStatus = "RESOLVED"
	StatusFalseAlarm     SOSStatus = "FALSE_ALARM"
)

// SOSIncident represents an emergency event
type SOSIncident struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	VehicleID       uuid.UUID       `db:"vehicle_id" json:"vehicle_id"`
	DriverID        *uuid.UUID      `db:"driver_id" json:"driver_id,omitempty"` // Nullable
	TripID          *uuid.UUID      `db:"trip_id" json:"trip_id,omitempty"`     // Nullable
	Timestamp       time.Time       `db:"timestamp" json:"timestamp"`
	Latitude        float64         `db:"latitude" json:"latitude"`
	Longitude       float64         `db:"longitude" json:"longitude"`
	LocationAddress string          `db:"location_address" json:"location_address"`
	Status          SOSStatus       `db:"status" json:"status"`
	Severity        SOSSeverity     `db:"severity" json:"severity"`
	Metadata        json.RawMessage `db:"metadata" json:"metadata"` // Stores device info, battery, etc.
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`

	// Join fields
	VehicleRegNumber string `db:"vehicle_reg_number" json:"vehicle_reg_number,omitempty"`
	DriverName       string `db:"driver_name" json:"driver_name,omitempty"`
}

// EmergencyContact represents a point of contact for escalation
type EmergencyContact struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	Role           string    `db:"role" json:"role"`
	Phone          string    `db:"phone" json:"phone"`
	Email          string    `db:"email" json:"email"`
	Priority       int       `db:"priority" json:"priority"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// IncidentLog represents an audit trail entry
type IncidentLog struct {
	ID         uuid.UUID `db:"id" json:"id"`
	IncidentID uuid.UUID `db:"incident_id" json:"incident_id"`
	UserID     *uuid.UUID `db:"user_id" json:"user_id,omitempty"` // Nullable for system actions
	Action     string    `db:"action" json:"action"`
	Notes      string    `db:"notes" json:"notes"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"`
}

// CreateSOSRequest payload for triggering an SOS
type CreateSOSRequest struct {
	VehicleID       string          `json:"vehicle_id" validate:"required,uuid"`
	DriverID        string          `json:"driver_id" validate:"omitempty,uuid"`
	TripID          string          `json:"trip_id" validate:"omitempty,uuid"`
	Latitude        float64         `json:"latitude" validate:"required"`
	Longitude       float64         `json:"longitude" validate:"required"`
	LocationAddress string          `json:"location_address"`
	Severity        SOSSeverity     `json:"severity" validate:"omitempty,oneof=LOW MEDIUM CRITICAL"`
	Metadata        json.RawMessage `json:"metadata"`
}

// UpdateIncidentRequest payload for acknowledging/resolving
type UpdateIncidentRequest struct {
	Status string `json:"status" validate:"required,oneof=ACKNOWLEDGED RESOLVED FALSE_ALARM"`
	Notes  string `json:"notes"`
}
