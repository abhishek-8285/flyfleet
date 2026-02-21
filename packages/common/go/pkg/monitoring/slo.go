package monitoring

import (
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
)

// SLORecorder utility to record SLO related metrics
type SLORecorder struct {
	ServiceName string
}

// NewSLORecorder creates a new SLO recorder
func NewSLORecorder(serviceName string) *SLORecorder {
	return &SLORecorder{ServiceName: serviceName}
}

// RecordLatency records the latency of an operation and compares against SLO
func (r *SLORecorder) RecordLatency(operation string, startTime time.Time, limit time.Duration) {
	duration := time.Since(startTime)
	metrics.RecordGRPC(r.ServiceName, operation, "ok", duration.Seconds())
	
	// Potential for internal SLO threshold violation counter
	if duration > limit {
		// Increment slo_violation_count{service="...", operation="..."}
	}
}

// RecordError records an error and potential SLO impact
func (r *SLORecorder) RecordError(operation string, err error) {
	if err != nil {
		metrics.RecordGRPC(r.ServiceName, operation, "error", 0)
	}
}

// HealthCheck standard
type HealthStatus struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp time.Time         `json:"timestamp"`
}

func NewHealthStatus() *HealthStatus {
	return &HealthStatus{
		Status:    "up",
		Checks:    make(map[string]string),
		Timestamp: time.Now(),
	}
}
