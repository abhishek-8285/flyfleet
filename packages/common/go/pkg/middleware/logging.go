package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Log request details
		duration := time.Since(start)
		
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("duration", duration),
		}
		
		// Add request ID if present
		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}
		
		// Add user ID if authenticated
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok {
				fields = append(fields, zap.String("user_id", id))
			}
		}
		
		// Log based on status code
		if len(c.Errors) > 0 {
			fields = append(fields, zap.Strings("errors", c.Errors.Errors()))
			logger.Error("HTTP request failed", fields...)
		} else if c.Writer.Status() >= 500 {
			logger.Error("HTTP server error", fields...)
		} else if c.Writer.Status() >= 400 {
			logger.Warn("HTTP client error", fields...)
		} else {
			logger.Info("HTTP request completed", fields...)
		}
	}
}