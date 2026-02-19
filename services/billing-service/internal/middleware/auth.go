package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/flyfleet/billing-service/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	OrgIDKey  contextKey = "org_id"
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteProblem(w, errors.Unauthorized("missing authorization header"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				errors.WriteProblem(w, errors.Unauthorized("invalid token format"))
				return
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil || !token.Valid {
				errors.WriteProblem(w, errors.Unauthorized("invalid token"))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				errors.WriteProblem(w, errors.Unauthorized("invalid claims"))
				return
			}

			orgIDStr, _ := claims["org_id"].(string)
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

