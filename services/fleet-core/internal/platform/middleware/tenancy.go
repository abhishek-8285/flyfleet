package middleware

import (
	"context"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantContextMiddleware extracts tenant_id from the context (set by Auth middleware)
// and ensures it is properly formatted and available for Repository RLS enforcement.
// SECURITY: This middleware ONLY uses tenant_id from the authenticated JWT token.
// Client-supplied headers like X-Organization-ID are IGNORED to prevent spoofing.
// Super_admin users bypass tenant isolation for system-wide access.
func TenantContextMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Check for super_admin role bypass - system-wide access
			if role, ok := ctx.Value(ctxkeys.RoleKey).(string); ok && role == "super_admin" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Extract TenantID from context (set by JWKSAuthGate)
			var tenantID uuid.UUID
			var ok bool

			if tenantIDVal := ctx.Value(ctxkeys.TenantIDKey); tenantIDVal != nil {
				// Try as string first (standard in jwks_auth)
				if tenantIDStr, isStr := tenantIDVal.(string); isStr {
					if id, err := uuid.Parse(tenantIDStr); err == nil {
						tenantID = id
						ok = true
					}
				} else if id, isUUID := tenantIDVal.(uuid.UUID); isUUID {
					// Fallback for types that store as UUID
					tenantID = id
					ok = true
				}
			}

			if !ok {
				// TenantID is required for all protected routes - fail securely
				logger.Warn("tenant context missing for authenticated request",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Organization context required", http.StatusForbidden)
				return
			}

			// Ensure it's available in both formats for wide compatibility
			// Ensure it's available in standard string format for wide compatibility
			ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, tenantID.String())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

