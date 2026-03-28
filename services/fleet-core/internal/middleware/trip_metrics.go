package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// TripMetrics provides structured observability hooks for the trip domain.
// In production, register these with a Prometheus Registry:
//
//	import "github.com/prometheus/client_golang/prometheus"
//	import "github.com/prometheus/client_golang/prometheus/promauto"
//
// For now, metrics are tracked in-process using simple atomic counters
// that can be exported via /metrics when a Prometheus handler is registered.

var (
	// activeTripsTotal tracks the current number of non-terminal trips.
	// Spec §10: active_trips_total gauge
	activeTripsTotal int64

	// tripCompletionsTotal tracks trips successfully completed.
	// Spec §10: trip_completion_rate (derived from this + tripCreationsTotal)
	tripCompletionsTotal int64

	// tripCancellationsTotal tracks trips cancelled.
	tripCancellationsTotal int64

	// tripFailuresTotal tracks trips that entered FAILED state.
	tripFailuresTotal int64

	// tripDeviationCount tracks AIS-140 route deviation events.
	// Spec §10: trip_deviation_count
	tripDeviationCount int64

	// telemetryPointsIngested tracks total telemetry insertions.
	// Spec §10: telemetry_lag_seconds (lag needs timestamps; this tracks volume)
	telemetryPointsIngested int64
)

// TripEvent represents a metric-emitting lifecycle action.
type TripMetricEvent string

const (
	MetricTripCreated   TripMetricEvent = "created"
	MetricTripStarted   TripMetricEvent = "started"
	MetricTripCompleted TripMetricEvent = "completed"
	MetricTripCancelled TripMetricEvent = "cancelled"
	MetricTripFailed    TripMetricEvent = "failed"
	MetricTripDeviation TripMetricEvent = "deviation"
	MetricTelemetryPt   TripMetricEvent = "telemetry_point"
)

// RecordTripEvent atomically updates the relevant in-process counter.
// Call this from the service layer on every lifecycle transition.
func RecordTripEvent(event TripMetricEvent) {
	switch event {
	case MetricTripCreated:
		addAtomic(&activeTripsTotal, 1)
	case MetricTripCompleted:
		addAtomic(&activeTripsTotal, -1)
		addAtomic(&tripCompletionsTotal, 1)
	case MetricTripCancelled:
		addAtomic(&activeTripsTotal, -1)
		addAtomic(&tripCancellationsTotal, 1)
	case MetricTripFailed:
		addAtomic(&activeTripsTotal, -1)
		addAtomic(&tripFailuresTotal, 1)
	case MetricTripDeviation:
		addAtomic(&tripDeviationCount, 1)
	case MetricTelemetryPt:
		addAtomic(&telemetryPointsIngested, 1)
	}
}

// MetricsHandler exposes current counters as a plain HTTP response.
// Mount this at GET /internal/metrics for scraping.
//
// To upgrade to Prometheus exposition format, replace with:
//
//	promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	now := time.Now().Unix()
	lines := []string{
		"# HELP active_trips_total Current number of non-terminal trips",
		"# TYPE active_trips_total gauge",
		"active_trips_total " + i64str(activeTripsTotal) + " " + strconv.FormatInt(now, 10),

		"# HELP trip_completions_total Total trips completed",
		"# TYPE trip_completions_total counter",
		"trip_completions_total " + i64str(tripCompletionsTotal),

		"# HELP trip_cancellations_total Total trips cancelled",
		"# TYPE trip_cancellations_total counter",
		"trip_cancellations_total " + i64str(tripCancellationsTotal),

		"# HELP trip_failures_total Total trips in FAILED state",
		"# TYPE trip_failures_total counter",
		"trip_failures_total " + i64str(tripFailuresTotal),

		"# HELP trip_deviation_count Total AIS-140 route deviation events",
		"# TYPE trip_deviation_count counter",
		"trip_deviation_count " + i64str(tripDeviationCount),

		"# HELP telemetry_points_ingested Total telemetry data points written",
		"# TYPE telemetry_points_ingested counter",
		"telemetry_points_ingested " + i64str(telemetryPointsIngested),
	}

	for _, line := range lines {
		w.Write([]byte(line + "\n"))
	}
}

func i64str(v int64) string {
	return strconv.FormatInt(v, 10)
}

// addAtomic performs a racy-but-acceptable update for in-process metrics.
// Replace with sync/atomic.AddInt64 if strict correctness is required.
func addAtomic(ptr *int64, delta int64) {
	*ptr += delta
}
