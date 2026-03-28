package geofencing

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Geofence management metrics
	GeofencesCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofences_created_total",
		Help: "Total geofences created",
	}, []string{"org_id", "geofence_type", "shape"})

	GeofencesUpdatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofences_updated_total",
		Help: "Total geofence updates",
	}, []string{"org_id", "update_type"})

	GeofencesDeletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofences_deleted_total",
		Help: "Total geofences deleted",
	}, []string{"org_id"})

	ActiveGeofences = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "geofences_active_total",
		Help: "Currently active geofences",
	}, []string{"org_id", "geofence_type"})

	GeofenceCreationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_creation_duration_seconds",
		Help:    "Time to create and index a geofence",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 6),
	}, []string{"org_id", "geofence_type"})

	// Validation metrics
	GeofenceValidationRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_validation_requests_total",
		Help: "Total geofence validation requests",
	}, []string{"org_id"})

	GeofenceValidationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_validation_duration_seconds",
		Help:    "Geofence validation lookup latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 8), // 1ms to ~2.5s
	}, []string{"org_id"})

	GeofenceValidationHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_validation_hits_total",
		Help: "Total validation requests where point is inside geofence",
	}, []string{"org_id"})

	GeofenceValidationMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_validation_misses_total",
		Help: "Total validation requests where point is outside all geofences",
	}, []string{"org_id"})

	// Violation detection metrics
	GeofenceViolationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_violations_total",
		Help: "Total geofence violations detected",
	}, []string{"org_id", "geofence_id", "violation_type", "severity"})

	GeofenceViolationDetectionLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_violation_detection_latency_seconds",
		Help:    "Time from telemetry receipt to violation detection",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 6),
	}, []string{"org_id"})

	// Alert metrics
	GeofenceAlertsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_alerts_sent_total",
		Help: "Total geofence violation alerts sent",
	}, []string{"org_id", "alert_type", "channel"})

	GeofenceAlertDeliverySuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "geofence_alert_delivery_success_rate_percentage",
		Help: "Geofence alert delivery success rate",
	}, []string{"org_id", "channel"})

	// Complexity metrics
	GeofenceVertexCount = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_vertex_count",
		Help:    "Number of vertices in geofence polygons",
		Buckets: prometheus.ExponentialBuckets(10, 2, 8),
	}, []string{"org_id"})

	GeofenceAreaSquareKm = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_area_square_km",
		Help:    "Area of geofences in square kilometers",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	}, []string{"org_id", "geofence_type"})

	// Performance metrics
	GeofenceIndexBuildDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_index_build_duration_seconds",
		Help:    "Time to build geofence spatial index",
		Buckets: prometheus.ExponentialBuckets(1, 2, 8),
	}, []string{"org_id", "index_type"})

	GeofenceIndexLookupLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_index_lookup_latency_seconds",
		Help:    "Geofence index lookup latency",
		Buckets: prometheus.ExponentialBuckets(0.0001, 2, 8),
	}, []string{"org_id", "index_type"})

	// Cache metrics
	GeofenceCacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_cache_hits_total",
		Help: "Total geofence validation cache hits",
	}, []string{"org_id", "cache_type"})

	GeofenceCacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_cache_misses_total",
		Help: "Total geofence validation cache misses",
	}, []string{"org_id", "cache_type"})

	// Batch processing metrics
	GeofenceBatchSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_batch_size",
		Help:    "Number of points in geofence validation batches",
		Buckets: prometheus.LinearBuckets(1, 100, 10),
	}, []string{"org_id"})

	GeofenceBatchProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "geofence_batch_processing_duration_seconds",
		Help:    "Time to process a batch of geofence validations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
	}, []string{"org_id"})

	// Overlap detection
	GeofenceOverlapsDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_overlaps_detected_total",
		Help: "Total overlapping geofences detected",
	}, []string{"org_id", "overlap_type"})

	// Zone type metrics
	GeofenceZoneTypeCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "geofence_zone_type_count",
		Help: "Count of geofences by zone type",
	}, []string{"org_id", "zone_type"})

	// Speed limit enforcement
	SpeedLimitViolationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_speed_limit_violations_total",
		Help: "Total speed limit violations within geofences",
	}, []string{"org_id", "geofence_id", "speed_limit_exceeded_by"})

	// Time-based restrictions
	TimeBasedViolationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofence_time_based_violations_total",
		Help: "Violations of time-based geofence restrictions",
	}, []string{"org_id", "restriction_type"})

	// Dynamic geofences
	DynamicGeofencesActiveTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "geofences_dynamic_active_total",
		Help: "Currently active dynamic geofences",
	}, []string{"org_id", "dynamic_type"})

	DynamicGeofenceUpdatesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "geofences_dynamic_updates_total",
		Help: "Total dynamic geofence position updates",
	}, []string{"org_id", "update_trigger"})
)

