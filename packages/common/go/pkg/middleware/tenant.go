package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

// TenantMiddleware extracts the tenant_id from header or JWT context and injects it.
// Permissive mode — allows through even without tenant_id (for backward compatibility).
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := resolveTenantID(r)
		if tenantID == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StrictTenantMiddleware rejects any request that does not carry a valid tenant_id.
// Use this on routes where multi-tenant isolation is mandatory.
func StrictTenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := resolveTenantID(r)
		if tenantID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "tenant_required",
				"message": "Tenant context is required for this operation",
			})
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolveTenantID checks multiple sources for the tenant_id: header, then JWT context.
func resolveTenantID(r *http.Request) string {
	// Priority 1: X-Tenant-ID header (modern)
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	// Priority 2: X-Organization-ID header (set by BFF)
	if tenantID := r.Header.Get("X-Organization-ID"); tenantID != "" {
		return tenantID
	}
	// Priority 3: X-Org-ID header (legacy)
	if tenantID := r.Header.Get("X-Org-ID"); tenantID != "" {
		return tenantID
	}
	// Priority 4: Already in context (from AuthMiddleware JWT extraction)
	if tenantID, ok := r.Context().Value(TenantIDKey).(string); ok && tenantID != "" {
		return tenantID
	}
	return ""
}

// GetTenantID extracts the TenantID from the context
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// MustGetTenantID extracts the TenantID from context and panics if missing.
// Use only in handlers protected by StrictTenantMiddleware.
func MustGetTenantID(ctx context.Context) string {
	tenantID := GetTenantID(ctx)
	if tenantID == "" {
		panic("tenant_id missing from context — StrictTenantMiddleware was not applied")
	}
	return tenantID
}
