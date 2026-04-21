package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter defines the interface for rate limiting logic
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// RedisRateLimiter implements RateLimiter using a sliding window algorithm in Redis
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

// Allow checks if a request should be allowed based on a sliding window algorithm
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// Lua script for atomic sliding window rate limiting
	// KEYS[1] - Rate limit key
	// ARGV[1] - Window in milliseconds
	// ARGV[2] - Limit
	// ARGV[3] - Current timestamp in milliseconds
	script := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local min = now - window

		-- Remove old entries
		redis.call('ZREMRANGEBYSCORE', key, 0, min)

		-- Count current entries
		local current = redis.call('ZCARD', key)

		if current < limit then
			redis.call('ZADD', key, now, now)
			redis.call('PEXPIRE', key, window)
			return 1
		else
			return 0
		end
	`

	now := time.Now().UnixNano() / int64(time.Millisecond)
	windowMs := int64(window / time.Millisecond)

	res, err := r.client.Eval(ctx, script, []string{key}, windowMs, limit, now).Int()
	if err != nil {
		r.logger.Error("Rate limiter Lua script failed", zap.String("key", key), zap.Error(err))
		return true, nil // Fail open to avoid blocking users
	}

	return res == 1, nil
}

// FixedRateLimit returns a middleware that enforces a fixed rate limit
func FixedRateLimit(limiter RateLimiter, limit int, window time.Duration, keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			allowed, err := limiter.Allow(r.Context(), "rl:"+key, limit, window)
			if err != nil || !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantRateLimit enforces a per-tenant rate limit
func TenantRateLimit(limiter RateLimiter, limit int, window time.Duration) func(http.Handler) http.Handler {
	return FixedRateLimit(limiter, limit, window, func(r *http.Request) string {
		// Try to get TenantID from context (populated by TenantMiddleware)
		if tenantID, ok := r.Context().Value(TenantIDKey).(string); ok {
			return "tenant:" + tenantID
		}
		// Fallback to IP address
		return "ip:" + r.RemoteAddr
	})
}

// UserRateLimit enforces a per-user rate limit
func UserRateLimit(limiter RateLimiter, limit int, window time.Duration) func(http.Handler) http.Handler {
	return FixedRateLimit(limiter, limit, window, func(r *http.Request) string {
		if userID, ok := r.Context().Value("user_id").(string); ok && userID != "" {
			return "user:" + userID
		}
		return "ip:" + r.RemoteAddr
	})
}

// GlobalRateLimit enforces a global IP-based rate limit
func GlobalRateLimit(limiter RateLimiter, limit int, window time.Duration) func(http.Handler) http.Handler {
	return FixedRateLimit(limiter, limit, window, func(r *http.Request) string {
		return "global:" + r.RemoteAddr
	})
}
