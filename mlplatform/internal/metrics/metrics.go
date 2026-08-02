package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all platform-level Prometheus collectors. Each collector is
// registered against the registry supplied to NewMetrics, which is also
// returned (via Registry) so the /metrics endpoint can serve from it.
type Metrics struct {
	Registry         *prometheus.Registry
	EventsIngested   *prometheus.CounterVec
	ScoreRequests    *prometheus.CounterVec
	ScoreLatency     *prometheus.HistogramVec
	FeatureCacheHits *prometheus.CounterVec
	DriftAlerts      *prometheus.CounterVec
	BusPublishErrors *prometheus.CounterVec
	ActiveModels     prometheus.Gauge
}

// NewMetrics constructs every collector against the given registry.
func NewMetrics(reg *prometheus.Registry) *Metrics {
	f := promauto.With(reg)
	return &Metrics{
		Registry: reg,

		EventsIngested: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mlplatform",
			Name:      "events_ingested_total",
			Help:      "Total events ingested by type.",
		}, []string{"event_type", "source"}),

		ScoreRequests: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mlplatform",
			Name:      "score_requests_total",
			Help:      "Total scoring requests by model, variant and status.",
		}, []string{"model", "variant", "status"}),

		ScoreLatency: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "mlplatform",
			Name:      "score_latency_seconds",
			Help:      "End-to-end scoring latency (feature fetch + inference).",
			Buckets:   []float64{0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}, []string{"model"}),

		FeatureCacheHits: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mlplatform",
			Name:      "feature_cache_total",
			Help:      "Feature store cache hits vs misses.",
		}, []string{"entity_type", "result"}),

		DriftAlerts: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mlplatform",
			Name:      "drift_alerts_total",
			Help:      "Drift detections by model and feature.",
		}, []string{"model", "feature"}),

		BusPublishErrors: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mlplatform",
			Name:      "bus_publish_errors_total",
			Help:      "Event bus publish failures.",
		}, []string{"subject"}),

		ActiveModels: f.NewGauge(prometheus.GaugeOpts{
			Namespace: "mlplatform",
			Name:      "active_models",
			Help:      "Number of models currently loaded in the inference engine.",
		}),
	}
}
