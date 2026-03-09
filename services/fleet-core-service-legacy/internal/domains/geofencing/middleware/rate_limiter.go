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

type GeofenceRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rejected *prometheus.CounterVec
	logger   *zap.Logger
}

var geofenceRateLimitedCounter = func() *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "flyfleet_geofence_rate_limited_total",
		Help: "Total geofencing API requests rejected by rate limiter",
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

func NewGeofenceRateLimiter(logger *zap.Logger) *GeofenceRateLimiter {
	rl := &GeofenceRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rejected: geofenceRateLimitedCounter,
		logger:   logger,
	}

	rl.limiters["POST:/api/v1/geofencing"] = rate.NewLimiter(rate.Every(time.Second/10), 10)
	rl.limiters["PUT:/api/v1/geofencing/"] = rate.NewLimiter(rate.Every(time.Second/10), 10)
	rl.limiters["DELETE:/api/v1/geofencing/"] = rate.NewLimiter(rate.Every(time.Second/5), 5)
	rl.limiters["GET:/api/v1/geofencing"] = rate.NewLimiter(rate.Every(time.Second/50), 50)
	rl.limiters["default"] = rate.NewLimiter(rate.Every(time.Second/100), 100)

	return rl
}

func (rl *GeofenceRateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rl.getLimiter(r.Method, r.URL.Path)
		if !limiter.Allow() {
			rl.rejected.WithLabelValues(r.URL.Path, r.Method).Inc()
			rl.logger.Warn("Geofence rate limit exceeded", zap.String("method", r.Method), zap.String("path", r.URL.Path))
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *GeofenceRateLimiter) getLimiter(method, path string) *rate.Limiter {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if method == http.MethodPost && path == "/api/v1/geofencing" {
		return rl.limiters["POST:/api/v1/geofencing"]
	}
	if method == http.MethodGet && path == "/api/v1/geofencing" {
		return rl.limiters["GET:/api/v1/geofencing"]
	}
	if method == http.MethodPut && strings.HasPrefix(path, "/api/v1/geofencing/") {
		return rl.limiters["PUT:/api/v1/geofencing/"]
	}
	if method == http.MethodDelete && strings.HasPrefix(path, "/api/v1/geofencing/") {
		return rl.limiters["DELETE:/api/v1/geofencing/"]
	}
	return rl.limiters["default"]
}
