package middleware

import (
	"context"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/platform/iam/middleware"
	"github.com/google/uuid"
)

func TenantContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, ok := r.Context().Value(middleware.OrgIDKey).(uuid.UUID)
			if !ok || orgID == uuid.Nil {
				http.Error(w, "missing organization context", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetOrganizationID(ctx context.Context) uuid.UUID {
	orgID, ok := ctx.Value(middleware.OrgIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return orgID
}
