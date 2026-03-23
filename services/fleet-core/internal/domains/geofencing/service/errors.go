package service

import "errors"

// Geofencing service errors
var (
	// Geofence errors
	ErrGeofenceNotFound       = errors.New("geofence not found")
	ErrInvalidGeofence        = errors.New("invalid geofence")
	ErrInvalidGeofenceType    = errors.New("invalid geofence type")
	ErrInvalidGeofenceName    = errors.New("geofence name is required")
	ErrInvalidRadius          = errors.New("radius must be between 10 and 50000 meters")
	ErrInvalidBuffer          = errors.New("buffer must be between 10 and 1000 meters")
	ErrInvalidSpeedLimit      = errors.New("speed limit must be between 10 and 200 km/h")
	ErrInvalidCoordinates     = errors.New("invalid coordinates")
	ErrInvalidPolygon         = errors.New("polygon requires at least 3 coordinates")
	ErrMissingCenter          = errors.New("center is required for circle geofence")
	ErrMissingRadius          = errors.New("radius_meters is required for circle geofence")
	ErrMissingPolygon         = errors.New("coordinates are required for polygon geofence")
	ErrMissingPolyline        = errors.New("route_polyline is required for route corridor geofence")
	ErrMissingBuffer          = errors.New("buffer_meters is required for route corridor geofence")
	
	// Assignment errors
	ErrInvalidAssignmentType  = errors.New("invalid assignment type")
	ErrGeofenceNotActive      = errors.New("geofence is not active")
	
	// Event errors
	ErrInvalidEventType       = errors.New("invalid event type")
	ErrDuplicateEvent         = errors.New("duplicate geofence event")
	
	// Validation errors
	ErrRequiredField          = errors.New("required field is missing")
	ErrInvalidOrganization    = errors.New("organization ID is required")
	ErrInvalidVehicle         = errors.New("vehicle ID is required")
)

// GeofenceError represents a geofencing domain error
type GeofenceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *GeofenceError) Error() string {
	return e.Message
}

// NewGeofenceError creates a new geofence error
func NewGeofenceError(code, message string) *GeofenceError {
	return &GeofenceError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *GeofenceError {
	return &GeofenceError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewGeofenceValidationError creates a geofence-specific validation error
func NewGeofenceValidationError(message string) *GeofenceError {
	return &GeofenceError{
		Code:    "GEOFENCE_VALIDATION_ERROR",
		Message: message,
	}
}
