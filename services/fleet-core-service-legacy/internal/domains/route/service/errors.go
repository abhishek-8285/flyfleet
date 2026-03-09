package service

import "errors"

// Route optimization service errors
var (
	// General errors
	ErrRouteNotFound          = errors.New("route not found")
	ErrInvalidRoute           = errors.New("invalid route")
	ErrInvalidRequest         = errors.New("invalid optimization request")
	ErrInvalidResult          = errors.New("invalid optimization result")
	
	// Request errors
	ErrInvalidOrganization    = errors.New("invalid organization ID")
	ErrInvalidStops           = errors.New("invalid stops")
	ErrInvalidVehicles        = errors.New("invalid vehicles")
	ErrInvalidCoordinates     = errors.New("invalid coordinates")
	ErrInvalidLocation        = errors.New("invalid location")
	
	// Optimization errors
	ErrOptimizationFailed     = errors.New("optimization failed")
	ErrNoVehiclesAvailable    = errors.New("no vehicles available")
	ErrNoStopsProvided        = errors.New("no stops provided")
	ErrInfeasibleConstraints  = errors.New("infeasible constraints")
	
	// Job errors
	ErrJobNotFound            = errors.New("job not found")
	ErrJobFailed              = errors.New("job failed")
	ErrJobAlreadyRunning      = errors.New("job already running")
	ErrJobAlreadyCompleted    = errors.New("job already completed")
	ErrInvalidJobStatus       = errors.New("invalid job status")
	
	// Result errors
	ErrResultNotFound         = errors.New("result not found")
	ErrResultGenerationFailed = errors.New("result generation failed")
	
	// Validation errors
	ErrRequiredField          = errors.New("required field is missing")
	ErrInvalidFormat          = errors.New("invalid format")
	ErrInvalidValue           = errors.New("invalid value")
	ErrInvalidDateRange       = errors.New("end date must be after start date")
	ErrInvalidLatitude        = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude       = errors.New("longitude must be between -180 and 180")
)

// RouteError represents a route optimization domain error
type RouteError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *RouteError) Error() string {
	return e.Message
}

// NewRouteError creates a new route error
func NewRouteError(code, message string) *RouteError {
	return &RouteError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *RouteError {
	return &RouteError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewOptimizationError creates an optimization error
func NewOptimizationError(message string) *RouteError {
	return &RouteError{
		Code:    "OPTIMIZATION_ERROR",
		Message: message,
	}
}

// NewJobError creates a job error
func NewJobError(message string) *RouteError {
	return &RouteError{
		Code:    "JOB_ERROR",
		Message: message,
	}
}

// NewRequestError creates a request error
func NewRequestError(message string) *RouteError {
	return &RouteError{
		Code:    "REQUEST_ERROR",
		Message: message,
	}
}

// NewResultError creates a result error
func NewResultError(message string) *RouteError {
	return &RouteError{
		Code:    "RESULT_ERROR",
		Message: message,
	}
}
