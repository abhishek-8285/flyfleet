package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

// RequestID is middleware that adds a unique request ID to each request
// and includes it in the response headers for tracing/debugging
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from upstream proxy)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = r.Header.Get("X-Correlation-ID")
		}

		// Generate new ID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to request context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add to response headers
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Correlation-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
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
