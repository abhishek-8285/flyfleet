package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

// TenantMiddleware extracts the org_id from header or JWT context and injects it.
// Permissive mode — allows through even without org_id (for backward compatibility).
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID := resolveOrgID(r)
		if orgID == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StrictTenantMiddleware rejects any request that does not carry a valid org_id.
// Use this on routes where multi-tenant isolation is mandatory.
func StrictTenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID := resolveOrgID(r)
		if orgID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "tenant_required",
				"message": "Organization context is required for this operation",
			})
			return
		}

		ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolveOrgID checks multiple sources for the org_id: header, then JWT context.
func resolveOrgID(r *http.Request) string {
	// Priority 1: X-Organization-ID header (set by BFF)
	if orgID := r.Header.Get("X-Organization-ID"); orgID != "" {
		return orgID
	}
	// Priority 2: X-Org-ID header (legacy)
	if orgID := r.Header.Get("X-Org-ID"); orgID != "" {
		return orgID
	}
	// Priority 3: Already in context (from AuthMiddleware JWT extraction)
	if orgID, ok := r.Context().Value(OrgIDKey).(string); ok && orgID != "" {
		return orgID
	}
	return ""
}

// GetOrgID extracts the OrgID from the context
func GetOrgID(ctx context.Context) string {
	if orgID, ok := ctx.Value(OrgIDKey).(string); ok {
		return orgID
	}
	return ""
}

// MustGetOrgID extracts the OrgID from context and panics if missing.
// Use only in handlers protected by StrictTenantMiddleware.
func MustGetOrgID(ctx context.Context) string {
	orgID := GetOrgID(ctx)
	if orgID == "" {
		panic("org_id missing from context — StrictTenantMiddleware was not applied")
	}
	return orgID
}
