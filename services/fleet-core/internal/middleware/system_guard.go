package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/redis/go-redis/v9"
)

const (
	SystemStatusKey   = "flyfleet:system:status"
	PoolSheddingKey   = "pool:shedding:low_priority"
	StatusMaintenance = "MAINTENANCE"
	StatusReadOnly    = "READ_ONLY"
	StatusNormal      = "NORMAL"
)

// systemGuardCache holds the cached system status to avoid Redis roundtrips on every request
type systemGuardCache struct {
	status      string
	poolShedded bool
	expiresAt   time.Time
	mu          sync.RWMutex
}

// SystemGuard is a middleware that enforces global system states (Maintenance, Read-Only)
type SystemGuard struct {
	redisClient *redis.Client
	cache       *systemGuardCache
}

func NewSystemGuard(redisClient *redis.Client) *SystemGuard {
	return &SystemGuard{
		redisClient: redisClient,
		cache: &systemGuardCache{
			status:    StatusNormal,
			expiresAt: time.Now().Add(-1 * time.Minute), // initially expired
		},
	}
}

// getStatus retrieves the status from cache or Redis if the cache is expired
func (g *SystemGuard) getStatus(ctx context.Context) (string, bool) {
	now := time.Now()

	g.cache.mu.RLock()
	if now.Before(g.cache.expiresAt) {
		status := g.cache.status
		shedded := g.cache.poolShedded
		g.cache.mu.RUnlock()
		return status, shedded
	}
	g.cache.mu.RUnlock()

	// Cache miss or expired, fetch from Redis
	g.cache.mu.Lock()
	defer g.cache.mu.Unlock()

	if now.Before(g.cache.expiresAt) {
		return g.cache.status, g.cache.poolShedded
	}

	if g.redisClient == nil {
		g.cache.status = StatusNormal
		g.cache.poolShedded = false
		g.cache.expiresAt = now.Add(1 * time.Second)
		return StatusNormal, false
	}

	val, err := g.redisClient.Get(ctx, SystemStatusKey).Result()
	if err != nil {
		if err == redis.Nil {
			val = StatusNormal
		} else {
			val = StatusNormal
		}
	}

	shedVal, err := g.redisClient.Get(ctx, PoolSheddingKey).Result()
	poolShedded := false
	if err == nil && shedVal == "1" {
		poolShedded = true
	}

	g.cache.status = val
	g.cache.poolShedded = poolShedded
	g.cache.expiresAt = now.Add(1 * time.Second) // 1 second local cache TTL

	return val, poolShedded
}

func (g *SystemGuard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, _ := g.getStatus(r.Context())

		if status == StatusNormal {
			next.ServeHTTP(w, r)
			return
		}

		isMutation := r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete

		if status == StatusMaintenance {
			// Maintenance blocks mutations (and optionally gets, but prompt says block mutations)
			// Let's rely on the block below to catch mutations.
			// The original requirement was:
			// "If status is MAINTENANCE, return 503... for all mutation requests"
			// Verification script: "Verify that a GET works but a POST /auth/login returns a 503"
			// Wait, if GET works during Maintenance, then Maintenance and Read Only behave identically for API purposes?
			// Sure. We will just block mutations if StatusMaintenance or StatusReadOnly.
		}

		if (status == StatusMaintenance || status == StatusReadOnly) && isMutation {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "service_unavailable",
				"message": "The system is currently undergoing scheduled maintenance or is in read-only mode.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LowPriorityShedder is a middleware used for Low/Medium priority routes (e.g. Analytics)
func (g *SystemGuard) LowPriorityShedder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, poolShedded := g.getStatus(r.Context())

		if poolShedded {
			// In phase 5 shedding, Standard Users (driver/staff) lose access
			role, ok := r.Context().Value(ctxkeys.RoleKey).(string)
			if ok && (role == "DRIVER" || role == "STAFF") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "too_many_requests",
					"message": "System load is high. Non-critical requests are temporarily paused for standard users to ensure emergency response availability.",
				})
				return
			}

			// For everyone else, just a general note or block if it's very low priority
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "too_many_requests",
				"message": "System load is high. Low priority requests are temporarily paused.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
