package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP extracts the client IP from various headers or falls back to RemoteAddr with validation
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from reverse proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			// Validate and sanitize the IP
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				return parsedIP.String()
			}
		}
	}
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		// Validate and sanitize the IP
		if parsedIP := net.ParseIP(realIP); parsedIP != nil {
			return parsedIP.String()
		}
	}
	// Fall back to RemoteAddr
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		ip := addr[:idx]
		// Validate and sanitize the IP
		if parsedIP := net.ParseIP(ip); parsedIP != nil {
			return parsedIP.String()
		}
		return "0.0.0.0" // Default fallback for invalid IPs
	}
	// Validate the full address if no port separator found
	if parsedIP := net.ParseIP(addr); parsedIP != nil {
		return parsedIP.String()
	}
	return "0.0.0.0" // Default fallback for invalid IPs
}