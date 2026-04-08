package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter wraps http.ResponseWriter to capture status code and body size
type responseWriter struct {
	http.ResponseWriter
	statusCode      int
	bodyBytesWritten int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bodyBytesWritten += n
	return n, err
}

// ResponseLogger is middleware that logs all HTTP responses with status, duration, and size
func ResponseLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			duration := time.Since(start)

			fields := []zap.Field{
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration_ms", duration),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("client_ip", r.RemoteAddr),
				zap.Int("response_bytes", wrapped.bodyBytesWritten),
			}

			requestID := GetRequestID(r.Context())
			if requestID != "" && requestID != "unknown" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			logger.Info("http response", fields...)
		})
	}
}
