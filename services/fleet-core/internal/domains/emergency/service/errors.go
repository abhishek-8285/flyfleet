package service

import "errors"

// Emergency service errors
var (
	// SOS Incident errors
	ErrIncidentNotFound      = errors.New("SOS incident not found")
	ErrInvalidIncidentStatus = errors.New("invalid incident status")
	ErrDuplicateSOS          = errors.New("duplicate SOS - incident already triggered recently")
	ErrInvalidSeverity       = errors.New("invalid severity level")
	
	// Status transition errors
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrIncidentAlreadyAcknowledged = errors.New("incident already acknowledged")
	ErrIncidentAlreadyResolved = errors.New("incident already resolved")
	
	// Emergency contact errors
	ErrNoEmergencyContacts = errors.New("no emergency contacts configured")
	ErrContactNotFound     = errors.New("emergency contact not found")
	ErrContactInactive     = errors.New("emergency contact is inactive")
	
	// Validation errors
	ErrInvalidCoordinates   = errors.New("invalid GPS coordinates")
	ErrInvalidLatitude      = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude     = errors.New("longitude must be between -180 and 180")
	ErrMissingVehicleID     = errors.New("vehicle ID is required")
	ErrMissingLocation      = errors.New("location is required")
	
	// Escalation errors
	ErrEscalationFailed    = errors.New("failed to escalate incident")
	ErrNotificationFailed  = errors.New("failed to send notification")
	ErrUnauthorized        = errors.New("unauthorized access to emergency incident")
	ErrAllProvidersFailed  = errors.New("all notification providers failed")

	// Drill errors
	ErrDrillAlreadyActive = errors.New("a drill is already in progress")
)

// EmergencyError represents an emergency domain error
type EmergencyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *EmergencyError) Error() string {
	return e.Message
}

// NewEmergencyError creates a new emergency error
func NewEmergencyError(code, message string) *EmergencyError {
	return &EmergencyError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *EmergencyError {
	return &EmergencyError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewCoordinateError creates a coordinate validation error
func NewCoordinateError(message string) *EmergencyError {
	return &EmergencyError{
		Code:    "COORDINATE_ERROR",
		Message: message,
	}
}
