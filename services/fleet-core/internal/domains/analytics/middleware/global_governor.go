package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/cache"
	"go.uber.org/zap"
)

// GlobalGovernor enforces a cluster-wide request limit for analytics (Phase 11)
type GlobalGovernor struct {
	cache  *cache.AnalyticsCache
	logger *zap.Logger
	limit  int // Global Cluster Limit
}

func NewGlobalGovernor(c *cache.AnalyticsCache, l *zap.Logger, limit int) *GlobalGovernor {
	if limit <= 0 {
		limit = 1000
	} // Default high limit
	return &GlobalGovernor{cache: c, logger: l, limit: limit}
}

// GlobalLimit protects the entire cluster from saturation
func (gg *GlobalGovernor) GlobalLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Use a global key with 1s precision
		key := fmt.Sprintf("global_governor:%d", time.Now().Unix())

		val, err := gg.cache.Increment(ctx, key, 1*time.Second)
		if err != nil {
			gg.logger.Error("Global governor cache error", zap.Error(err))
			next.ServeHTTP(w, r)
			return
		}

		if int(val) > gg.limit {
			gg.logger.Error("GLOBAL CLUSTER LIMIT EXCEEDED - THROTTLING ALL ANALYTICS",
				zap.Int("limit", gg.limit),
				zap.Int64("current", val))

			w.Header().Set("Retry-After", "5")
			http.Error(w, "Cluster-wide analytics capacity exceeded. Please retry later.", http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r)
	})
}
