package service

import "errors"

// India integration service errors
var (
	// FASTag errors
	ErrFASTagTransactionNotFound  = errors.New("fastag transaction not found")
	ErrInvalidFASTagTransaction   = errors.New("invalid fastag transaction")
	ErrInvalidTagNumber           = errors.New("invalid tag number")
	ErrInvalidTransactionType     = errors.New("invalid transaction type")
	ErrInvalidTransactionAmount   = errors.New("transaction amount must be positive")
	ErrInvalidTollPlaza           = errors.New("toll plaza is required")
	ErrInvalidTransactionTime     = errors.New("transaction time is required")
	ErrInvalidTransactionStatus   = errors.New("invalid transaction status")

	// E-Way Bill errors
	ErrEWayBillNotFound           = errors.New("e-way bill not found")
	ErrInvalidEWayBill            = errors.New("invalid e-way bill")
	ErrInvalidEWayBillNumber      = errors.New("invalid e-way bill number")
	ErrInvalidDocumentNumber      = errors.New("document number is required")
	ErrInvalidDocumentType        = errors.New("invalid document type")
	ErrInvalidGoodsValue          = errors.New("goods value must be positive")
	ErrInvalidPincode             = errors.New("invalid pincode")
	ErrInvalidStateCode           = errors.New("invalid state code")
	ErrInvalidTransportMode       = errors.New("invalid transport mode")
	ErrInvalidGSTIN               = errors.New("invalid GSTIN format")
	ErrEWayBillExpired            = errors.New("e-way bill has expired")
	ErrEWayBillCancelled          = errors.New("e-way bill is cancelled")

	// Vehicle Verification (VAHAN) errors
	ErrVehicleVerificationNotFound = errors.New("vehicle verification not found")
	ErrInvalidRegistrationNumber   = errors.New("invalid registration number")
	ErrInvalidVerificationStatus   = errors.New("invalid verification status")
	ErrInvalidRCStatus             = errors.New("invalid RC status")
	ErrInsuranceExpired            = errors.New("insurance has expired")
	ErrPUCExpired                  = errors.New("PUC has expired")

	// Driver Verification (SARATHI) errors
	ErrDriverVerificationNotFound  = errors.New("driver verification not found")
	ErrInvalidLicenseNumber        = errors.New("invalid license number")
	ErrLicenseExpired              = errors.New("driver license has expired")
	ErrInvalidLicenseClass         = errors.New("invalid license class")
	ErrInvalidDateOfBirth          = errors.New("invalid date of birth")

	// Local Permit errors
	ErrPermitNotFound              = errors.New("permit not found")
	ErrInvalidPermitNumber         = errors.New("invalid permit number")
	ErrInvalidPermitType           = errors.New("invalid permit type")
	ErrPermitExpired               = errors.New("permit has expired")
	ErrInvalidPermitStatus         = errors.New("invalid permit status")
	ErrInvalidJurisdiction         = errors.New("jurisdiction is required")

	// GST Compliance errors
	ErrGSTComplianceNotFound       = errors.New("gst compliance not found")
	ErrGSTINFormat                 = errors.New("invalid GSTIN format")
	ErrInvalidGSTStatus            = errors.New("invalid GST status")
	ErrGSTNotActive                = errors.New("GST is not active")
	ErrReturnFilingPending         = errors.New("return filing is pending")
	ErrPaymentPending              = errors.New("payment is pending")

	// Integration Config errors
 ErrIntegrationConfigNotFound    = errors.New("integration config not found")
	ErrInvalidServiceType          = errors.New("invalid service type")
	ErrInvalidAPIEndpoint          = errors.New("invalid API endpoint")
	ErrInvalidEnvironment          = errors.New("invalid environment")
	ErrIntegrationNotActive        = errors.New("integration is not active")
	ErrInvalidCredentials          = errors.New("invalid credentials")

	// Common errors
	ErrRequiredField               = errors.New("required field is missing")
	ErrInvalidOrganization         = errors.New("organization ID is required")
	ErrInvalidVehicle              = errors.New("vehicle ID is required")
	ErrInvalidDriver               = errors.New("driver ID is required")
	ErrInvalidDateFormat           = errors.New("invalid date format")
	ErrDateRangeInvalid            = errors.New("end date cannot be before start date")
	ErrDuplicateRecord             = errors.New("duplicate record exists")
	ErrUnauthorized                = errors.New("unauthorized access")
	ErrForbidden                   = errors.New("access forbidden")
)

// IndiaIntegrationError represents an India integration domain error
type IndiaIntegrationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *IndiaIntegrationError) Error() string {
	return e.Message
}

// NewIndiaIntegrationError creates a new India integration error
func NewIndiaIntegrationError(code, message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewFASTagError creates a FASTag-specific error
func NewFASTagError(message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    "FASTAG_ERROR",
		Message: message,
	}
}

// NewEWayBillError creates an E-Way Bill-specific error
func NewEWayBillError(message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    "EWAYBILL_ERROR",
		Message: message,
	}
}

// NewVehicleVerificationError creates a vehicle verification error
func NewVehicleVerificationError(message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    "VAHAN_ERROR",
		Message: message,
	}
}

// NewDriverVerificationError creates a driver verification error
func NewDriverVerificationError(message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    "SARATHI_ERROR",
		Message: message,
	}
}

// NewPermitError creates a permit error
func NewPermitError(message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    "PERMIT_ERROR",
		Message: message,
	}
}

// NewGSTError creates a GST error
func NewGSTError(message string) *IndiaIntegrationError {
	return &IndiaIntegrationError{
		Code:    "GST_ERROR",
		Message: message,
	}
}

// ValidateGSTIN validates GSTIN format
func ValidateGSTIN(gstin string) bool {
	if len(gstin) != 15 {
		return false
	}
	// Basic format: 2 digits state code + 10 chars PAN + 1 digit entity + 1 char Z + 1 digit check
	// This is a simplified check - in production, use proper validation
	return true
}

// ValidatePincode validates Indian pincode format
func ValidatePincode(pincode string) bool {
	if len(pincode) != 6 {
		return false
	}
	// Check if all characters are digits
	for _, c := range pincode {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ValidateVehicleRegistration validates Indian vehicle registration format
func ValidateVehicleRegistration(reg string) bool {
	// Basic format: XX XX XXXX (state code, district code, series, number)
	// Example: MH 12 AB 1234
	if len(reg) < 10 || len(reg) > 15 {
		return false
	}
	return true
}

// ValidateLicenseNumber validates Indian driving license format
func ValidateLicenseNumber(license string) bool {
	// Format varies by state, but typically 15-16 characters
	if len(license) < 15 || len(license) > 16 {
		return false
	}
	return true
}
