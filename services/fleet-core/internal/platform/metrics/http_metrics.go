package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	sharedmetrics "github.com/fleet-platform/shared/pkg/metrics"

	"github.com/flyfleet/fleet-core-service/internal/middleware" // for GetTraceID
)

// statusWriter captures response status code (local copy, as middleware.StatusWriter is unexported)
type statusWriter struct {
	http.ResponseWriter
	Status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

// HTTPMetricsMiddleware records request rate, error rate, and latency (RED method)
func HTTPMetricsMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip metrics for health checks and internal endpoints to avoid noise
			path := r.URL.Path
			if path == "/health" || path == "/healthz" || path == "/ready" || path == "/readyz" ||
				path == "/metrics" || path == "/swagger" || path == "/swagger/*" ||
				path == "/api/control/*" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, Status: http.StatusOK}

			// Get route pattern for better cardinality (e.g., /api/v1/vehicles/:id instead of /api/v1/vehicles/123)
			routePattern := getRoutePattern(r)

			next.ServeHTTP(sw, r)

			duration := time.Since(start)
			status := sw.Status

			// Record request count with labels
			RequestsTotal.WithLabelValues(r.Method, routePattern, strconv.Itoa(status)).Inc()

			// Record latency histogram (using shared metrics to avoid duplicate registration)
			sharedmetrics.HTTPRequestDuration.WithLabelValues(r.Method, routePattern, strconv.Itoa(status)).Observe(duration.Seconds())

			// Log slow requests (> 1s) at warning level
			if duration > time.Second {
				logger.Warn("Slow request detected",
					zap.String("method", r.Method),
					zap.String("path", path),
					zap.String("pattern", routePattern),
					zap.Duration("duration", duration),
					zap.Int("status", status),
					zap.String("trace_id", middleware.GetTraceID(r.Context())),
				)
			}
		})
	}
}

// getRoutePattern extracts the chi route pattern if available, otherwise returns the raw path
func getRoutePattern(r *http.Request) string {
	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		if pattern := routeCtx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	// Fallback to raw path (higher cardinality)
	return r.URL.Path
}
