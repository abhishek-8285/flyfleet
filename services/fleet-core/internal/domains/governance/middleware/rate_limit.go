package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// DSARRateLimit prevents abuse of DSAR requests by limiting users to 5 requests per hour.
func DSARRateLimit(rdb *redis.Client, nc *nats.Conn, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := r.Header.Get("X-User-ID")
			orgIDStr := r.Header.Get("X-Organization-ID")
			if userIDStr == "" || rdb == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("governance_rate_limit:%s", userIDStr)
			ctx := r.Context()

			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				logger.Error("failed to increment rate limit counter", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rdb.Expire(ctx, key, time.Hour)
			}

			if count > 10 && nc != nil {
				// Publish Security Alert for DSAR Abuse
				payload, _ := json.Marshal(map[string]interface{}{
					"alert_type":      "DSAR_ABUSE_DETECTED",
					"tenant_id": orgIDStr,
					"user_id":         userIDStr,
					"timestamp":       time.Now().Unix(),
				})
				_ = nc.Publish("governance.security.alert", payload)
			}

			if count > 5 {
				logger.Warn("DSAR rate limit exceeded", zap.String("userID", userIDStr))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error": "Too many requests. Limit is 5 per hour."}`)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
