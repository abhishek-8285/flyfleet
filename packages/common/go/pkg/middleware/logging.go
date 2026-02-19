package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter is a wrapper around http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := wrapResponseWriter(w)

			// Process request
			next.ServeHTTP(ww, r)

			// Log request details
			duration := time.Since(start)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.String("client_ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Duration("duration", duration),
			}

			// Add request ID if present in context
			if requestID, ok := r.Context().Value(RequestIDKey).(string); ok {
				fields = append(fields, zap.String("request_id", requestID))
			}

			// Add user ID if present in context
			if userID, ok := r.Context().Value(UserIDKey).(string); ok {
				fields = append(fields, zap.String("user_id", userID))
			}

			// Log based on status code
			if ww.Status() >= 500 {
				logger.Error("HTTP server error", fields...)
			} else if ww.Status() >= 400 {
				logger.Warn("HTTP client error", fields...)
			} else {
				logger.Info("HTTP request completed", fields...)
			}
		})
	}
}