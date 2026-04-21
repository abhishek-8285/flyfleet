package middleware

import (
	"net/http"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"go.uber.org/zap"
)

// TenantContextMiddleware extracts tenant_id from the request context
// and ensures it is present for protected routes.
func TenantContextMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip check for super_admin roles
			if role := ctxkeys.GetRole(r.Context()); role == "super_admin" {
				next.ServeHTTP(w, r)
				return
			}

			// tenant_id is already in context from JWKSAuthGate/TenantMiddleware
			_, ok := ctxkeys.GetTenantID(r.Context())

			if !ok {
				logger.Warn("tenant context missing in TenantContextMiddleware", zap.String("path", r.URL.Path))
				http.Error(w, "tenant context required", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserRoles extracts the user's roles from the request context.
func GetUserRoles(r *http.Request) []string {
	v := r.Context().Value(ctxkeys.RolesKey)
	if v == nil {
		return nil
	}
	roles, _ := v.([]string)
	return roles
}
