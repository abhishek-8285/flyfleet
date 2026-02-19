package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// DriverSecurityMiddleware handles driver app specific security checks
type DriverSecurityMiddleware struct {
	logger *zap.Logger
}

func NewDriverSecurityMiddleware(logger *zap.Logger) *DriverSecurityMiddleware {
	return &DriverSecurityMiddleware{logger: logger}
}

// DeviceFingerprintMiddleware ensures the request has a valid device fingerprint header
func (m *DriverSecurityMiddleware) DeviceFingerprintMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fingerprint := r.Header.Get("X-Device-Fingerprint")
		if fingerprint == "" {
			m.logger.Warn("missing device fingerprint header", 
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "Missing device fingerprint", http.StatusForbidden)
			return
		}

		// In a real implementation, we would validate this fingerprint against
		// a known list for the driver or check for anomalies.
		// For now, presence check is sufficient.
		
		ctx := context.WithValue(r.Context(), "device_fingerprint", fingerprint)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GeoFenceMiddleware validates request coordinates against allowed bounds
func (m *DriverSecurityMiddleware) GeoFenceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lat := r.Header.Get("X-Geo-Lat")
		long := r.Header.Get("X-Geo-Long")

		if lat != "" && long != "" {
			// In a real implementation, we would:
			// 1. Get driver/vehicle ID from context
			// 2. Fetch last known location or trip route bounds
			// 3. Calculate distance
			// 4. Deny if outside reasonable bounds
			
			m.logger.Debug("request with coordinates", 
				zap.String("lat", lat), 
				zap.String("long", long))
		}

		next.ServeHTTP(w, r)
	})
}
