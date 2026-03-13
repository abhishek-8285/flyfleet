package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ProviderHealthService struct {
	redis  *redis.Client
	logger *zap.Logger
}

func NewProviderHealthService(rdb *redis.Client, logger *zap.Logger) *ProviderHealthService {
	return &ProviderHealthService{
		redis:  rdb,
		logger: logger,
	}
}

func (s *ProviderHealthService) RecordSuccess(ctx context.Context, providerName string) {
	s.recordAttempt(ctx, providerName, true)
}

func (s *ProviderHealthService) RecordFailure(ctx context.Context, providerName string) {
	s.recordAttempt(ctx, providerName, false)
}

func (s *ProviderHealthService) recordAttempt(ctx context.Context, providerName string, success bool) {
	key := fmt.Sprintf("emergency:provider_health:%s", providerName)
	pipe := s.redis.Pipeline()
	
	if success {
		pipe.Incr(ctx, key+":success")
	} else {
		pipe.Incr(ctx, key+":failure")
	}
	
	// Keep a list of last 20 attempts for rolling failure rate
	listKey := key + ":history"
	outcome := "1"
	if !success {
		outcome = "0"
	}
	pipe.LPush(ctx, listKey, outcome)
	pipe.LTrim(ctx, listKey, 0, 19)
	pipe.Set(ctx, key+":last_attempt", time.Now().Format(time.RFC3339), 0)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Error("Failed to record provider attempt in Redis", zap.String("provider", providerName), zap.Error(err))
	}
}

func (s *ProviderHealthService) IsHealthy(ctx context.Context, providerName string) bool {
	key := fmt.Sprintf("emergency:provider_health:%s:history", providerName)
	history, err := s.redis.LRange(ctx, key, 0, -1).Result()
	if err != nil || len(history) < 10 {
		return true // Opt-in to try if not enough data
	}

	failures := 0
	for _, h := range history {
		if h == "0" {
			failures++
		}
	}

	failureRate := float64(failures) / float64(len(history))
	if failureRate >= 0.5 {
		s.logger.Warn("Provider marked UNHEALTHY due to high failure rate", 
			zap.String("provider", providerName), 
			zap.Float64("failure_rate", failureRate))
		return false
	}

	return true
}
