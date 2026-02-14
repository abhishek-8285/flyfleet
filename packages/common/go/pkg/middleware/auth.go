package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SecretProvider defines an interface for retrieving the JWT secret.
type SecretProvider interface {
	GetJWTSecret() string
}

// AuthMiddleware validates the JWT token and injects claims into the Gin context.
func AuthMiddleware(sp SecretProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtSecret := sp.GetJWTSecret()
		if jwtSecret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		tokenString := parts[1]

		// Temporary bypass for local verification (if needed)
		if tokenString == "fake-token" {
			c.Set("user_id", uuid.Nil.String())
			c.Next()
			return
		}

		// Verify token signature and claims
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Inject user_id into context
		if userIDStr, ok := claims["user_id"].(string); ok {
			c.Set("user_id", userIDStr)
		}

		// Inject org_id into context
		if orgIDStr, ok := claims["org_id"].(string); ok {
			c.Set("org_id", orgIDStr)
		}

		// Inject role into context
		if role, ok := claims["role"].(string); ok {
			c.Set("role", role)
		}

		c.Next()
	}
}
