package analytics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds the prometheus metrics for the Analytics Module.
type Metrics struct {
	QueryTotal              *prometheus.CounterVec
	QueryDurationSeconds    *prometheus.HistogramVec
	CacheHitsTotal          *prometheus.CounterVec
	CacheMissesTotal        *prometheus.CounterVec
	ActiveJobsGauge         *prometheus.GaugeVec
	JobCompletionCounter    *prometheus.CounterVec
	CrossTenantAccessTotal  *prometheus.CounterVec
	CacheInvalidationsTotal *prometheus.CounterVec

	// Phase 8 & 9 ML Metrics
	PredictiveAnalyticCounter *prometheus.CounterVec
	ModelAccuracyGauge        *prometheus.GaugeVec
	InferenceLatencyHistogram *prometheus.HistogramVec
	PredictionErrorCounter    *prometheus.CounterVec
}

// NewMetrics creates and registers analytics metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		QueryTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_query_total",
			Help: "Total number of analytics queries performed.",
		}, []string{"query_type", "organization_id"}),

		QueryDurationSeconds: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "analytics_query_duration_seconds",
			Help:    "Latency of analytics queries in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"query_type", "organization_id"}),

		CacheHitsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_cache_hits_total",
			Help: "Total number of analytics cache hits.",
		}, []string{"query_type", "organization_id"}),

		CacheMissesTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_cache_misses_total",
			Help: "Total number of analytics cache misses.",
		}, []string{"query_type", "organization_id"}),

		ActiveJobsGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "analytics_active_jobs",
			Help: "Number of currently running analytics jobs.",
		}, []string{"job_type", "organization_id"}),

		JobCompletionCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_job_completion_total",
			Help: "Total number of completed analytics jobs.",
		}, []string{"job_type", "status", "organization_id"}),

		CrossTenantAccessTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_cross_tenant_attempts_total",
			Help: "Total number of cross-tenant access attempts blocked by the guard.",
		}, []string{"organization_id", "resource_type"}),

		CacheInvalidationsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_cache_invalidations_total",
			Help: "Total number of analytics cache invalidation events.",
		}, []string{"invalidation_type", "organization_id"}),

		PredictiveAnalyticCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_predictive_analytic_total",
			Help: "Total number of predictive analytics performed.",
		}, []string{"organization_id", "model_type", "status"}),

		ModelAccuracyGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "analytics_model_accuracy",
			Help: "Live accuracy of predictive models.",
		}, []string{"organization_id", "model_id"}),

		InferenceLatencyHistogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "analytics_inference_latency_seconds",
			Help:    "Latency of ML model inference orchestration.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		}, []string{"organization_id", "model_id"}),

		PredictionErrorCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_prediction_error_total",
			Help: "Total number of prediction errors recorded.",
		}, []string{"organization_id", "model_id", "error_type"}),
	}
}
