package reports

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	reportCacheTTL     = 1 * time.Hour
	reportCacheVersion = "v1" // bump to invalidate all cached keys on schema change
)

// ReportCache wraps Redis to cache identical report requests by a stable hash key.
type ReportCache struct {
	Client *redis.Client
	logger *zap.Logger
}

func NewReportCache(client *redis.Client, logger *zap.Logger) *ReportCache {
	return &ReportCache{
		Client: client,
		logger: logger,
	}
}

// CacheKey builds a deterministic, collision-resistant key.
// Fixes Issue #2:
//   - Normalizes parameter map to sorted JSON (ordering-independent)
//   - Includes format, tenantID, reportType, and version in key
//   - All components are explicitly concatenated in a fixed order
func (c *ReportCache) CacheKey(tenantID, reportType, format string, params map[string]interface{}) string {
	// Normalize params: sort keys before marshalling
	normalized := normalizeParams(params)
	raw, _ := json.Marshal(normalized)
	input := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		reportCacheVersion, tenantID, reportType, format, string(raw), reportType)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("reports:cache:%x", hash)
}

// normalizeParams returns a deterministic JSON-serializable representation,
// sorting all keys alphabetically to eliminate ordering differences.
func normalizeParams(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return nil
	}
	result := make(map[string]interface{}, len(params))
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result[k] = params[k]
	}
	return result
}

// Get returns the cached report ID for an identical request, or ("", false).
func (c *ReportCache) Get(ctx context.Context, key string) (string, bool) {
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// Set caches the report ID under the given key with a 1h TTL.
func (c *ReportCache) Set(ctx context.Context, key, reportID string) {
	err := c.Client.Set(ctx, key, reportID, reportCacheTTL).Err()
	if err != nil {
		c.logger.Warn("Failed to cache report", zap.String("key", key), zap.Error(err))
	}
}

// Invalidate removes a cached entry.
func (c *ReportCache) Invalidate(ctx context.Context, key string) {
	c.Client.Del(ctx, key)
}
