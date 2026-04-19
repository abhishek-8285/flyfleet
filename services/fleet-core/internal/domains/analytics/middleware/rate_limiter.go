package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics"
	"github.com/flyfleet/fleet-core/internal/domains/analytics/cache"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const OrgIDKey contextKey = "tenant_id"

// RateLimiter enforces a request quota per organization using Redis
type RateLimiter struct {
	cache   *cache.AnalyticsCache
	service analytics.AnalyticsService
	logger  *zap.Logger
}

func NewRateLimiter(c *cache.AnalyticsCache, s analytics.AnalyticsService, l *zap.Logger) *RateLimiter {
	return &RateLimiter{cache: c, service: s, logger: l}
}

// OrgRateLimit restricts organizations using Phase 10 Adaptive Rate Limiting
func (rl *RateLimiter) OrgRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orgID, ok := ctx.Value(OrgIDKey).(uuid.UUID)
		if !ok || orgID == uuid.Nil {
			next.ServeHTTP(w, r)
			return
		}

		limit, err := rl.service.GetDynamicRateLimit(ctx, orgID)
		if err != nil {
			limit = 10 // Phase 7 fallback
		}

		key := fmt.Sprintf("rate_limit:%s:%d", orgID.String(), time.Now().Unix())

		val, err := rl.cache.Increment(ctx, key, 1*time.Second)
		if err != nil {
			rl.logger.Error("Rate limiter cache error", zap.Error(err))
			next.ServeHTTP(w, r)
			return
		}

		if int(val) > limit {
			rl.logger.Warn("Organization rate limit exceeded",
				zap.String("tenant_id", orgID.String()),
				zap.Int("limit", limit))
			w.Header().Set("Retry-After", "1")
			http.Error(w, fmt.Sprintf("Organization rate limit exceeded (%d req/sec)", limit), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
