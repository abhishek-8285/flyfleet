package metrics

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics (RED Method)
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fleet_requests_total",
		Help: "Total number of platform requests",
	}, []string{"method", "path", "status"})

	RequestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "fleet_request_latency_seconds",
		Help:    "Request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	// Business Metrics
	ActiveOrganizations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fleet_active_organizations",
		Help: "Number of active organizations",
	})

	SubscriptionUsage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fleet_subscription_usage",
		Help: "Usage of subscription quotas",
	}, []string{"organization_id", "metric"})

	// Vehicle & Fleet Metrics
	VehiclesOnline = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fleet_vehicles_online",
		Help: "Number of vehicles currently online",
	})

	ActiveTrips = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fleet_active_trips",
		Help: "Number of currently active trips",
	})

	TripsCompletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fleet_trips_completed_total",
		Help: "Total number of completed trips",
	})

	InvoicesGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fleet_invoices_generated_total",
		Help: "Total number of invoices generated",
	})

	// Commands & Control
	CommandsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fleet_commands_sent_total",
		Help: "Total number of device commands sent",
	}, []string{"command_type", "status"})

	CommandQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fleet_command_queue_depth",
		Help: "Number of pending commands in queue",
	}, []string{"command_type"})

	// Alerting
	AlertsTriggeredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fleet_alerts_triggered_total",
		Help: "Total number of alerts triggered",
	}, []string{"severity", "alert_type"})

	// Users & Auth
	ActiveUserSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fleet_active_user_sessions",
		Help: "Number of active user sessions",
	})

	// Go Runtime Metrics
	GoRoutines = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "go_goroutines",
		Help: "Number of goroutines",
	})

	GoMemAllocated = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_allocated_bytes",
		Help: "Number of bytes allocated and still in use",
	})

	GoTotalGC = promauto.NewCounter(prometheus.CounterOpts{
		Name: "go_gc_total",
		Help: "Total number of GC cycles completed",
	})

	GoGCPausesTotal = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "go_gc_pauses_seconds",
		Help:    "Garbage collection pause duration",
		Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10), // 100µs to ~100s
	})
)

// UpdateRuntimeMetrics updates Go runtime metrics. Call periodically (e.g., every 30s).
func UpdateRuntimeMetrics() {
	// Goroutines
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	GoRoutines.Set(float64(runtime.NumGoroutine()))
	GoMemAllocated.Set(float64(m.Alloc))
	GoTotalGC.Set(float64(m.NumGC))
	// Note: GC pause total is cumulative nanoseconds; we could track per-GC if needed
}
