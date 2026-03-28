package service

import "errors"

// Finance service errors
var (
	// Fuel transaction errors
	ErrInvalidFuelTransaction   = errors.New("invalid fuel transaction")
	ErrDuplicateFuelTransaction = errors.New("duplicate fuel transaction")
	ErrFuelAnomalyDetected      = errors.New("fuel anomaly detected")
	ErrInvalidLiters            = errors.New("liters must be greater than zero")
	ErrInvalidAmount            = errors.New("amount must be greater than zero")
	ErrMissingVehicleID         = errors.New("vehicle ID is required")

	// Expense policy errors
	ErrInvalidExpensePolicy = errors.New("invalid expense policy")
	ErrPolicyNotFound       = errors.New("expense policy not found")
	ErrInvalidCategory      = errors.New("invalid expense category")
	ErrInvalidLimitAmount   = errors.New("limit amount must be non-negative")

	// Expense report errors
	ErrInvalidExpenseReport  = errors.New("invalid expense report")
	ErrReportNotFound        = errors.New("expense report not found")
	ErrInvalidTotalAmount    = errors.New("total amount must be greater than zero")
	ErrInvalidReportStatus   = errors.New("invalid report status")
	ErrReportAlreadyApproved = errors.New("report already approved")
	ErrReportAlreadyRejected = errors.New("report already rejected")
	ErrReportNotSubmitted    = errors.New("report must be submitted before approval")

	// Revenue forecast errors
	ErrInvalidRouteID   = errors.New("route ID is required")
	ErrForecastNotFound = errors.New("revenue forecast not found")

	// Validation errors
	ErrRequiredField       = errors.New("required field is missing")
	ErrInvalidCurrency     = errors.New("invalid currency code")
	ErrInvalidOrganization = errors.New("organization ID is required")
	ErrInvalidSubmitter    = errors.New("submitter ID is required")
)

// FinanceError represents a finance domain error
type FinanceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *FinanceError) Error() string {
	return e.Message
}

// NewFinanceError creates a new finance error
func NewFinanceError(code, message string) *FinanceError {
	return &FinanceError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *FinanceError {
	return &FinanceError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewFuelError creates a fuel-related error
func NewFuelError(message string) *FinanceError {
	return &FinanceError{
		Code:    "FUEL_ERROR",
		Message: message,
	}
}

// NewExpenseError creates an expense-related error
func NewExpenseError(message string) *FinanceError {
	return &FinanceError{
		Code:    "EXPENSE_ERROR",
		Message: message,
	}
}
