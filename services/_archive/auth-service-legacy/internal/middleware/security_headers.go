package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeadersMiddleware adds important security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME-type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Prevent pages from being loaded in frames (clickjacking protection)
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Enable XSS protection in browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Prevent referrer header from being sent to external sites
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy
		csp := "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';"
		w.Header().Set("Content-Security-Policy", csp)
		
		// Strict Transport Security (HSTS) - only for production
		if strings.Contains(r.Host, ".com") || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// Disable caching for sensitive endpoints
		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") || 
		   strings.HasPrefix(r.URL.Path, "/api/v1/admin/") ||
		   strings.HasPrefix(r.URL.Path, "/api/v1/security/") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		
		next.ServeHTTP(w, r)
	})
}