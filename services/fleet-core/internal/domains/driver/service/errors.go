package service

import "errors"

// Driver service errors
var (
	// Driver errors
	ErrDriverNotFound          = errors.New("driver not found")
	ErrDriverAlreadyExists     = errors.New("driver with this email or license already exists")
	ErrInvalidDriverStatus     = errors.New("invalid driver status")
	ErrDriverHasActiveVehicle  = errors.New("driver has active vehicle assignment")
	ErrDriverHasActiveTrip     = errors.New("driver has active trip")
	ErrDriverNotActive         = errors.New("driver is not active")
	
	// License errors
	ErrLicenseExpired          = errors.New("driver license has expired")
	ErrLicenseInvalid          = errors.New("invalid driver license")
	ErrLicenseExpiringSoon     = errors.New("driver license expiring soon")
	
	// Behavior score errors
	ErrInvalidScoreChange      = errors.New("score change must be between -20 and 20")
	ErrScoreOutOfBounds        = errors.New("behavior score must be between 0 and 100")
	
	// Document errors
	ErrDocumentNotFound        = errors.New("document not found")
	ErrDocumentExpired         = errors.New("document has expired")
	ErrDocumentExpiringSoon    = errors.New("document expiring soon")
	ErrDocumentNotVerified     = errors.New("document not verified")
	
	// Duty status errors
	ErrInvalidDutyStatus       = errors.New("invalid duty status")
	ErrInvalidStatusTransition = errors.New("invalid duty status transition")
	ErrDutyLogNotFound         = errors.New("duty status log not found")
	
	// Message errors
	ErrMessageNotFound         = errors.New("message not found")
	ErrInvalidMessageDirection = errors.New("invalid message direction")
	ErrInvalidMessageType      = errors.New("invalid message type")
	
	// Expense errors
	ErrExpenseNotFound         = errors.New("expense not found")
	ErrInvalidExpenseType      = errors.New("invalid expense type")
	ErrInvalidExpenseAmount    = errors.New("expense amount must be positive")
	ErrExpenseAlreadyProcessed = errors.New("expense already processed")
	
	// Validation errors
	ErrRequiredField           = errors.New("required field is missing")
	ErrInvalidFormat           = errors.New("invalid field format")
	ErrInvalidEmail            = errors.New("invalid email address")
	ErrInvalidPhone            = errors.New("invalid phone number")
	ErrInvalidUUID             = errors.New("invalid UUID format")
)

// DriverError represents a driver domain error
type DriverError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *DriverError) Error() string {
	return e.Message
}

// NewDriverError creates a new driver error
func NewDriverError(code, message string) *DriverError {
	return &DriverError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *DriverError {
	return &DriverError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewLicenseError creates a license-related error
func NewLicenseError(message string) *DriverError {
	return &DriverError{
		Code:    "LICENSE_ERROR",
		Message: message,
	}
}

// NewBehaviorError creates a behavior score error
func NewBehaviorError(message string) *DriverError {
	return &DriverError{
		Code:    "BEHAVIOR_ERROR",
		Message: message,
	}
}
