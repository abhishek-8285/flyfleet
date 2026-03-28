package middleware

import (
	"fmt"
	"net/http"
	"time"

	sharedmetrics "github.com/fleet-platform/shared/pkg/metrics"

	"github.com/flyfleet/fleet-core-service/internal/platform/metrics"
	"go.uber.org/zap"
)

// Metrics returns a middleware that records HTTP request metrics.
func Metrics(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// Wrap response writer to capture status code if needed (simplified: use 200 if not set)
			// For now we just record latency and count after request completes.
			rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rr, r)

			duration := time.Since(start).Seconds()
			method := r.Method
			path := r.URL.Path
			status := rr.statusCode

			// Record metrics
			metrics.RequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
			sharedmetrics.HTTPRequestDuration.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Observe(duration)
		})
	}
}

// responseRecorder is a minimal wrapper to capture the HTTP status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}
