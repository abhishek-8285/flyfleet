package middleware

import (
	"context"
	"net/http"
	"strings"

	sharedmiddleware "github.com/fleet-platform/shared/pkg/middleware"
	billingMiddleware "github.com/flyfleet/fleet-core/internal/domains/billing/middleware"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"go.uber.org/zap"
)

// TenantMiddleware ensures that all requests have a valid organization_id in the context.
// SECURITY: The organization ID MUST come from the authenticated JWT token set by the
// Auth middleware. Client-supplied headers are NOT trusted to prevent org spoofing.
func TenantMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Skip tenant check for global admin or identity endpoints
			path := r.URL.Path
			if strings.HasPrefix(path, "/api/v1/admin/organizations") ||
				strings.HasPrefix(path, "/api/v1/auth/profile") ||
				strings.HasPrefix(path, "/api/v1/auth/orgs") ||
				strings.HasPrefix(path, "/api/v1/dashboard/stats") {
				next.ServeHTTP(w, r)
				return
			}

			// Extract TenantID from context (populated by Auth middleware)
			tenantID, ok := ctxkeys.GetTenantID(ctx)

			if !ok {
				// No valid tenant_id from authentication - check for header as fallback ONLY for logging purposes
				// so we can see what the client was TRYING to do.
				headerTenantID := r.Header.Get("X-Tenant-ID")
				logger.Warn("organization context missing for authenticated request",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("header_tenant_id", headerTenantID))
				
				http.Error(w, "Organization context required", http.StatusUnauthorized)
				return
			}

			// Ensure tenant_id is available in standardized format for all handlers.
			// We set it using the typed contextKey to avoid any shadowing.
			ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, tenantID)
			
			// Also ensure it is in the raw string format for handlers that might still rely on it.
			ctx = context.WithValue(ctx, "tenant_id", tenantID.String())
			
			// Shared Library Compatibility
			ctx = context.WithValue(ctx, sharedmiddleware.TenantIDKey, tenantID)
			
			// Billing Domain Compatibility
			ctx = context.WithValue(ctx, billingMiddleware.TenantIDKey, tenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
