package middleware

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
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

// GetOrgID extracts the organization UUID from the request context.
func GetOrgID(r *http.Request) uuid.UUID {
	v := r.Context().Value(CtxOrgID)
	if v == nil {
		return uuid.Nil
	}
	return v.(uuid.UUID)
}

// GetUserID extracts the user UUID from the request context.
func GetUserID(r *http.Request) uuid.UUID {
	v := r.Context().Value(CtxUserID)
	if v == nil {
		return uuid.Nil
	}
	return v.(uuid.UUID)
}

// GetUserRoles extracts the user's roles from the request context.
func GetUserRoles(r *http.Request) []string {
	v := r.Context().Value(CtxUserRoles)
	if v == nil {
		return nil
	}
	return v.([]string)
}
