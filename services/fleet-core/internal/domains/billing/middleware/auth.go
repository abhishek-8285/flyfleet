package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	OrgIDKey  contextKey = "tenant_id"
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid token format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid claims", http.StatusUnauthorized)
				return
			}

			orgIDStr, _ := claims["tenant_id"].(string)
			userIDStr, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)

			ctx := r.Context()
			if orgIDStr != "" {
				orgID, _ := uuid.Parse(orgIDStr)
				ctx = context.WithValue(ctx, OrgIDKey, orgID)
			}
			if userIDStr != "" {
				userID, _ := uuid.Parse(userIDStr)
				ctx = context.WithValue(ctx, UserIDKey, userID)
			}
			ctx = context.WithValue(ctx, RoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetOrgID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(OrgIDKey).(uuid.UUID)
	return id, ok
}
