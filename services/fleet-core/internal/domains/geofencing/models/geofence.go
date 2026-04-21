package models

import (
	"time"

	"github.com/google/uuid"
)

// Geofence represents a geographic boundary for monitoring
type Geofence struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	TenantID         uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name          string                 `json:"name" db:"name"`
	Type          string                 `json:"type" db:"type"`                               // circle, polygon, route_corridor
	Center        *Location              `json:"center,omitempty" db:"center"`                 // For circle geofences
	RadiusMeters  *int                   `json:"radius,omitempty" db:"radius_meters"`          // For circle geofences
	Coordinates   []Location             `json:"coordinates,omitempty" db:"coordinates"`       // For polygon geofences
	RoutePolyline *string                `json:"route_polyline,omitempty" db:"route_polyline"` // For route corridors
	BufferMeters  *int                   `json:"buffer_meters,omitempty" db:"buffer_meters"`   // For route corridors
	EntryAlert    bool                   `json:"entry_alert" db:"entry_alert"`
	ExitAlert     bool                   `json:"exit_alert" db:"exit_alert"`
	SpeedLimit    *int                   `json:"speed_limit,omitempty" db:"speed_limit"`
	Vehicles      []uuid.UUID            `json:"vehicles,omitempty" db:"vehicles"`         // Specific vehicles or empty for all
	ActiveHours   *ActiveHours           `json:"active_hours,omitempty" db:"active_hours"` // When the geofence is active
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Active        bool                   `json:"active" db:"active"`
	CreatedBy     uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

type Location struct {
	Lat float64 `json:"lat" db:"lat"`
	Lng float64 `json:"lng" db:"lng"`
}

type ActiveHours struct {
	Start string `json:"start" db:"start"` // HH:MM format
	End   string `json:"end" db:"end"`     // HH:MM format
	Days  []int  `json:"days" db:"days"`   // 0=Sunday, 1=Monday, etc.
}

// CreateGeofenceRequest represents the request to create a geofence
type CreateGeofenceRequest struct {
	Name          string                 `json:"name" validate:"required,max=100"`
	Type          string                 `json:"type" validate:"required,oneof=circle polygon route_corridor"`
	Center        *Location              `json:"center,omitempty" validate:"omitempty"`                        // Required for circle
	RadiusMeters  *int                   `json:"radius,omitempty" validate:"omitempty,min=10,max=50000"`       // Required for circle
	Coordinates   []Location             `json:"coordinates,omitempty" validate:"omitempty,min=3"`             // Required for polygon (min 3 points)
	RoutePolyline *string                `json:"route_polyline,omitempty" validate:"omitempty"`                // Required for route corridor
	BufferMeters  *int                   `json:"buffer_meters,omitempty" validate:"omitempty,min=10,max=1000"` // Required for route corridor
	EntryAlert    bool                   `json:"entry_alert"`
	ExitAlert     bool                   `json:"exit_alert"`
	SpeedLimit    *int                   `json:"speed_limit,omitempty" validate:"omitempty,min=10,max=200"`
	Vehicles      []uuid.UUID            `json:"vehicles"`
	ActiveHours   *ActiveHours           `json:"active_hours,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateGeofenceRequest represents the request to update a geofence
type UpdateGeofenceRequest struct {
	Name          *string                `json:"name,omitempty" validate:"omitempty,max=100"`
	Type          *string                `json:"type,omitempty" validate:"omitempty,oneof=circle polygon route_corridor"`
	Center        *Location              `json:"center,omitempty" validate:"omitempty"`
	RadiusMeters  *int                   `json:"radius,omitempty" validate:"omitempty,min=10,max=50000"`
	Coordinates   []Location             `json:"coordinates,omitempty" validate:"omitempty,min=3"`
	RoutePolyline *string                `json:"route_polyline,omitempty" validate:"omitempty"`
	BufferMeters  *int                   `json:"buffer_meters,omitempty" validate:"omitempty,min=10,max=1000"`
	EntryAlert    *bool                  `json:"entry_alert,omitempty"`
	ExitAlert     *bool                  `json:"exit_alert,omitempty"`
	SpeedLimit    *int                   `json:"speed_limit,omitempty" validate:"omitempty,min=10,max=200"`
	Vehicles      []uuid.UUID            `json:"vehicles,omitempty"`
	ActiveHours   *ActiveHours           `json:"active_hours,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Active        *bool                  `json:"active,omitempty"`
}

// ListGeofencesRequest represents the request to list geofences
type ListGeofencesRequest struct {
	Page   int    `form:"page,default=1" validate:"gte=1"`
	Limit  int    `form:"limit,default=20" validate:"gte=1,max=100"`
	Search string `form:"search" validate:"omitempty,max=255"`
	Type   string `form:"type" validate:"omitempty,oneof=circle polygon route_corridor"`
	Active *bool  `form:"active"`
	Sort   string `form:"sort,default=created_at:desc"`
}

// GeofenceListResponse represents the response for listing geofences
type GeofenceListResponse struct {
	Geofences []Geofence `json:"geofences"`
	Total     int64      `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
	HasMore   bool       `json:"has_more"`
}

// GeofenceEvent represents a geofence entry/exit/dwell event
type GeofenceEvent struct {
	ID               uuid.UUID `json:"id"`
	GeofenceID       uuid.UUID `json:"geofence_id"`
	GeofenceName     string    `json:"geofence_name"`
	VehicleID        uuid.UUID `json:"vehicle_id"`
	VehicleRegNumber string    `json:"vehicle_registration_number"`
	EventType        string    `json:"event_type"` // ENTRY, EXIT, DWELL
	Location         Location  `json:"location"`
	DurationMinutes  *int      `json:"duration_minutes,omitempty"` // For exit/dwell events
	Timestamp        time.Time `json:"timestamp"`
	CreatedAt        time.Time `json:"created_at"`
}

// GeofenceEventFilter represents the filter for geofence events
type GeofenceEventFilter struct {
	TenantID      uuid.UUID `json:"-" form:"-"`
	GeofenceID uuid.UUID `form:"geofence_id"`
	VehicleID  uuid.UUID `form:"vehicle_id"`
	EventType  string    `form:"event_type" validate:"omitempty,oneof=enter exit dwell speed"`
	From       time.Time `form:"from"`
	To         time.Time `form:"to"`
	Page       int       `form:"page,default=1" validate:"gte=1"`
	Limit      int       `form:"limit,default=20" validate:"gte=1,max=100"`
}
