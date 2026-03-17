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
func TenantContextMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract OrgID from context (set by JWKSAuthGate)
			orgID, ok := ctx.Value(iamCtx.OrgIDKey).(uuid.UUID)
			if !ok {
				// Fallback to legacy string if needed
				if orgIDStr, ok := ctx.Value(ctxkeys.OrgIDKey).(string); ok {
					if id, err := uuid.Parse(orgIDStr); err == nil {
						orgID = id
						ok = true
					}
				}
			}

			if !ok {
				// If this is a protected route (which should have been caught by AuthGate),
				// and we still don't have an OrgID, we might want to block it here.
				logger.Warn("tenant context missing for request", zap.String("path", r.URL.Path))
			} else {
				headerOrgIDStr := r.Header.Get("X-Org-ID")
				if headerOrgIDStr != "" && headerOrgIDStr != orgID.String() {
					logger.Warn("tenant-safe violation: header org_id does not match token", 
						zap.String("header_org_id", headerOrgIDStr),
						zap.String("token_org_id", orgID.String()),
						zap.String("path", r.URL.Path))
					http.Error(w, "Forbidden: Security Context Mismatch", http.StatusForbidden)
					return
				}

				// Ensure it's available in both formats for wide compatibility
				ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgID.String())
				ctx = context.WithValue(ctx, iamCtx.OrgIDKey, orgID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
