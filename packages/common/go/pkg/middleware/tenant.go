package middleware

import (
	"context"
	"net/http"
)

// TenantMiddleware extracts the X-Org-ID header and injects it into the context
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID := r.Header.Get("X-Org-ID")
		if orgID == "" {
			// In a real production scenario, we might want to return 401/403
			// for now, we'll just continue and RLS will fail if strict
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetOrgID extracts the OrgID from the context
func GetOrgID(ctx context.Context) string {
	if orgID, ok := ctx.Value(OrgIDKey).(string); ok {
		return orgID
	}
	return ""
}
