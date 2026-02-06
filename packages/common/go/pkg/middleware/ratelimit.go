package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RateLimiter struct {
	limit    int
	window   time.Duration
	requests map[string][]time.Time
	mutex    sync.Mutex
	logger   *zap.Logger
}

func NewRateLimiter(limit int, window time.Duration, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
		logger:   logger,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		rl.mutex.Lock()
		
		// Clean old requests outside the window
		if requests, exists := rl.requests[clientIP]; exists {
			validRequests := make([]time.Time, 0)
			for _, reqTime := range requests {
				if now.Sub(reqTime) <= rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[clientIP] = validRequests
		} else {
			rl.requests[clientIP] = make([]time.Time, 0)
		}
		
		// Check if limit exceeded
		if len(rl.requests[clientIP]) >= rl.limit {
			rl.mutex.Unlock()
			
			rl.logger.Warn("Rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.Int("limit", rl.limit),
				zap.Duration("window", rl.window))
			
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			return
		}
		
		// Add current request
		rl.requests[clientIP] = append(rl.requests[clientIP], now)
		remaining := rl.limit - len(rl.requests[clientIP])
		
		rl.mutex.Unlock()
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", now.Add(rl.window).Format(time.RFC3339))
		
		c.Next()
	}
}