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
			rl.mu.Unlock()
			http.Error(w, `{"error": "Too many requests. Please try again later."}`, http.StatusTooManyRequests)
			return
		}

		recent = append(recent, now)
		rl.ips[ip] = recent
		rl.mu.Unlock()

		next(w, r)
	}
}

// Enhanced rate limiter for authentication endpoints with different limits
type AuthRateLimiter struct {
	attempts map[string][]time.Time  // Track login attempts by IP
	emails   map[string][]time.Time  // Track login attempts by email
	mu       sync.Mutex
	loginLimit int
	window   time.Duration
	logger   *zap.Logger
}

func NewAuthRateLimiter(loginLimit int, window time.Duration, logger *zap.Logger) *AuthRateLimiter {
	return &AuthRateLimiter{
		attempts:   make(map[string][]time.Time),
		emails:     make(map[string][]time.Time),
		loginLimit: loginLimit,
		window:     window,
		logger:     logger,
	}
}

func (arl *AuthRateLimiter) LimitLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		
		// Parse request body to get email if available
		var email string
		if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
			// Note: In a real implementation you'd want to parse the body to extract email for rate limiting,
			// but we're not parsing it here to avoid consuming the request body
		}

		arl.mu.Lock()

		now := time.Now()
		expiry := now.Add(-arl.window)

		// Clean old requests for IP
		var ipRecent []time.Time
		for _, t := range arl.attempts[ip] {
			if t.After(expiry) {
				ipRecent = append(ipRecent, t)
			}
		}

		// Check if IP has exceeded limit
		if len(ipRecent) >= arl.loginLimit {
			arl.logger.Warn("Login rate limit exceeded by IP", zap.String("ip", ip))
			arl.mu.Unlock()
			http.Error(w, `{"error": "Too many login attempts. Please try again later."}`, http.StatusTooManyRequests)
			return
		}

		// If email is known, check email-based limit too
		if email != "" {
			var emailRecent []time.Time
			for _, t := range arl.emails[email] {
				if t.After(expiry) {
					emailRecent = append(emailRecent, t)
				}
			}

			if len(emailRecent) >= arl.loginLimit {
				arl.logger.Warn("Login rate limit exceeded for email", zap.String("email", email))
				arl.mu.Unlock()
				http.Error(w, `{"error": "Too many login attempts for this account. Please try again later."}`, http.StatusTooManyRequests)
				return
			}

			emailRecent = append(emailRecent, now)
			arl.emails[email] = emailRecent
		}

		ipRecent = append(ipRecent, now)
		arl.attempts[ip] = ipRecent
		arl.mu.Unlock()

		next(w, r)
	}
}
