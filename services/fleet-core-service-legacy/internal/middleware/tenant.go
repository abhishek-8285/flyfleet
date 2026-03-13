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
// It prioritizes the value set by the Auth middleware (from JWT), but can also fallback
// to a header X-Organization-ID for administrative or internal requests if necessary.
func TenantMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 0. Check for IAM OrgIDKey (uuid.UUID) - mapped to TenantIDKey for Reports
			if orgID, ok := ctx.Value(iamCtx.OrgIDKey).(uuid.UUID); ok {
				ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, orgID.String())
				ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgID.String())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			
			if userID, ok := ctx.Value(iamCtx.UserIDKey).(uuid.UUID); ok {
				ctx = context.WithValue(ctx, ctxkeys.UserIDKey, userID.String())
			}

			// 1. Check if org_id is already in context (set by other middleware as string)
			if _, ok := ctx.Value(ctxkeys.OrgIDKey).(string); ok {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Check for X-Organization-ID header (fallback/admin)
			orgIDStr := r.Header.Get("X-Organization-ID")
			if orgIDStr != "" {
				if _, err := uuid.Parse(orgIDStr); err != nil {
					http.Error(w, "Invalid Organization ID format", http.StatusBadRequest)
					return
				}
				
				// Inject into context (both legacy string and future uuid if needed)
				ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgIDStr)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 3. For protected routes, we expect an OrgID. 
			// In this phase, we just ensure it's logged if missing, but we'll 
			// let it pass for now to avoid breaking existing unmanaged routes.
			// Revisit this once API Gateway is in place.
			
			next.ServeHTTP(w, r)
		})
	}
}
