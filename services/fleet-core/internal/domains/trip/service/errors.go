package service

import "errors"

// Trip service errors
var (
	// General errors
	ErrTripNotFound      = errors.New("trip not found")
	ErrTripAlreadyExists = errors.New("trip already exists")
	ErrInvalidTripID     = errors.New("invalid trip ID")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInvalidTenantID      = errors.New("invalid organization ID")

	// Trip lifecycle errors
	ErrInvalidTripStatus     = errors.New("invalid trip status")
	ErrTripCannotBeStarted   = errors.New("trip cannot be started from current status")
	ErrTripMustBeInProgress  = errors.New("trip must be in progress")
	ErrTripMustBePaused      = errors.New("trip must be paused")
	ErrTripAlreadyCompleted  = errors.New("trip already completed")
	ErrTripAlreadyCancelled  = errors.New("trip already cancelled")
	ErrTripCannotBeCancelled = errors.New("trip cannot be cancelled")
	ErrTripCannotBeDeleted   = errors.New("trip cannot be deleted")
	ErrTripCannotBeUpdated   = errors.New("trip cannot be updated in current status")

	// Authorization errors
	ErrDriverNotAssigned   = errors.New("driver not assigned to this trip")
	ErrVehicleNotAvailable = errors.New("vehicle not available")
	ErrDriverNotAvailable  = errors.New("driver not available")

	// Validation errors
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidFieldValue    = errors.New("invalid field value")
	ErrInvalidRoute         = errors.New("invalid route: minimum 2 waypoints required")
	ErrInvalidCargoDetails  = errors.New("invalid cargo details")
	ErrInvalidCustomerInfo  = errors.New("invalid customer information")
	ErrInvalidCoordinates   = errors.New("invalid geographic coordinates")

	// POD errors
	ErrPODRequired           = errors.New("POD documents required for trip completion")
	ErrPODUploadFailed       = errors.New("POD upload failed")
	ErrPODVerificationFailed = errors.New("POD verification failed")
	ErrInvalidPODType        = errors.New("invalid POD type")

	// Event errors
	ErrInvalidEventType    = errors.New("invalid event type")
	ErrEventCreationFailed = errors.New("failed to create event")

	// Offer errors
	ErrOfferNotFound         = errors.New("offer not found")
	ErrOfferExpired          = errors.New("offer has expired")
	ErrOfferAlreadyResponded = errors.New("offer already responded to")
	ErrInvalidOfferResponse  = errors.New("invalid offer response")
	ErrTripCannotBeOffered   = errors.New("trip cannot be offered in current status")

	// Analytics errors
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrAnalyticsNotFound = errors.New("analytics data not found")
)

// TripError represents a trip domain error with additional context
type TripError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Field      string `json:"field,omitempty"`
	TripID     string `json:"trip_id,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *TripError) Error() string {
	return e.Message
}

// NewTripError creates a new trip error
func NewTripError(code, message string) *TripError {
	return &TripError{
		Code:    code,
		Message: message,
	}
}

// NewTripErrorWithField creates a new trip error with field context
func NewTripErrorWithField(code, message, field string) *TripError {
	return &TripError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// NewTripErrorWithTripID creates a new trip error with trip ID context
func NewTripErrorWithTripID(code, message, tripID string) *TripError {
	return &TripError{
		Code:    code,
		Message: message,
		TripID:  tripID,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *TripError {
	return &TripError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		Field:      field,
		StatusCode: 400,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, id string) *TripError {
	return &TripError{
		Code:       "NOT_FOUND",
		Message:    resource + " not found: " + id,
		StatusCode: 404,
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *TripError {
	return &TripError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: 403,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *TripError {
	return &TripError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: 409,
	}
}
