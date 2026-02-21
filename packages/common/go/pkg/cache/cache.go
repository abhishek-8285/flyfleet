package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache represents a Redis-backed cache
type Cache struct {
	client *redis.Client
	prefix string
}

// NewCache creates a new cache instance
func NewCache(client *redis.Client, prefix string) *Cache {
	return &Cache{
		client: client,
		prefix: prefix,
	}
}

// Set stores a value in the cache with a TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	fullKey := fmt.Sprintf("%s:%s", c.prefix, key)
	return c.client.Set(ctx, fullKey, data, ttl).Err()
}

// Get retrieves a value from the cache
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := fmt.Sprintf("%s:%s", c.prefix, key)
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss: %s", key)
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a value from the cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("%s:%s", c.prefix, key)
	return c.client.Del(ctx, fullKey).Err()
}

// InvalidatePattern invalidates keys matching a pattern
func (c *Cache) InvalidatePattern(ctx context.Context, pattern string) error {
	fullPattern := fmt.Sprintf("%s:%s", c.prefix, pattern)
	iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
