package middleware

import (
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// IPFirewallMiddleware blocks restricted IP addresses based on rules from auth-service
func IPFirewallMiddleware(fm *FirewallManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if fm.IsBlocked(clientIP) {
				fm.logger.Warn("IP blocked by firewall policy", zap.String("ip", clientIP), zap.String("path", r.URL.Path))
				http.Error(w, "Access Denied: Your IP has been restricted by security policy.", http.StatusForbidden)
				return
			}

			// Placeholder for "Emergency Block" header (for verification/testing)
			if r.Header.Get("X-Firewall-Block") == "true" {
				fm.logger.Warn("IP manually blocked by emergency header", zap.String("ip", clientIP))
				http.Error(w, "Access Denied: Your IP has been restricted by security policy.", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	// 1. Check X-Forwarded-For (from proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	// 2. Check X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// 3. Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
