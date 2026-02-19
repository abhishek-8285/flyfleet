package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/flyfleet/maintenance-service/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	OrgIDKey  contextKey = "org_id"
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func Auth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteProblem(w, errors.Unauthorized("authorization header missing"))
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				errors.WriteProblem(w, errors.Unauthorized("invalid or expired token"))
				return
			}

			orgIDStr, _ := claims["org_id"].(string)
			userIDStr, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)

			ctx := r.Context()
			if orgID, err := uuid.Parse(orgIDStr); err == nil {
				ctx = context.WithValue(ctx, OrgIDKey, orgID)
			}
			if userID, err := uuid.Parse(userIDStr); err == nil {
				ctx = context.WithValue(ctx, UserIDKey, userID)
			}
			ctx = context.WithValue(ctx, RoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Tenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID, ok := r.Context().Value(OrgIDKey).(uuid.UUID)
		if !ok || orgID == uuid.Nil {
			errors.WriteProblem(w, errors.Forbidden("tenant isolation: org_id missing or invalid"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetOrgID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(OrgIDKey).(uuid.UUID)
	return id, ok
}

