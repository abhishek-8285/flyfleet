package middleware

import (
	"net/http"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core/internal/pkg/permissions"
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

// RequirePermissions explicitly asserts that the requester has ALL the provided Bitmask Permissions
func RequirePermissions(logger *zap.Logger, required ...permissions.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			maskFloat, ok := r.Context().Value(ctxkeys.PermissionMaskKey).(float64)
			var userMask permissions.Permission = 0
			if ok {
				userMask = permissions.Permission(maskFloat)
			} else {
				// Also handle natively parsed uint64 if the JWT middleware passes it that way
				maskInt, okInt := r.Context().Value(ctxkeys.PermissionMaskKey).(uint64)
				if okInt {
					userMask = permissions.Permission(maskInt)
				}
			}

			// Deny access if mask is completely missing or empty
			if userMask == 0 {
				logger.Warn("Access denied: missing or empty permission mask in context")
				http.Error(w, "Forbidden: Missing Permissions", http.StatusForbidden)
				return
			}

			// Validate all required permissions
			for _, p := range required {
				if !permissions.HasPermission(userMask, p) {
					logger.Warn("Access denied: insufficient permission bit",
						zap.Uint64("user_mask", uint64(userMask)),
						zap.Uint64("missing_bit", uint64(p)))
					http.Error(w, "Forbidden: Insufficient Permissions", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