// Global metrics instance
var GlobalMetrics = struct {
	GeofencesCreatedTotal             *prometheus.CounterVec
	GeofencesUpdatedTotal             *prometheus.CounterVec
	GeofencesDeletedTotal             *prometheus.CounterVec
	ActiveGeofences                   *prometheus.GaugeVec
	GeofenceCreationDuration          *prometheus.HistogramVec
	GeofenceValidationRequestsTotal   *prometheus.CounterVec
	GeofenceValidationDuration        *prometheus.HistogramVec
	GeofenceValidationHits            *prometheus.CounterVec
	GeofenceValidationMisses          *prometheus.CounterVec
	GeofenceViolationsTotal           *prometheus.CounterVec
	GeofenceViolationDetectionLatency *prometheus.HistogramVec
	GeofenceAlertsSentTotal           *prometheus.CounterVec
	GeofenceAlertDeliverySuccessRate  *prometheus.GaugeVec
	GeofenceVertexCount               *prometheus.HistogramVec
	GeofenceAreaSquareKm              *prometheus.HistogramVec
	GeofenceIndexBuildDuration        *prometheus.HistogramVec
	GeofenceIndexLookupLatency        *prometheus.HistogramVec
	GeofenceCacheHitsTotal            *prometheus.CounterVec
	GeofenceCacheMissesTotal          *prometheus.CounterVec
	GeofenceBatchSize                 *prometheus.HistogramVec
	GeofenceBatchProcessingDuration   *prometheus.HistogramVec
	GeofenceOverlapsDetectedTotal     *prometheus.CounterVec
	GeofenceZoneTypeCount             *prometheus.GaugeVec
	SpeedLimitViolationsTotal         *prometheus.CounterVec
	TimeBasedViolationsTotal          *prometheus.CounterVec
	DynamicGeofencesActiveTotal       *prometheus.GaugeVec
	DynamicGeofenceUpdatesTotal       *prometheus.CounterVec
}{
	GeofencesCreatedTotal:             GeofencesCreatedTotal,
	GeofencesUpdatedTotal:             GeofencesUpdatedTotal,
	GeofencesDeletedTotal:             GeofencesDeletedTotal,
	ActiveGeofences:                   ActiveGeofences,
	GeofenceCreationDuration:          GeofenceCreationDuration,
	GeofenceValidationRequestsTotal:   GeofenceValidationRequestsTotal,
	GeofenceValidationDuration:        GeofenceValidationDuration,
	GeofenceValidationHits:            GeofenceValidationHits,
	GeofenceValidationMisses:          GeofenceValidationMisses,
	GeofenceViolationsTotal:           GeofenceViolationsTotal,
	GeofenceViolationDetectionLatency: GeofenceViolationDetectionLatency,
	GeofenceAlertsSentTotal:           GeofenceAlertsSentTotal,
	GeofenceAlertDeliverySuccessRate:  GeofenceAlertDeliverySuccessRate,
	GeofenceVertexCount:               GeofenceVertexCount,
	GeofenceAreaSquareKm:              GeofenceAreaSquareKm,
	GeofenceIndexBuildDuration:        GeofenceIndexBuildDuration,
	GeofenceIndexLookupLatency:        GeofenceIndexLookupLatency,
	GeofenceCacheHitsTotal:            GeofenceCacheHitsTotal,
	GeofenceCacheMissesTotal:          GeofenceCacheMissesTotal,
	GeofenceBatchSize:                 GeofenceBatchSize,
	GeofenceBatchProcessingDuration:   GeofenceBatchProcessingDuration,
	GeofenceOverlapsDetectedTotal:     GeofenceOverlapsDetectedTotal,
	GeofenceZoneTypeCount:             GeofenceZoneTypeCount,
	SpeedLimitViolationsTotal:         SpeedLimitViolationsTotal,
	TimeBasedViolationsTotal:          TimeBasedViolationsTotal,
	DynamicGeofencesActiveTotal:       DynamicGeofencesActiveTotal,
	DynamicGeofenceUpdatesTotal:       DynamicGeofenceUpdatesTotal,
}
