package service

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestProviderHealthService(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	// Skip if Redis is not available
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		t.Skip("Redis not available - skipping provider health service test")
		return
	}

	rdb.FlushDB(ctx)

	logger, _ := zap.NewDevelopment()
	service := NewProviderHealthService(rdb, logger)

	provider := "MSG91"

	// 1. Initially healthy
	if !service.IsHealthy(ctx, provider) {
		t.Errorf("Provider %s should be healthy initially", provider)
	}

	// 2. Record 11 failures out of 20 (threshold is > 50% failures over last 20)
	for i := 0; i < 11; i++ {
		service.RecordFailure(ctx, provider)
	}
	for i := 0; i < 9; i++ {
		service.RecordSuccess(ctx, provider)
	}

	// 3. Should be unhealthy
	if service.IsHealthy(ctx, provider) {
		t.Errorf("Provider %s should be unhealthy after 11 failures in 20 attempts", provider)
	}

	// 4. Record more successes to recover
	for i := 0; i < 15; i++ {
		service.RecordSuccess(ctx, provider)
	}

	// 5. Should be healthy again
	if !service.IsHealthy(ctx, provider) {
		t.Errorf("Provider %s should be healthy after recovery", provider)
	}
}
