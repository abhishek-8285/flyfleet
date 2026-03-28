package models

import (
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusUp            HealthStatus = "up"
	HealthStatusDown          HealthStatus = "down"
	HealthStatusDegraded      HealthStatus = "degraded"
	HealthStatusNotConfigured HealthStatus = "not_configured"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Component    string       `json:"component"`
	Status       HealthStatus `json:"status"`
	ResponseTime int64        `json:"response_time_ms,omitempty"`
	Error        string       `json:"error,omitempty"`
	LastChecked  time.Time    `json:"last_checked"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status      HealthStatus           `json:"status"`
	Components  map[string]HealthCheck `json:"components"`
	Uptime      float64                `json:"uptime_percentage"`
	LastChecked time.Time              `json:"last_checked"`
}

// SLAMetrics represents SLA metrics
type SLAMetrics struct {
	Period         string    `json:"period"`
	Availability   float64   `json:"availability_percentage"`
	TotalDowntime  int64     `json:"total_downtime_seconds"`
	TotalIncidents int       `json:"total_incidents"`
	MTTR           int64     `json:"mttr_seconds"` // Mean Time To Recovery
	MTBF           int64     `json:"mtbf_seconds"` // Mean Time Between Failures
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
}

// ComponentMetrics represents metrics for a specific component
type ComponentMetrics struct {
	Component    string       `json:"component"`
	Availability float64      `json:"availability_percentage"`
	ResponseTime MetricsStats `json:"response_time"`
	ErrorRate    float64      `json:"error_rate_percentage"`
	RequestCount int64        `json:"request_count"`
	ErrorCount   int64        `json:"error_count"`
	LastUpdated  time.Time    `json:"last_updated"`
}

// MetricsStats represents statistical metrics
type MetricsStats struct {
	Min     int64   `json:"min_ms"`
	Max     int64   `json:"max_ms"`
	Average float64 `json:"average_ms"`
	P50     int64   `json:"p50_ms"`
	P95     int64   `json:"p95_ms"`
	P99     int64   `json:"p99_ms"`
}

// HealthHistory represents historical health data
type HealthHistory struct {
	ID           string       `json:"id"`
	Timestamp    time.Time    `json:"timestamp"`
	Component    string       `json:"component"`
	Status       HealthStatus `json:"status"`
	ResponseTime int64        `json:"response_time_ms"`
	Error        string       `json:"error,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

// AlertThreshold represents alerting thresholds
type AlertThreshold struct {
	Component            string   `json:"component"`
	Metric               string   `json:"metric"`
	Threshold            float64  `json:"threshold"`
	Operator             string   `json:"operator"` // gt, lt, gte, lte, eq
	Severity             string   `json:"severity"` // critical, warning, info
	Enabled              bool     `json:"enabled"`
	NotificationChannels []string `json:"notification_channels"` // email, slack, sms
}

// HealthAlert represents a health alert
type HealthAlert struct {
	ID             string     `json:"id"`
	AlertType      string     `json:"alert_type"`
	Component      string     `json:"component"`
	Metric         string     `json:"metric"`
	CurrentValue   float64    `json:"current_value"`
	Threshold      float64    `json:"threshold"`
	Severity       string     `json:"severity"`
	Message        string     `json:"message"`
	Status         string     `json:"status"` // active, acknowledged, resolved
	TriggeredAt    time.Time  `json:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// SystemMetrics represents overall system metrics
type SystemMetrics struct {
	TotalRequests       int64                       `json:"total_requests"`
	ActiveRequests      int64                       `json:"active_requests"`
	FailedRequests      int64                       `json:"failed_requests"`
	AverageResponseTime float64                     `json:"average_response_time_ms"`
	ErrorRate           float64                     `json:"error_rate_percentage"`
	Uptime              float64                     `json:"uptime_percentage"`
	ComponentMetrics    map[string]ComponentMetrics `json:"component_metrics"`
	LastUpdated         time.Time                   `json:"last_updated"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Component          string `json:"component"`
	Enabled            bool   `json:"enabled"`
	IntervalSeconds    int    `json:"interval_seconds"`
	TimeoutSeconds     int    `json:"timeout_seconds"`
	HealthyThreshold   int    `json:"healthy_threshold"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
}
