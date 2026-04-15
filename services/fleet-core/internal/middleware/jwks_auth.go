package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// contextKey is a private type to prevent context key collisions.
type contextKey string

const (
	CtxUserID    contextKey = "user_id"
	CtxOrgID     contextKey = "org_id"
	CtxUserRoles contextKey = "user_roles"
)

// JWKSAuthGate validates the JWT in the Authorization header.
// In production this would fetch and validate against a JWKS endpoint.
// For now it uses HMAC-SHA256 with the shared secret.
func JWKSAuthGate(jwtSecret string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenStr string

			// Try Authorization header first
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
				if tokenStr == authHeader {
					http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
					return
				}
			}

			// Fall back to query param (used for WebSocket connections)
			if tokenStr == "" {
				tokenStr = r.URL.Query().Get("token")
			}

			claims, err := validateToken(tokenStr, jwtSecret)
			if err != nil {
				logger.Debug("jwt validation failed", zap.Error(err))
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxOrgID, claims.OrgID)
			ctx = context.WithValue(ctx, CtxUserRoles, claims.Roles)
			// Set tenant context keys for TenantMiddleware compatibility
			ctx = context.WithValue(ctx, ctxkeys.UserIDKey, claims.UserID.String())
			ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, claims.OrgID.String())
			ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, claims.OrgID.String())
			// Set IAM handler string keys (profile handler reads "user_id" string key)
			ctx = context.WithValue(ctx, "user_id", claims.UserID.String())
			ctx = context.WithValue(ctx, "org_id", claims.OrgID.String())
			ctx = context.WithValue(ctx, "role", claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type Claims struct {
	UserID uuid.UUID `json:"sub"`
	OrgID  uuid.UUID `json:"org_id"`
	Roles  []string  `json:"roles"`
	jwt.RegisteredClaims
}

func validateToken(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ValidateJWTFromQuery extracts and validates a JWT from the URL query
// string. Used for WebSocket handshake where headers can't be set.
func ValidateJWTFromQuery(r *http.Request, jwtSecret string) (*Claims, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		return nil, jwt.ErrTokenMalformed
	}
	return validateToken(tokenStr, jwtSecret)
}
