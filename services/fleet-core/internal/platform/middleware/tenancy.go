package middleware

import (
	"context"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	iamCtx "github.com/flyfleet/fleet-core/internal/platform/iam/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantContextMiddleware extracts org_id from the context (set by Auth middleware)
// and ensures it is properly formatted and available for Repository RLS enforcement.
// SECURITY: This middleware ONLY uses org_id from the authenticated JWT token.
// Client-supplied headers like X-Organization-ID are IGNORED to prevent spoofing.
// Super_admin users bypass tenant isolation for system-wide access.
func TenantContextMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Check for super_admin role bypass - system-wide access
			if role, ok := ctx.Value(ctxkeys.RoleKey).(string); ok && role == "super_admin" {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

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
				// OrgID is required for all protected routes - fail securely
				logger.Warn("tenant context missing for authenticated request",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Organization context required", http.StatusForbidden)
				return
			}

			// Ensure it's available in both formats for wide compatibility
			ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgID.String())
			ctx = context.WithValue(ctx, iamCtx.OrgIDKey, orgID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetOrganizationID retrieves the organization ID from the context.
// Returns uuid.Nil if not found.
func GetOrganizationID(ctx context.Context) uuid.UUID {
	// Try the IAM context key first
	if orgID, ok := ctx.Value(iamCtx.OrgIDKey).(uuid.UUID); ok {
		return orgID
	}
	// Fallback to legacy string format
	if orgIDStr, ok := ctx.Value(ctxkeys.OrgIDKey).(string); ok {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			return id
		}
	}
	return uuid.Nil
}
