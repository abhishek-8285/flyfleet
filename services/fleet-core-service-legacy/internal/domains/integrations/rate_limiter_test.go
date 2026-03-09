//go:build integration
// +build integration

package integrations

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRateLimiter_CapacityAndRefill(t *testing.T) {
	registry := NewRateLimiterRegistry()

	integrationID := uuid.New().String()

	// Configure exactly 5 tokens, refill 5 per second.
	cfg := &RateLimitConfig{
		Capacity:   5,
		RefillRate: 5.0,
	}
	limiter := registry.For(integrationID, cfg)

	// Consume 5 tokens immediately. Should all succeed.
	for i := 0; i < 5; i++ {
		if err := limiter.Allow(1.0); err != nil {
			t.Errorf("expected token %d to be allowed, got error: %v", i+1, err)
		}
	}

	// 6th token should be rejected (bucket empty)
	if err := limiter.Allow(1.0); err == nil {
		t.Errorf("expected 6th token to be rejected by rate limiter")
	}

	// Wait 1.1 seconds for refill
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	if err := limiter.Allow(1.0); err != nil {
		t.Errorf("expected token to be allowed after bucket refill: %v", err)
	}
}
