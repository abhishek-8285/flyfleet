package middleware

import (
	"net/http"
)

// TenantContextMiddleware extracts org_id from the request context
// (previously set by JWKSAuthGate) and makes it available via
// GetOrgID helper. This is a pass-through since the JWT already
// contains the org claim — it just provides a convenient access point.
func TenantContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// org_id is already in context from JWKSAuthGate
		// This middleware exists as a named layer for clarity
		// and future extension (e.g., setting DB search_path).
		next.ServeHTTP(w, r)
	})
}

// GetUserRoles extracts the user's roles from the request context.
func GetUserRoles(r *http.Request) []string {
	v := r.Context().Value(CtxUserRoles)
	if v == nil {
		return nil
	}
	return v.([]string)
}
