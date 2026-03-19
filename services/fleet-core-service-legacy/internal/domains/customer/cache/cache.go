package cache

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Cache struct {
	logger *zap.Logger
}

func NewCache(logger *zap.Logger) *Cache {
	return &Cache{
		logger: logger,
	}
}

// SetStub stubs setting an item in redis cache
func (c *Cache) SetStub(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	c.logger.Debug("Setting cache item", zap.String("key", key))
	return nil
}

// GetStub stubs getting an item in redis cache
func (c *Cache) GetStub(ctx context.Context, key string) (interface{}, error) {
	c.logger.Debug("Getting cache item", zap.String("key", key))
	return nil, nil
}
