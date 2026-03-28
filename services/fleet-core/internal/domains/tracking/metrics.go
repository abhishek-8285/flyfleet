package trackingv2

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds the prometheus metrics for the Tracking V2 Module.
type Metrics struct {
	TelemetryDroppedTotal     *prometheus.CounterVec
	EventsReceivedTotal       *prometheus.CounterVec
	IngestionLatency          *prometheus.HistogramVec
	TelemetryZombiePingsTotal *prometheus.CounterVec
	GapsTotal                 *prometheus.CounterVec
	ExpectedTotal             *prometheus.CounterVec
}

// NewMetrics creates and registers tracking metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		TelemetryDroppedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "telemetry_dropped_total",
			Help: "Total number of telemetry packets dropped.",
		}, []string{"reason", "device_id"}),

		EventsReceivedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "telemetry_events_received_total",
			Help: "Total number of telemetry events received.",
		}, []string{"status", "org_id"}),

		IngestionLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "telemetry_ingestion_latency_seconds",
			Help:    "Latency of telemetry ingestion in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"type"}),

		TelemetryZombiePingsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "telemetry_zombie_pings_total",
			Help: "Total number of blocked zombie pings.",
		}, []string{"org_id"}),

		GapsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "telemetry_gaps_total",
			Help: "Total number of telemetry gaps detected (missing messages).",
		}, []string{"org_id", "vehicle_id", "gap_reason"}),

		ExpectedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "telemetry_expected_total",
			Help: "Total number of expected telemetry messages based on schedule.",
		}, []string{"org_id", "vehicle_id"}),
	}
}
