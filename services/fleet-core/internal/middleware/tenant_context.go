package middleware

import (
	"context"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	iamCtx "github.com/flyfleet/fleet-core-service/internal/platform/iam/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantContextMiddleware extracts org_id from the context (set by Auth middleware)
// and ensures it is properly formatted and available for Repository RLS enforcement.
// SECURITY: This middleware ONLY uses org_id from the authenticated JWT token.
// Client-supplied headers like X-Organization-ID are IGNORED to prevent spoofing.
func TenantContextMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract OrgID from context (set by JWKSAuthGate)
			orgID, ok := ctx.Value(iamCtx.OrgIDKey).(uuid.UUID)
			if ok {
				logger.Debug("tenant context found via iamCtx.OrgIDKey", zap.String("org_id", orgID.String()))
			} else {
				// Fallback to legacy string if needed
				if orgIDStr, okStr := ctx.Value(ctxkeys.OrgIDKey).(string); okStr {
					if id, err := uuid.Parse(orgIDStr); err == nil {
						orgID = id
						ok = true
						logger.Debug("tenant context found via fallback ctxkeys.OrgIDKey string", zap.String("org_id", orgID.String()))
					} else {
						logger.Warn("tenant context fallback found invalid org_id string", zap.String("value", orgIDStr))
					}
				}
			}

			// If org_id is missing, check for super_admin role bypass
			if !ok {
				role, _ := ctx.Value(ctxkeys.RoleKey).(string)
				if role == "super_admin" {
					// Super admin can proceed - log that we're bypassing tenant context
					logger.Info("super_admin bypassing tenant context check (no org_id found)",
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method))
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// OrgID is required for all protected routes - fail securely
				logger.Warn("tenant context missing for authenticated request",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("role", role))
				http.Error(w, "Organization context required", http.StatusUnauthorized)
				return
			}

			// Ensure it's available in both formats for wide compatibility
			ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgID.String())
			ctx = context.WithValue(ctx, iamCtx.OrgIDKey, orgID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
