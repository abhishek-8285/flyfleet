package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for high-frequency telemetry ingestion to prevent log flooding
			if r.URL.Path == "/api/v1/tracking/ingest" || r.URL.Path == "/api/v1/tracking/ingest/v2" || r.URL.Path == "/api/v1/compat/telemetry" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			duration := time.Since(start)

			// Build log fields with trace ID and user context
			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", sw.status),
				zap.Duration("duration", duration),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			}

			// Extract trace ID from context (set by TraceIDMiddleware)
			if traceID := GetTraceID(r.Context()); traceID != "" {
				fields = append(fields, zap.String("trace_id", traceID))
			}

			// Extract user ID from context (if available from auth middleware)
			if userID, ok := r.Context().Value("user_id").(string); ok && userID != "" {
				fields = append(fields, zap.String("user_id", userID))
			}

			// Extract org ID from context (if available from tenant middleware)
			if orgID, ok := r.Context().Value("org_id").(string); ok && orgID != "" {
				fields = append(fields, zap.String("org_id", orgID))
			}

			logger.Info("HTTP request", fields...)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
