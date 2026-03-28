package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter defines the interface for rate limiting logic
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// RedisRateLimiter implements RateLimiter using a Redis-backed fixed window counter
type RedisRateLimiter struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisRateLimiter(client *redis.Client, logger *zap.Logger) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		logger: logger,
	}
}

// Allow checks if a request should be allowed based on a key and limit/window
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if r.client == nil {
		return true, nil // Fail open if Redis is down (Series-B operational guideline)
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("redis rate limit execution failed", zap.Error(err), zap.String("key", key))
		return true, nil // Fail open
	}

	return incr.Val() <= int64(limit), nil
}

// Handler returns a middleware that enforces rate limits
func (r *RedisRateLimiter) Handler(keyPrefix string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Extract IP securely. In a real setup, validate X-Forwarded-For against a trusted proxy list.
			ip := req.RemoteAddr
			if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
				ip = host
			}

			key := fmt.Sprintf("%s:%s", keyPrefix, ip)

			allowed, err := r.Allow(req.Context(), key, limit, window)
			if err != nil || !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "too_many_requests",
					"message": "Rate limit exceeded. Please try again later.",
				})
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
