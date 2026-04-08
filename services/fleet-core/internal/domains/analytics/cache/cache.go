package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const defaultTTL = 5 * time.Minute

// AnalyticsCache handles caching for analytics query results
type AnalyticsCache struct {
	client  *redis.Client
	metrics *analytics.Metrics
	logger  *zap.Logger
}

// NewAnalyticsCache creates a new instance of AnalyticsCache
func NewAnalyticsCache(client *redis.Client, metrics *analytics.Metrics, logger *zap.Logger) *AnalyticsCache {
	return &AnalyticsCache{
		client:  client,
		metrics: metrics,
		logger:  logger,
	}
}

// PublishInvalidation broadcasts a cache invalidation event to all service nodes
func (c *AnalyticsCache) PublishInvalidation(ctx context.Context, pattern string) error {
	return c.client.Publish(ctx, "analytics:cache:invalidation", pattern).Err()
}

// SubscribeToInvalidations listens for invalidation events and applies them locally
func (c *AnalyticsCache) SubscribeToInvalidations(ctx context.Context) {
	pubsub := c.client.Subscribe(ctx, "analytics:cache:invalidation")

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			c.logger.Info("Received distributed cache invalidation event", zap.String("pattern", msg.Payload))
			if err := c.InvalidateByPattern(ctx, msg.Payload); err != nil {
				c.logger.Error("Failed to apply distributed invalidation", zap.Error(err))
			}
		}
	}()
}

// GenerateKey builds a hierarchical cache key.
// resourceType: "vehicle", "driver", "fleet"
func (c *AnalyticsCache) GenerateKey(ctx context.Context, tenantID, resourceType, resourceID, queryType string, params map[string]interface{}) string {
	// Sort parameters for deterministic hashing
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	normalized := make(map[string]interface{})
	for _, k := range keys {
		normalized[k] = params[k]
	}

	paramsJSON, _ := json.Marshal(normalized)

	var keyPrefix string
	if resourceType == "fleet" {
		version := c.GetFleetVersion(ctx, tenantID)
		keyPrefix = fmt.Sprintf("analytics:%s:fleet:v%d", tenantID, version)
	} else {
		keyPrefix = fmt.Sprintf("analytics:%s:%s:%s", tenantID, resourceType, resourceID)
	}

	input := fmt.Sprintf("%s:%s:%s", keyPrefix, queryType, string(paramsJSON))
	hash := sha256.Sum256([]byte(input))

	return fmt.Sprintf("%s:%s:%x", keyPrefix, queryType, hash)
}

// GetFleetVersion retrieves the current version for fleet analytics from Redis
func (c *AnalyticsCache) GetFleetVersion(ctx context.Context, tenantID string) int64 {
	versionKey := fmt.Sprintf("analytics:%s:fleet:version", tenantID)
	val, err := c.client.Get(ctx, versionKey).Int64()
	if err != nil {
		return 1 // Default version
	}
	return val
}

// IncrementFleetVersion increments the fleet version to effectively invalidate all fleet cache entries
func (c *AnalyticsCache) IncrementFleetVersion(ctx context.Context, tenantID string) error {
	versionKey := fmt.Sprintf("analytics:%s:fleet:version", tenantID)
	return c.client.Incr(ctx, versionKey).Err()
}

// Get retrieves a cached result into the destination object
func (c *AnalyticsCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		if c.metrics != nil {
			c.metrics.CacheMissesTotal.WithLabelValues("unknown", "unknown").Inc()
		}
		return false, nil
	}
	if err != nil {
		c.logger.Debug("Cache get error", zap.String("key", key), zap.Error(err))
		return false, err
	}

	if c.metrics != nil {
		c.metrics.CacheHitsTotal.WithLabelValues("unknown", "unknown").Inc()
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.Warn("Failed to unmarshal cached data", zap.String("key", key), zap.Error(err))
		return false, err
	}

	return true, nil
}

// Set stores a result in the cache with a specified TTL
func (c *AnalyticsCache) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = defaultTTL
	}

	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := c.client.Set(ctx, key, string(val), ttl).Err(); err != nil {
		c.logger.Warn("Failed to set cache entry", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}

// SetRaw stores raw bytes in the cache with the given TTL
func (c *AnalyticsCache) SetRaw(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = defaultTTL
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("Failed to set raw cache key", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// GetRaw retrieves raw bytes from the cache
func (c *AnalyticsCache) GetRaw(ctx context.Context, key string) ([]byte, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			c.logger.Error("Failed to get raw cache key", zap.String("key", key), zap.Error(err))
		}
		return nil, err
	}
	return data, nil
}

// Invalidate removes a cache entry
func (c *AnalyticsCache) Invalidate(ctx context.Context, key string) error {
	if c.metrics != nil {
		c.metrics.CacheInvalidationsTotal.WithLabelValues("direct", "unknown").Inc()
	}
	return c.client.Del(ctx, key).Err()
}

// InvalidateByPattern removes cache entries matching a pattern using SCAN
func (c *AnalyticsCache) InvalidateByPattern(ctx context.Context, pattern string) error {
	if c.metrics != nil {
		c.metrics.CacheInvalidationsTotal.WithLabelValues("pattern", "unknown").Inc()
	}

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			c.logger.Error("Failed to delete key during pattern invalidation",
				zap.String("key", iter.Val()), zap.Error(err))
		}
	}
	return iter.Err()
}

// Debounce checks if an operation with the given key should be throttled.
// Returns true if the operation is debounced (should be skipped).
func (c *AnalyticsCache) Debounce(ctx context.Context, key string, ttl time.Duration) bool {
	debounceKey := fmt.Sprintf("debounce:%s", key)
	// SETNX returns true if the key was set
	set, err := c.client.SetNX(ctx, debounceKey, "1", ttl).Result()
	if err != nil {
		c.logger.Error("Debounce check error", zap.Error(err))
		return false
	}
	return !set // If it wasn't set, it means it already exists, so debounce it.
}

// Increment increments a key and sets a TTL if it's a new key
func (c *AnalyticsCache) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
