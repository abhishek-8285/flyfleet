package middleware

import (
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"go.uber.org/zap"
)

// RequireRoleForMutations enforces that mutating requests (POST, PUT, PATCH, DELETE)
// are only allowed by users with one of the specified roles. Read operations (GET, HEAD, OPTIONS)
// are allowed for any authenticated user.
func RequireRoleForMutations(logger *zap.Logger, allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow safe methods without role check
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// For mutating methods, check role
			role, ok := r.Context().Value(ctxkeys.RoleKey).(string)
			if !ok {
				logger.Warn("missing role for mutating request",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Check if user has one of the allowed roles
			for _, allowedRole := range allowedRoles {
				if strings.EqualFold(role, allowedRole) {
					next.ServeHTTP(w, r)
					return
				}
			}

			logger.Warn("role not authorized for mutating request",
				zap.String("role", role),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
		})
	}
}
