package middleware

import (
	"net/http"
)

func CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3006",
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

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true") // Often needed when specific origins are used
		} else {
			// If origin is not in the allowed list, you might choose to not set the header
			// or set a default, or return an error. For now, we'll just not set it.
			// If you want to allow all for non-listed origins, you could set it to "*" here,
			// but that defeats the purpose of a specific list.
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
