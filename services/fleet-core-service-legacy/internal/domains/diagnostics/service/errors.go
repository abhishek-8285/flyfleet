package service

import "errors"

// Diagnostics service errors
var (
	// Device errors
	ErrDeviceNotFound      = errors.New("device not found")
	ErrDeviceAlreadyExists = errors.New("device with this serial number already exists")
	ErrInvalidDeviceStatus = errors.New("invalid device status")
	ErrDeviceOffline       = errors.New("device is offline")
	
	// Command errors
	ErrCommandNotFound     = errors.New("command not found")
	ErrInvalidCommand      = errors.New("invalid command")
	ErrInvalidTransition   = errors.New("invalid state transition")
	ErrCommandTimedOut     = errors.New("command timed out")
	ErrCommandFailed       = errors.New("command failed")
	
	// Telemetry errors
	ErrTelemetryNotFound   = errors.New("telemetry not found")
	ErrInvalidTelemetry    = errors.New("invalid telemetry data")
	
	// Device twin errors
	ErrTwinNotFound        = errors.New("device twin not found")
	ErrStateSyncFailed     = errors.New("failed to sync device state")
	
	// Firmware errors
	ErrFirmwareNotFound    = errors.New("firmware package not found")
	ErrInvalidFirmware     = errors.New("invalid firmware package")
	ErrFirmwareNotPublished = errors.New("firmware not published")
	
	// Rollout errors
	ErrRolloutNotFound     = errors.New("rollout not found")
	ErrInvalidRollout      = errors.New("invalid rollout configuration")
	ErrRolloutAlreadyStarted = errors.New("rollout already started")
	ErrRolloutAlreadyCompleted = errors.New("rollout already completed")
	ErrInvalidRolloutTransition = errors.New("invalid rollout status transition")
	
	// Validation errors
	ErrRequiredField       = errors.New("required field is missing")
	ErrInvalidFormat       = errors.New("invalid field format")
	ErrInvalidValue        = errors.New("invalid value")
)

// DiagnosticsError represents a diagnostics domain error
type DiagnosticsError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *DiagnosticsError) Error() string {
	return e.Message
}

// NewDiagnosticsError creates a new diagnostics error
func NewDiagnosticsError(code, message string) *DiagnosticsError {
	return &DiagnosticsError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *DiagnosticsError {
	return &DiagnosticsError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}
