package models

import (
	"time"

	"github.com/google/uuid"
)

// Trip represents a trip record
type Trip struct {
	ID             uuid.UUID  `json:"id"`
	TripID         string     `json:"trip_id"`
	VehicleID      uuid.UUID  `json:"vehicle_id"`
	DriverID       *uuid.UUID `json:"driver_id,omitempty"`
	StartLocation  GeoPoint   `json:"start_location"`
	EndLocation    GeoPoint   `json:"end_location"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	Distance       float64    `json:"distance_km"`
	Status         string     `json:"status"` // scheduled, in_progress, completed, cancelled
	FuelConsumed   float64    `json:"fuel_consumed_liters,omitempty"`
	Route          []GeoPoint `json:"route,omitempty"`
	Waypoints      []Waypoint `json:"waypoints,omitempty"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type Waypoint struct {
	GeoPoint
	Type          string     `json:"type"`
	ETA           *time.Time `json:"eta,omitempty"`
	ArrivalTime   *time.Time `json:"arrival_time,omitempty"`
	DepartureTime *time.Time `json:"departure_time,omitempty"`
}

// Telemetry represents a telemetry event
type Telemetry struct {
	ID        uuid.UUID `json:"id"`
	TripID    uuid.UUID `json:"trip_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	EventType string    `json:"event_type"`
	EventTime time.Time `json:"event_time"`
}

// TripSettings represents organization-level trip configuration
type TripSettings struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	RetentionPeriod int                    `json:"retention_period"`
	Policies        map[string]interface{} `json:"policies"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// CreateTripRequest is the DTO for creating a trip
type CreateTripRequest struct {
	VehicleID     uuid.UUID  `json:"vehicle_id" binding:"required"`
	DriverID      *uuid.UUID `json:"driver_id"`
	StartLocation GeoPoint   `json:"start_location" binding:"required"`
	EndLocation   GeoPoint   `json:"end_location" binding:"required"`
	StartTime     time.Time  `json:"start_time" binding:"required"`
}

// UpdateTripRequest is the DTO for updating a trip
type UpdateTripRequest struct {
	DriverID      *uuid.UUID `json:"driver_id"`
	StartLocation *GeoPoint  `json:"start_location"`
	EndLocation   *GeoPoint  `json:"end_location"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Distance      *float64   `json:"distance_km"`
	Status        string     `json:"status"`
}

// TripListParams represents query parameters for listing trips
type TripListParams struct {
	TenantID       uuid.UUID
	Status         string
	VehicleID      *uuid.UUID
	DriverID       *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	Page           int
	Limit          int
}
