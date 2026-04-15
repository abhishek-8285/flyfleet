package middleware

import (
	"net/http"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	iamCtx "github.com/flyfleet/fleet-core/internal/platform/iam/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantContextMiddleware extracts org_id from the request context
// and ensures it is present for protected routes.
func TenantContextMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip check for super_admin roles
			if role, ok := r.Context().Value(ctxkeys.RoleKey).(string); ok && role == "super_admin" {
				next.ServeHTTP(w, r)
				return
			}

			// org_id is already in context from JWKSAuthGate/TenantMiddleware
			_, ok := r.Context().Value(iamCtx.OrgIDKey).(uuid.UUID)
			if !ok {
				// Fallback: check string key
				if orgIDStr, okStr := r.Context().Value(ctxkeys.OrgIDKey).(string); okStr {
					if _, err := uuid.Parse(orgIDStr); err == nil {
						ok = true
					}
				}
			}

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
	v := r.Context().Value(CtxUserRoles)
	if v == nil {
		return nil
	}
	return v.([]string)
}
