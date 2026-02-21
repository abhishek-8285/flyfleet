package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/fleet-platform/shared/pkg/serviceauth"
	"go.uber.org/zap"
)

type serviceCtxKey struct{}

// ServiceIDFromContext extracts the calling service identity from the
// request context.  Returns "" if the request was not authenticated
// via the S2S middleware.
func ServiceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(serviceCtxKey{}).(string); ok {
		return v
	}
	return ""
}

// ServiceAuthMiddleware verifies incoming requests carry a valid
// Service-to-Service JWT in the "X-Service-Token" header.
// Requests that fail verification are rejected with 403 Forbidden.
func ServiceAuthMiddleware(verifier *serviceauth.TokenVerifier, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := strings.TrimSpace(r.Header.Get("X-Service-Token"))
			if tokenStr == "" {
				logger.Warn("Missing X-Service-Token header",
					zap.String("path", r.URL.Path),
					zap.String("remote", r.RemoteAddr),
				)
				http.Error(w, "Forbidden: missing service token", http.StatusForbidden)
				return
			}

			claims, err := verifier.Verify(tokenStr)
			if err != nil {
				logger.Warn("Invalid service token",
					zap.String("path", r.URL.Path),
					zap.Error(err),
				)
				http.Error(w, "Forbidden: invalid service token", http.StatusForbidden)
				return
			}

			logger.Debug("Service request authenticated",
				zap.String("service_id", claims.ServiceID),
				zap.String("path", r.URL.Path),
			)

			ctx := context.WithValue(r.Context(), serviceCtxKey{}, claims.ServiceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
