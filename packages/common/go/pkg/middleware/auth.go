package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is a dedicated type for context keys to avoid collisions
type ContextKey string

const (
	UserIDKey ContextKey = "user_id"
	OrgIDKey  ContextKey = "org_id"
	RoleKey   ContextKey = "role"
)

// SecretProvider defines an interface for retrieving the JWT secret.
type SecretProvider interface {
	GetJWTSecret() string
}

// AuthMiddleware validates the JWT token and injects claims into the request context.
func AuthMiddleware(sp SecretProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtSecret := sp.GetJWTSecret()
			if jwtSecret == "" {
				http.Error(w, "JWT secret not configured", http.StatusInternalServerError)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Verify token signature and claims
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Token verification failed", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()

			// Inject user_id into context
			if userIDStr, ok := claims["user_id"].(string); ok {
				ctx = context.WithValue(ctx, UserIDKey, userIDStr)
			}

			// Inject org_id into context
			if orgIDStr, ok := claims["org_id"].(string); ok {
				ctx = context.WithValue(ctx, OrgIDKey, orgIDStr)
			}

			// Inject role into context
			if role, ok := claims["role"].(string); ok {
				ctx = context.WithValue(ctx, RoleKey, role)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
