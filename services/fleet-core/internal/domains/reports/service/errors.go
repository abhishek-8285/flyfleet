package service

import "errors"

// Reports V2 service errors
var (
	// General errors
	ErrReportNotFound      = errors.New("report not found")
	ErrInvalidReport       = errors.New("invalid report")
	ErrInvalidReportType   = errors.New("invalid report type")
	ErrInvalidReportStatus = errors.New("invalid report status")

	// Report generation errors
	ErrReportGenerationFailed = errors.New("report generation failed")
	ErrReportProcessingFailed = errors.New("report processing failed")
	ErrReportExportFailed     = errors.New("report export failed")

	// Schedule errors
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrInvalidSchedule       = errors.New("invalid schedule")
	ErrInvalidFrequency      = errors.New("invalid frequency")
	ErrInvalidCronExpression = errors.New("invalid cron expression")
	ErrScheduleAlreadyActive = errors.New("schedule is already active")
	ErrScheduleAlreadyPaused = errors.New("schedule is already paused")

	// Execution errors
	ErrExecutionNotFound       = errors.New("execution not found")
	ErrExecutionFailed         = errors.New("execution failed")
	ErrExecutionTimeout        = errors.New("execution timeout")
	ErrExecutionAlreadyRunning = errors.New("execution already running")

	// Export errors
	ErrInvalidFormat        = errors.New("invalid export format")
	ErrFileNotFound         = errors.New("file not found")
	ErrFileGenerationFailed = errors.New("file generation failed")

	// Queue errors
	ErrQueueConnectionFailed = errors.New("queue connection failed")
	ErrQueueEnqueueFailed    = errors.New("failed to enqueue job")
	ErrQueueDequeueFailed    = errors.New("failed to dequeue job")

	// Settings errors
	ErrSettingsNotFound = errors.New("settings not found")
	ErrInvalidSettings  = errors.New("invalid settings")

	// Validation errors
	ErrRequiredField    = errors.New("required field is missing")
	ErrInvalidValue     = errors.New("invalid value")
	ErrInvalidDateRange = errors.New("end date must be after start date")
)

// ReportsError represents a reports domain error
type ReportsError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *ReportsError) Error() string {
	return e.Message
}

// NewReportsError creates a new reports error
func NewReportsError(code, message string) *ReportsError {
	return &ReportsError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *ReportsError {
	return &ReportsError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewReportError creates a report error
func NewReportError(message string) *ReportsError {
	return &ReportsError{
		Code:    "REPORT_ERROR",
		Message: message,
	}
}

// NewScheduleError creates a schedule error
func NewScheduleError(message string) *ReportsError {
	return &ReportsError{
		Code:    "SCHEDULE_ERROR",
		Message: message,
	}
}

// NewExecutionError creates an execution error
func NewExecutionError(message string) *ReportsError {
	return &ReportsError{
		Code:    "EXECUTION_ERROR",
		Message: message,
	}
}

// NewExportError creates an export error
func NewExportError(message string) *ReportsError {
	return &ReportsError{
		Code:    "EXPORT_ERROR",
		Message: message,
	}
}

// NewQueueError creates a queue error
func NewQueueError(message string) *ReportsError {
	return &ReportsError{
		Code:    "QUEUE_ERROR",
		Message: message,
	}
}

// NewSettingsError creates a settings error
func NewSettingsError(message string) *ReportsError {
	return &ReportsError{
		Code:    "SETTINGS_ERROR",
		Message: message,
	}
}
