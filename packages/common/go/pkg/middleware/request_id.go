package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists (from client)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}
		
		// Add request ID to context
		c.Set("request_id", requestID)
		
		// Add request ID to response headers
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}