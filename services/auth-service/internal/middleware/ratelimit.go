package middleware

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type RateLimiter struct {
	ips    map[string][]time.Time
	mu     sync.Mutex
	limit  int
	window time.Duration
	logger *zap.Logger
}

func NewRateLimiter(limit int, window time.Duration, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		ips:    make(map[string][]time.Time),
		limit:  limit,
		window: window,
		logger: logger,
	}
}

func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		expiry := now.Add(-rl.window)

		// Clean old requests
		var recent []time.Time
		for _, t := range rl.ips[ip] {
			if t.After(expiry) {
				recent = append(recent, t)
			}
		}

		if len(recent) >= rl.limit {
			rl.logger.Warn("Rate limit exceeded", zap.String("ip", ip), zap.String("path", r.URL.Path))
			http.Error(w, `{"error": "Too many requests. Please try again later."}`, http.StatusTooManyRequests)
			return
		}

		recent = append(recent, now)
		rl.ips[ip] = recent

		next(w, r)
	}
}
