package handlers

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	tenantIDKey  contextKey = "tenant_id"
	userIDKey contextKey = "user_id"
)

// ContextMiddleware provides org/user values expected by the reports handlers.
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))
		if tenantID == "" {
			tenantID = strings.TrimSpace(r.URL.Query().Get("orgId"))
		}
		userID := strings.TrimSpace(r.Header.Get("X-User-ID"))

		ctx := r.Context()
		if tenantID != "" {
			ctx = context.WithValue(ctx, tenantIDKey, tenantID)
		}
		if userID != "" {
			ctx = context.WithValue(ctx, userIDKey, userID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
