package service

import "errors"

// Fuel service errors
var (
	// Fuel log errors
	ErrLogNotFound                 = errors.New("fuel log not found")
	ErrInvalidOdometer             = errors.New("odometer reading cannot be lower than previous entry")
	ErrInvalidCost                 = errors.New("total cost must match quantity multiplied by unit price")
	ErrInvalidFuelLog              = errors.New("invalid fuel log")
	ErrDuplicateFuelLog            = errors.New("duplicate fuel transaction within time window")
	ErrFuelQuantityExceedsCapacity = errors.New("fuel quantity exceeds vehicle tank capacity")
	ErrDailyFuelLimitExceeded      = errors.New("daily fuel limit exceeded for driver")
	ErrMonthlyFuelLimitExceeded    = errors.New("monthly fuel limit exceeded for driver")

	// Status errors
	ErrInvalidStatus      = errors.New("invalid fuel log status")
	ErrStatusNotPending   = errors.New("only pending fuel logs can be approved or rejected")
	ErrLogAlreadyApproved = errors.New("approved fuel logs cannot be modified")
	ErrLogAlreadyRejected = errors.New("rejected fuel logs cannot be modified")
	ErrLogImmutable       = errors.New("approved or rejected fuel logs cannot be modified")

	// Settings errors
	ErrInvalidSettings  = errors.New("invalid fuel settings")
	ErrSettingsNotFound = errors.New("fuel settings not found")
	ErrInvalidFuelType  = errors.New("invalid fuel type")
	ErrInvalidCurrency  = errors.New("invalid currency code")

	// Validation errors
	ErrRequiredField       = errors.New("required field is missing")
	ErrInvalidQuantity     = errors.New("quantity must be greater than zero")
	ErrInvalidPrice        = errors.New("price must be greater than zero")
	ErrInvalidOrganization = errors.New("organization ID is required")
	ErrInvalidVehicle      = errors.New("vehicle ID is required")
	ErrInvalidDriver       = errors.New("driver ID is required")

	// Anomaly errors
	ErrAnomalyDetected   = errors.New("fuel anomaly detected")
	ErrOdometerTampering = errors.New("possible odometer tampering detected")
	ErrFuelTheft         = errors.New("possible fuel theft detected")

	// Integration errors
	ErrUnsupportedProvider = errors.New("integration provider not supported")
)

// FuelError represents a fuel domain error
type FuelError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *FuelError) Error() string {
	return e.Message
}

// NewFuelError creates a new fuel error
func NewFuelError(code, message string) *FuelError {
	return &FuelError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *FuelError {
	return &FuelError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewAnomalyError creates an anomaly detection error
func NewAnomalyError(anomalyType, message string) *FuelError {
	return &FuelError{
		Code:    "ANOMALY_" + anomalyType,
		Message: message,
	}
}
