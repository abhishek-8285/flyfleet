package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// GeofenceType defines the shape of a geofence boundary.
const (
	GeofenceCircle  = "circle"
	GeofencePolygon = "polygon"
)

// Geofence represents a virtual geographic boundary for an organization.
type Geofence struct {
	ID           uuid.UUID  `json:"id"`
	OrgID        uuid.UUID  `json:"org_id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Type         string     `json:"type"` // circle | polygon
	CenterLat    *float64   `json:"center_lat,omitempty"`
	CenterLng    *float64   `json:"center_lng,omitempty"`
	RadiusM      *float64   `json:"radius_m,omitempty"`
	PolygonWKT   string     `json:"polygon_wkt,omitempty"`
	Status       string     `json:"status"` // active | inactive
	AlertChannels []string  `json:"alert_channels"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// GeofenceEventType enumerates possible boundary-crossing events.
const (
	EventTypeEntry = "entry"
	EventTypeExit  = "exit"
)

// GeofenceEvent is logged when a vehicle enters or exits a geofence.
type GeofenceEvent struct {
	ID           uuid.UUID `json:"id"`
	GeofenceID   uuid.UUID `json:"geofence_id"`
	VehicleID    uuid.UUID `json:"vehicle_id"`
	EventType    string    `json:"event_type"` // entry | exit
	OccurredAt   time.Time `json:"occurred_at"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	AlertSent    bool      `json:"alert_sent"`
	AlertChannels []string `json:"alert_channels"`
	CreatedAt    time.Time `json:"created_at"`
}
