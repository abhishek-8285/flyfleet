package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment variable (comma-separated)
		// Default to localhost:3000 for development
		originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
		if strings.TrimSpace(originsEnv) == "" {
			originsEnv = "http://localhost:3000"
		}

		allowedOrigins := strings.Split(originsEnv, ",")
		requestOrigin := r.Header.Get("Origin")

		// Check if request origin is in the allowlist
		var originAllowed string
		for _, origin := range allowedOrigins {
			origin = strings.TrimSpace(origin)
			if origin == requestOrigin {
				originAllowed = origin
				break
			}
		}

		// If no match found and we have an Origin header, deny CORS
		if requestOrigin != "" && originAllowed == "" {
			// Don't set CORS headers - browser will block
			next.ServeHTTP(w, r)
			return
		}

		// Set CORS headers (either matching origin or if no Origin header, don't set)
		if originAllowed != "" {
			w.Header().Set("Access-Control-Allow-Origin", originAllowed)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Org-ID, X-Organization-ID, X-Requested-With, X-Trace-Id, X-Request-Id")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
