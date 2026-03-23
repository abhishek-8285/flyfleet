package service

import "errors"

// Maintenance service errors
var (
	// Work order errors
	ErrWorkOrderNotFound       = errors.New("work order not found")
	ErrInvalidWorkOrder        = errors.New("invalid work order")
	ErrInvalidWorkOrderType    = errors.New("invalid maintenance type")
	ErrInvalidWorkOrderStatus  = errors.New("invalid work order status")
	ErrInvalidPriority         = errors.New("invalid priority")
	ErrInvalidCost             = errors.New("cost must be non-negative")
	ErrInvalidDate             = errors.New("end date must be after start date")
	ErrDuplicateWorkOrder      = errors.New("duplicate work order for vehicle")
	
	// Inventory errors
	ErrPartNotFound           = errors.New("inventory part not found")
	ErrInvalidPart            = errors.New("invalid inventory part")
	ErrInvalidQuantity        = errors.New("quantity must be positive")
	ErrInsufficientStock      = errors.New("insufficient stock")
	ErrNegativeStock          = errors.New("stock cannot be negative")
	ErrInvalidUnitCost        = errors.New("unit cost must be non-negative")
	
	// Vendor errors
	ErrVendorNotFound         = errors.New("vendor not found")
	ErrInvalidVendor          = errors.New("invalid vendor")
	ErrInvalidVendorRating    = errors.New("vendor rating must be between 0 and 5")
	
	// Inspection errors
	ErrInspectionNotFound     = errors.New("inspection not found")
	ErrInvalidInspection      = errors.New("invalid inspection")
	ErrInvalidInspectionStatus = errors.New("invalid inspection status")
	ErrInvalidTemplate        = errors.New("invalid inspection template")
	
	// Schedule errors
	ErrScheduleNotFound       = errors.New("maintenance schedule not found")
	ErrInvalidSchedule        = errors.New("invalid maintenance schedule")
	ErrInvalidScheduleType    = errors.New("invalid maintenance type")
	ErrScheduleAlreadyCompleted = errors.New("schedule already completed")
	
	// Log errors
	ErrLogNotFound            = errors.New("maintenance log not found")
	ErrInvalidLog             = errors.New("invalid maintenance log")
	ErrLogAlreadyCompleted    = errors.New("log already completed")
	
	// Settings errors
	ErrSettingsNotFound       = errors.New("maintenance settings not found")
	ErrInvalidSLA             = errors.New("SLA hours must be positive")
	ErrInvalidThreshold       = errors.New("threshold must be non-negative")
	
	// Workflow errors
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrWorkOrderAlreadyCompleted = errors.New("work order already completed")
	ErrInsufficientApproval    = errors.New("insufficient approval for this action")
	
	// SLA errors
	ErrSLABreach              = errors.New("SLA breach detected")
	ErrSLAExceeded            = errors.New("SLA time exceeded")
	
	// Validation errors
	ErrRequiredField          = errors.New("required field is missing")
	ErrInvalidOrganization    = errors.New("organization ID is required")
	ErrInvalidVehicle         = errors.New("vehicle ID is required")
	ErrInvalidMechanic        = errors.New("mechanic ID is required")
)

// MaintenanceError represents a maintenance domain error
type MaintenanceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *MaintenanceError) Error() string {
	return e.Message
}

// NewMaintenanceError creates a new maintenance error
func NewMaintenanceError(code, message string) *MaintenanceError {
	return &MaintenanceError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *MaintenanceError {
	return &MaintenanceError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewWorkOrderError creates a work order error
func NewWorkOrderError(message string) *MaintenanceError {
	return &MaintenanceError{
		Code:    "WORK_ORDER_ERROR",
		Message: message,
	}
}

// NewInventoryError creates an inventory error
func NewInventoryError(message string) *MaintenanceError {
	return &MaintenanceError{
		Code:    "INVENTORY_ERROR",
		Message: message,
	}
}

// NewSLAError creates an SLA error
func NewSLAError(message string) *MaintenanceError {
	return &MaintenanceError{
		Code:    "SLA_ERROR",
		Message: message,
	}
}
