package service

import "errors"

// Customer service errors
var (
	// Authentication errors
	ErrInvalidAPIKey      = errors.New("invalid API key")
	ErrCustomerInactive   = errors.New("customer account is not active")
	ErrInvalidToken       = errors.New("invalid or expired token")
	
	// Customer errors
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrCustomerExists     = errors.New("customer with this email already exists")
	ErrInvalidCustomerID  = errors.New("invalid customer ID")
	
	// Shipment errors
	ErrShipmentNotFound   = errors.New("shipment not found")
	ErrInvalidTracking    = errors.New("invalid tracking number")
	ErrShipmentNotLinked  = errors.New("shipment not linked to trip")
	
	// Booking errors
	ErrInvalidBooking     = errors.New("invalid booking request")
	ErrBookingNotFound    = errors.New("booking request not found")
	ErrInvalidRoute       = errors.New("origin and destination cannot be the same")
	ErrInvalidPickupTime  = errors.New("requested pickup time must be in the future")
	
	// Invoice errors
	ErrInvoiceNotFound    = errors.New("invoice not found")
	ErrInvoiceNotPaid     = errors.New("invoice is not paid")
	
	// Status transition errors
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	
	// Validation errors
	ErrRequiredField    = errors.New("required field is missing")
	ErrInvalidFormat    = errors.New("invalid field format")
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrInvalidPhone     = errors.New("invalid phone number")
)

// CustomerError represents a customer domain error
type CustomerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *CustomerError) Error() string {
	return e.Message
}

// NewCustomerError creates a new customer error
func NewCustomerError(code, message string) *CustomerError {
	return &CustomerError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *CustomerError {
	return &CustomerError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}
