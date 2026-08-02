package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
)

// RequestID is middleware that adds a unique request ID / trace ID to each request
// and includes it in the response headers for tracing/debugging
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from upstream proxy or client)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = r.Header.Get("X-Trace-ID")
		}
		if requestID == "" {
			requestID = r.Header.Get("X-Correlation-ID")
		}

		// Generate new ID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to request context across all common context keys
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, TraceIDKey, requestID)
		ctx = context.WithValue(ctx, "request_id", requestID)
		ctx = context.WithValue(ctx, "trace_id", requestID)
		r = r.WithContext(ctx)

		// Add to response headers for client & browser inspection
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Trace-ID", requestID)
		w.Header().Set("X-Correlation-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// GetRequestID extracts the request ID / trace ID from context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	keys := []any{RequestIDKey, TraceIDKey, "request_id", "trace_id"}
	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			if idStr, ok := val.(string); ok && idStr != "" {
				return idStr
			}
			if u, ok := val.(uuid.UUID); ok && u != uuid.Nil {
				return u.String()
			}
		}
	}
	return ""
}

// RequestIDFromContext is a helper that returns request ID or "unknown" if not set
func RequestIDFromContext(ctx context.Context) string {
	if id := GetRequestID(ctx); id != "" {
		return id
	}
	return "unknown"
}
