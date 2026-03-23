package service

import "errors"

// Vehicle service errors
var (
	// General errors
	ErrVehicleNotFound      = errors.New("vehicle not found")
	ErrVehicleAlreadyExists = errors.New("vehicle already exists")
	ErrInvalidVehicleID     = errors.New("invalid vehicle ID")
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrInvalidOrgID         = errors.New("invalid organization ID")

	// Validation errors
	ErrMissingRequiredField   = errors.New("missing required field")
	ErrInvalidFieldValue      = errors.New("invalid field value")
	ErrInvalidLicensePlate    = errors.New("invalid license plate format")
	ErrInvalidVIN             = errors.New("invalid VIN format")
	ErrInvalidYear            = errors.New("invalid vehicle year")
	ErrInvalidCapacity        = errors.New("invalid capacity")
	ErrInvalidFuelType        = errors.New("invalid fuel type")
	ErrInvalidStatus          = errors.New("invalid vehicle status")
	ErrInvalidOdometer        = errors.New("invalid odometer reading")

	// Business logic errors
	ErrVehicleInUse           = errors.New("vehicle is currently in use")
	ErrVehicleCannotBeDeleted = errors.New("vehicle cannot be deleted")
	ErrVehicleCannotBeUpdated = errors.New("vehicle cannot be updated in current status")
	ErrDriverAlreadyAssigned  = errors.New("driver already assigned to vehicle")
	ErrDriverNotAssigned      = errors.New("no driver assigned to vehicle")

	// Bulk operation errors
	ErrInvalidFileType        = errors.New("invalid file type")
	ErrFileTooLarge           = errors.New("file too large")
	ErrInvalidCSVFormat       = errors.New("invalid CSV format")
	ErrBulkJobNotFound        = errors.New("bulk job not found")
	ErrBulkJobAlreadyProcessed = errors.New("bulk job already processed")
	ErrMaxRetryLimitExceeded  = errors.New("max retry limit exceeded")
	ErrFileNotFound           = errors.New("original file not found")

	// Trip assignment errors
	ErrTripAlreadyAssigned    = errors.New("trip already assigned to vehicle")
	ErrTripNotAssigned        = errors.New("trip not assigned to vehicle")
	ErrVehicleNotAvailable    = errors.New("vehicle not available for trip")

	// Enterprise lifecycle errors (per spec §4 Business Rules)
	ErrLegalHoldActive            = errors.New("vehicle is on legal hold — modification prevented")
	ErrDecommissionedIrreversible = errors.New("DECOMMISSIONED status is irreversible")
	ErrComplianceRequired         = errors.New("vehicle cannot be activated: compliance requirement not met")
	ErrGPSDeviceConflict          = errors.New("GPS device is already bound to another vehicle")
	ErrTenantMismatch             = errors.New("parent vehicle must belong to the same tenant")
)

// VehicleError represents a vehicle domain error with additional context
type VehicleError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Field      string `json:"field,omitempty"`
	VehicleID  string `json:"vehicle_id,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *VehicleError) Error() string {
	return e.Message
}

// NewVehicleError creates a new vehicle error
func NewVehicleError(code, message string) *VehicleError {
	return &VehicleError{
		Code:    code,
		Message: message,
	}
}

// NewVehicleErrorWithField creates a new vehicle error with field context
func NewVehicleErrorWithField(code, message, field string) *VehicleError {
	return &VehicleError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// NewVehicleErrorWithVehicleID creates a new vehicle error with vehicle ID context
func NewVehicleErrorWithVehicleID(code, message, vehicleID string) *VehicleError {
	return &VehicleError{
		Code:    code,
		Message: message,
		VehicleID: vehicleID,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *VehicleError {
	return &VehicleError{
		Code:     "VALIDATION_ERROR",
		Message:  message,
		Field:    field,
		StatusCode: 400,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, id string) *VehicleError {
	return &VehicleError{
		Code:     "NOT_FOUND",
		Message:  resource + " not found: " + id,
		StatusCode: 404,
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *VehicleError {
	return &VehicleError{
		Code:     "UNAUTHORIZED",
		Message:  message,
		StatusCode: 403,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *VehicleError {
	return &VehicleError{
		Code:     "CONFLICT",
		Message:  message,
		StatusCode: 409,
	}
}

// NewBulkError creates a bulk operation error
func NewBulkError(code, message string) *VehicleError {
	return &VehicleError{
		Code:     "BULK_ERROR_" + code,
		Message:  message,
		StatusCode: 400,
	}
}
