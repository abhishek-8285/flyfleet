package integrations

import (
	"fmt"
	"sync"
	"time"
)

// ─── Token Bucket Rate Limiter ────────────────────────────────────────────────
//
// Limits request rate per integration/partner.
// Default: capacity=100, refill=10 tokens/sec.
// Partner-specific overrides set via IntegrationConfig.RateLimitConfig.

type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per nanosecond
	lastRefill time.Time
}

func NewTokenBucket(capacity int, refillPerSec float64) *TokenBucket {
	return &TokenBucket{
		tokens:     float64(capacity),
		capacity:   float64(capacity),
		refillRate: refillPerSec / 1e9, // convert to per-nanosecond
		lastRefill: time.Now(),
	}
}

// Allow attempts to consume n tokens. Returns error if rate limit exceeded.
func (tb *TokenBucket) Allow(n float64) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := float64(now.Sub(tb.lastRefill))
	tb.tokens = min64(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now

	if tb.tokens < n {
		return fmt.Errorf("rate limit exceeded: available=%.1f requested=%.1f", tb.tokens, n)
	}
	tb.tokens -= n
	return nil
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ─── Rate Limiter Registry ────────────────────────────────────────────────────

type RateLimiterRegistry struct {
	mu       sync.RWMutex
	limiters map[string]*TokenBucket
}

func NewRateLimiterRegistry() *RateLimiterRegistry {
	return &RateLimiterRegistry{limiters: make(map[string]*TokenBucket)}
}

// For returns the rate limiter for an integration, creating one with defaults if absent.
func (r *RateLimiterRegistry) For(integrationID string, cfg *RateLimitConfig) *TokenBucket {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tb, ok := r.limiters[integrationID]; ok {
		return tb
	}
	capacity := 100
	refill := 10.0
	if cfg != nil {
		capacity = cfg.Capacity
		refill = cfg.RefillRate
	}
	tb := NewTokenBucket(capacity, refill)
	r.limiters[integrationID] = tb
	return tb
}
