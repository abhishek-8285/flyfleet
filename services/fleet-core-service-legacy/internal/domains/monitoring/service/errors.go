package service

import "errors"

// Monitoring service errors
var (
	// Health check errors
	ErrHealthCheckFailed     = errors.New("health check failed")
	ErrDatabaseHealthFailed  = errors.New("database health check failed")
	ErrRedisHealthFailed     = errors.New("redis health check failed")
	ErrNATSHealthFailed      = errors.New("NATS health check failed")
	
	// SLA errors
	ErrSLACalculationFailed  = errors.New("SLA calculation failed")
	ErrInvalidTimeRange      = errors.New("invalid time range")
	ErrMetricsNotFound       = errors.New("metrics not found")
	
	// Metrics errors
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")
	ErrInvalidMetricName     = errors.New("invalid metric name")
	ErrInvalidMetricValue    = errors.New("invalid metric value")
	
	// Alert errors
	ErrAlertCreationFailed   = errors.New("alert creation failed")
	ErrAlertDeliveryFailed   = errors.New("alert delivery failed")
	ErrInvalidAlertThreshold = errors.New("invalid alert threshold")
	
	// History errors
	ErrHistoryNotFound       = errors.New("health history not found")
	ErrHistoryInsertFailed   = errors.New("failed to insert health history")
)

// MonitoringError represents a monitoring domain error
type MonitoringError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *MonitoringError) Error() string {
	return e.Message
}

// NewMonitoringError creates a new monitoring error
func NewMonitoringError(code, message string) *MonitoringError {
	return &MonitoringError{
		Code:    code,
		Message: message,
	}
}

// NewHealthError creates a health check error
func NewHealthError(component, message string) *MonitoringError {
	return &MonitoringError{
		Code:    "HEALTH_CHECK_FAILED",
		Message: component + ": " + message,
		Field:   component,
	}
}

// NewSLAError creates an SLA error
func NewSLAError(message string) *MonitoringError {
	return &MonitoringError{
		Code:    "SLA_ERROR",
		Message: message,
	}
}

// NewMetricsError creates a metrics error
func NewMetricsError(message string) *MonitoringError {
	return &MonitoringError{
		Code:    "METRICS_ERROR",
		Message: message,
	}
}

// NewAlertError creates an alert error
func NewAlertError(message string) *MonitoringError {
	return &MonitoringError{
		Code:    "ALERT_ERROR",
		Message: message,
	}
}
