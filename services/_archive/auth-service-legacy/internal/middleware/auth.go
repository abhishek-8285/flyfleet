package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/fleet-platform/shared/pkg/security"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthMiddleware validates the JWT token and injects identity claims into context.
// Unlike RBACMiddleware, it does NOT enforce specific resource permissions.
func AuthMiddleware(sp security.SecretProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtSecret := sp.GetJWTSecret()

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

			// Remove temporary bypass for local verification - this is a security vulnerability
			// The following lines were commented out but should be completely removed
			// if tokenString == "fake-token" {
			// 	next.ServeHTTP(w, r)
			// 	return
			// }

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

			// Accept canonical `sub` and legacy `user_id` claims.
			userIDStr, _ := claims["sub"].(string)
			if userIDStr == "" {
				userIDStr, _ = claims["user_id"].(string)
			}
			if userIDStr == "" {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			uid, err := uuid.Parse(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", uid)

			if orgIDStr, ok := claims["org_id"].(string); ok {
				if orgID, err := uuid.Parse(orgIDStr); err == nil {
					ctx = context.WithValue(ctx, "org_id", orgID)
				}
			}
			if role, ok := claims["role"].(string); ok && role != "" {
				ctx = context.WithValue(ctx, "role", role)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
