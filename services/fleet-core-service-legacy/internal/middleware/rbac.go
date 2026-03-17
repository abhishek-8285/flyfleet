package middleware

import (
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"go.uber.org/zap"
)

// RequireRole ensures the user has one of the required roles
func RequireRole(logger *zap.Logger, requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(ctxkeys.RoleKey).(string)
			if !ok || role == "" {
				logger.Warn("Access denied: no role found in context")
				http.Error(w, "Forbidden: No Role Assigned", http.StatusForbidden)
				return
			}

			allowed := false
			for _, rr := range requiredRoles {
				if role == rr {
					allowed = true
					break
				}
			}

			if !allowed {
				logger.Warn("Access denied: insufficient role", 
					zap.String("user_role", role), 
					zap.Any("required_roles", requiredRoles))
				http.Error(w, "Forbidden: Insufficient Permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
