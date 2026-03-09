//go:build integration
// +build integration

package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimit_BlocksAfterMaxAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Assuming NewRedisRateLimiter is implemented in middleware package
	limiter := middleware.NewRedisRateLimiter(mr.Addr(), 5, time.Minute)
	ip := "192.168.1.100"
	ctx := context.Background()

	// 5 allowed attempts
	for i := 0; i < 5; i++ {
		allowed, err := limiter.CheckLimit(ctx, ip, "login")
		require.NoError(t, err)
		assert.True(t, allowed, "Attempt %d should be allowed", i+1)
	}

	// 6th attempt blocked
	allowed, err := limiter.CheckLimit(ctx, ip, "login")
	require.NoError(t, err)
	assert.False(t, allowed, "6th attempt should be blocked")
}
