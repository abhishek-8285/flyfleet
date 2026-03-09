package service

import "errors"

// Platform compatibility service errors
var (
	// General errors
	ErrCompatibilityFailed   = errors.New("compatibility check failed")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrServiceNotFound       = errors.New("service not found")
	ErrServiceUnavailable    = errors.New("service unavailable")
	
	// Route optimization errors
	ErrInvalidRoute          = errors.New("invalid route")
	ErrRouteNotFound         = errors.New("route not found")
	ErrInvalidLocations      = errors.New("invalid locations")
	ErrInsufficientLocations = errors.New("at least two locations are required")
	
	// ML service errors
	ErrMLServiceUnavailable  = errors.New("ML service unavailable")
	ErrInvalidPrediction     = errors.New("invalid prediction request")
	ErrInvalidPricing        = errors.New("invalid pricing request")
	ErrTrainingFailed        = errors.New("training job failed")
	
	// Telemetry errors
	ErrInvalidTelemetry      = errors.New("invalid telemetry data")
	ErrTelemetryIngestFailed = errors.New("telemetry ingestion failed")
	
	// Compliance errors
	ErrInvalidHOSData        = errors.New("invalid HOS data")
	ErrHOSRecordFailed       = errors.New("HOS record failed")
	
	// Rate limiting errors
	ErrRateLimitExceeded     = errors.New("rate limit exceeded")
	
	// Validation errors
	ErrRequiredField         = errors.New("required field is missing")
	ErrInvalidFormat         = errors.New("invalid format")
	ErrInvalidValue          = errors.New("invalid value")
)

// PlatformError represents a platform compatibility error
type PlatformError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *PlatformError) Error() string {
	return e.Message
}

// NewPlatformError creates a new platform error
func NewPlatformError(code, message string) *PlatformError {
	return &PlatformError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *PlatformError {
	return &PlatformError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewRouteError creates a route error
func NewRouteError(message string) *PlatformError {
	return &PlatformError{
		Code:    "ROUTE_ERROR",
		Message: message,
	}
}

// NewMLError creates an ML service error
func NewMLError(message string) *PlatformError {
	return &PlatformError{
		Code:    "ML_ERROR",
		Message: message,
	}
}

// NewTelemetryError creates a telemetry error
func NewTelemetryError(message string) *PlatformError {
	return &PlatformError{
		Code:    "TELEMETRY_ERROR",
		Message: message,
	}
}

// NewComplianceError creates a compliance error
func NewComplianceError(message string) *PlatformError {
	return &PlatformError{
		Code:    "COMPLIANCE_ERROR",
		Message: message,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *PlatformError {
	return &PlatformError{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: message,
	}
}
