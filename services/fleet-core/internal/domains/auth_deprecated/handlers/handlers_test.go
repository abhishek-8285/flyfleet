package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRateLimit_BlocksAfterMaxAttempts(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create a redis client for miniredis
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	logger, _ := zap.NewDevelopment()
	limiter := middleware.NewRedisRateLimiter(rdb, logger)
	ip := "192.168.1.100"
	ctx := context.Background()

	// 5 allowed attempts
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "login:"+ip, 5, time.Minute)
		require.NoError(t, err)
		assert.True(t, allowed, "Attempt %d should be allowed", i+1)
	}

	// 6th attempt blocked
	allowed, err := limiter.Allow(ctx, "login:"+ip, 5, time.Minute)
	require.NoError(t, err)
	assert.False(t, allowed, "6th attempt should be blocked")
}
