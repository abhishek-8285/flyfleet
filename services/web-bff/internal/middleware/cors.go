package middleware

import (
"os"
	"net/http"
)

func CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := []string{
		os.Getenv("FRONTEND_URL"),
		os.Getenv("ADMIN_FRONTEND_URL"),
		os.Getenv("MOBILE_APP_URL"),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				break
			}
		}

		// Clear potentially conflicting headers from proxy
		w.Header().Del("Access-Control-Allow-Origin")
		w.Header().Del("Access-Control-Allow-Credentials")
		w.Header().Del("Access-Control-Allow-Methods")
		w.Header().Del("Access-Control-Allow-Headers")

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// Default for local development if origin is empty or not matched but we suspect it should be 3000
			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-ID, X-Organization-ID")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
