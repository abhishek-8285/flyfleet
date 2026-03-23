package middleware

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// FinanceRateLimiter provides per-endpoint rate limiting for finance APIs.
type FinanceRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rejected *prometheus.CounterVec
	logger   *zap.Logger
}

var financeRateLimitedCounter = func() *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "flyfleet_finance_rate_limited_total",
		Help: "Total finance API requests rejected by rate limiter",
	}, []string{"path", "method"})

	if err := prometheus.Register(counter); err != nil {
		var alreadyRegisteredErr prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredErr) {
			if existing, ok := alreadyRegisteredErr.ExistingCollector.(*prometheus.CounterVec); ok {
				return existing
			}
		}
	}

	return counter
}()

// NewFinanceRateLimiter creates a rate limiter with per-endpoint limits.
func NewFinanceRateLimiter(logger *zap.Logger) *FinanceRateLimiter {
	rl := &FinanceRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rejected: financeRateLimitedCounter,
		logger:   logger,
	}

	// Pre-configure known route limits
	rl.limiters["/api/v1/finance/fuel/ingest"] = rate.NewLimiter(rate.Every(time.Second/50), 50)
	rl.limiters["/api/v1/finance/expenses/reports"] = rate.NewLimiter(rate.Every(time.Second/20), 20)
	rl.limiters["/api/v1/gst"] = rate.NewLimiter(rate.Every(time.Second/10), 10)
	rl.limiters["/api/v1/vendors/payments"] = rate.NewLimiter(rate.Every(time.Second/20), 20)
	rl.limiters["default"] = rate.NewLimiter(rate.Every(time.Second/100), 100)

	return rl
}

// Handler returns an HTTP middleware that enforces rate limits per route prefix.
func (rl *FinanceRateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rl.getLimiter(r.URL.Path)

		if !limiter.Allow() {
			rl.rejected.WithLabelValues(r.URL.Path, r.Method).Inc()
			rl.logger.Warn("Rate limit exceeded",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *FinanceRateLimiter) getLimiter(path string) *rate.Limiter {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Match longest prefix
	for prefix, limiter := range rl.limiters {
		if prefix != "default" && strings.HasPrefix(path, prefix) {
			return limiter
		}
	}
	return rl.limiters["default"]
}
