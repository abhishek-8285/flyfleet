package middleware

import (
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/platform/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Observability returns a middleware that instruments requests with OpenTelemetry tracing.
func Observability(serviceName string) func(http.Handler) http.Handler {
	tracer := observability.GetTracer()
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip tracing for high-frequency telemetry to reduce CPU overhead
			if r.URL.Path == "/api/v1/tracking/ingest" || r.URL.Path == "/api/v1/tracking/ingest/v2" || r.URL.Path == "/api/v1/compat/telemetry" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract trace context from headers
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start a new span
			opts := []trace.SpanStartOption{
				trace.WithAttributes(semconv.HTTPMethodKey.String(r.Method)),
				trace.WithAttributes(semconv.HTTPURLKey.String(r.URL.Path)),
				trace.WithSpanKind(trace.SpanKindServer),
			}

			spanName := r.Method + " " + r.URL.Path
			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			// Update request with new context
			next.ServeHTTP(w, r.WithContext(ctx))

			// Record status code (simplified, would need a response capture wrapper for full detail)
			// span.SetAttributes(semconv.HTTPStatusCodeKey.Int(status))
		})
	}
}
