package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics (RED Method)
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	// Business Metrics
	ActiveOrganizations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_organizations",
		Help: "Number of active organizations",
	})

	SubscriptionUsage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "subscription_usage",
		Help: "Usage of subscription quotas",
	}, []string{"organization_id", "metric"})

	// Vehicle & Fleet Metrics
	VehiclesOnline = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vehicles_online",
		Help: "Number of vehicles currently online",
	})

	ActiveTrips = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_trips",
		Help: "Number of currently active trips",
	})

	TripsCompletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "trips_completed_total",
		Help: "Total number of completed trips",
	})

	InvoicesGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "invoices_generated_total",
		Help: "Total number of invoices generated",
	})

	// Commands & Control
	DeviceCommandEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "device_command_events_total",
		Help: "Total number of device command events (sent/acknowledged)",
	}, []string{"command_type", "result"})

	DeviceCommandAckLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "device_command_ack_latency_seconds",
		Help:    "Device command acknowledgement latency",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
	}, []string{"command_type"})

	CommandQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "command_queue_depth",
		Help: "Number of pending commands in queue",
	}, []string{"command_type"})

	// Alerting
	AlertsTriggeredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_triggered_total",
		Help: "Total number of alerts triggered",
	}, []string{"severity", "alert_type"})

	// Users & Auth
	ActiveUserSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_user_sessions",
		Help: "Number of active user sessions",
	})

	// Database Connection Pool Metrics
	DatabaseConnectionsActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "database_connections_active",
		Help: "Number of currently active database connections",
	}, []string{"pool"}) // pool: critical, high, medium, low

	DatabaseConnectionsIdle = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "database_connections_idle",
		Help: "Number of idle database connections in the pool",
	}, []string{"pool"})

	DatabaseConnectionsMax = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "database_connections_max",
		Help: "Maximum configured database connections per pool",
	}, []string{"pool"})

	DatabaseConnectionUtilization = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "database_connection_utilization_percentage",
		Help: "Percentage of max connections currently in use",
	}, []string{"pool"})
)
