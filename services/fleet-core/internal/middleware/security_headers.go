package middleware

import (
	"net/http"
	"os"
	"strings"
)

// SecurityHeaders sets important HTTP security headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict-Transport-Security - Force HTTPS for 1 year including subdomains
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Determine environment (production vs development)
		env := strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
		isProduction := env == "production" || env == "staging" || env == ""

		// Content-Security-Policy - Prevent XSS by restricting resources
		if isProduction {
			// Production: Strict CSP without 'unsafe-inline' for scripts
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
		} else {
			// Development: Allow unsafe-inline and unsafe-eval for HMR, React DevTools, etc.
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
		}

		// X-Frame-Options - Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options - Prevent MIME-type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Referrer-Policy - Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy - Restrict browser features (geolocation, camera, microphone, etc.)
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		next.ServeHTTP(w, r)
	})
}
