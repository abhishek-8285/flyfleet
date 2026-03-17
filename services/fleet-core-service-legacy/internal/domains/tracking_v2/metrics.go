package trackingv2

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds the prometheus metrics for the Tracking V2 Module.
type Metrics struct {
	TelemetryDroppedTotal     *prometheus.CounterVec
	GPSIngestionTotal         *prometheus.CounterVec
	IngestionLatency          *prometheus.HistogramVec
	TelemetryZombiePingsTotal *prometheus.CounterVec
}

// NewMetrics creates and registers tracking metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		TelemetryDroppedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "telemetry_dropped_total",
			Help: "Total number of telemetry packets dropped.",
		}, []string{"reason", "device_id"}),

		GPSIngestionTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "tracking_gps_ingestion_total",
			Help: "Total number of GPS pings ingested.",
		}, []string{"status", "org_id"}),

		IngestionLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "tracking_ingestion_duration_seconds",
			Help:    "Latency of telemetry ingestion in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"type"}),

		TelemetryZombiePingsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "telemetry_zombie_pings_total",
			Help: "Total number of blocked zombie pings.",
		}, []string{"org_id"}),
	}
}
