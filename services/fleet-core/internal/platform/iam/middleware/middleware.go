package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core/internal/platform/iam/service"
)


func AuthMiddleware(iamSvc service.IAMService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			claims, err := iamSvc.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Check session revocation — if the token's JTI is no longer active, reject
			if claims.JTI != "" {
				active, err := iamSvc.IsSessionActive(r.Context(), claims.UserID, claims.TenantID, claims.JTI)
				if err == nil && !active {
					http.Error(w, "session revoked", http.StatusUnauthorized)
					return
				}
			}

			// Store UUIDs directly to maintain type safety
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxkeys.UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, claims.TenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RBACMiddleware(iamSvc service.IAMService, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use standardized extraction
			userID, ok := ctxkeys.GetUserID(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			allowed, err := iamSvc.CheckPermission(r.Context(), userID, resource, action)
			if err != nil || !allowed {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
