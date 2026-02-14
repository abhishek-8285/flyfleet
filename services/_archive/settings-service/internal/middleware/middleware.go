package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	UserIPKey contextKey = "user_ip"
)

func UserContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract User ID
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr != "" {
			if uid, err := uuid.Parse(userIDStr); err == nil {
				ctx = context.WithValue(ctx, UserIDKey, uid)
			}
		}

		// Extract IP
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		} else {
			// X-Forwarded-For can be a comma-separated list
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		}
		ctx = context.WithValue(ctx, UserIPKey, ip)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
