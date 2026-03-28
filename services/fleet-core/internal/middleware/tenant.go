package middleware

import (
	"context"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	iamCtx "github.com/flyfleet/fleet-core-service/internal/platform/iam/middleware"
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

			// Extract OrgID from JWT via auth middleware (IAM or legacy)
			var orgID uuid.UUID
			var ok bool

			// Try IAM context first
			if orgID, ok = ctx.Value(iamCtx.OrgIDKey).(uuid.UUID); ok {
				// Found org_id from authenticated JWT
			} else if orgIDStr, ok := ctx.Value(ctxkeys.OrgIDKey).(string); ok {
				// Fallback to legacy string from auth middleware
				if id, err := uuid.Parse(orgIDStr); err == nil {
					orgID = id
					ok = true
				}
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
