package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	iamCtx "github.com/flyfleet/fleet-core/internal/platform/iam/middleware"
	"github.com/google/uuid"
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

			// Extract OrgID from JWT via auth middleware (IAM or legacy)
			var orgID uuid.UUID
			var ok bool

			// Try IAM context first (UUID)
			if orgID, ok = ctx.Value(iamCtx.OrgIDKey).(uuid.UUID); !ok {
				// Try string context fallback
				if orgIDStr, okStr := ctx.Value(ctxkeys.OrgIDKey).(string); okStr {
					if id, err := uuid.Parse(orgIDStr); err == nil {
						orgID = id
						ok = true
					}
				}
			} else {
				ok = true
			}

			if !ok {
				// No valid org_id from authentication - reject request
				logger.Warn("organization context missing for authenticated request",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Organization context required", http.StatusUnauthorized)
				return
			}

			// Ensure org_id is available in both formats for compatibility
			ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, orgID.String())
			ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgID.String())
			ctx = context.WithValue(ctx, iamCtx.OrgIDKey, orgID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
