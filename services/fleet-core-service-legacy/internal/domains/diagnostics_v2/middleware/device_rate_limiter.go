package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

type DeviceRateLimiter struct {
	client *redis.Client
	limit  int64
	window time.Duration
}

func NewDeviceRateLimiter(redisAddr string) *DeviceRateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &DeviceRateLimiter{
		client: rdb,
		limit:  500, // 500 requests
		window: 1 * time.Minute,
	}
}

func (rl *DeviceRateLimiter) LimitTelemetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceID := chi.URLParam(r, "id")
		if deviceID == "" {
			http.Error(w, "Device ID required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		key := fmt.Sprintf("rate_limit:telemetry:%s", deviceID)

		// Increment request count
		count, err := rl.client.Incr(ctx, key).Result()
		if err != nil {
			// Fail open on Redis error so telemetry isn't entirely dropped, or fail closed.
			// Defaulting to fail open for telemetry stability.
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			// Set expiration on first request
			rl.client.Expire(ctx, key, rl.window)
		}

		if count > rl.limit {
			http.Error(w, "Rate limit exceeded (Max 500/min)", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
