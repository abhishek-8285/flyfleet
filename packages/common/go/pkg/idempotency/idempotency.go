package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IdempotencyManager handles request deduplication
type IdempotencyManager struct {
	redisClient *redis.Client
	prefix      string
}

// NewIdempotencyManager creates a new manager
func NewIdempotencyManager(redisClient *redis.Client, prefix string) *IdempotencyManager {
	return &IdempotencyManager{
		redisClient: redisClient,
		prefix:      prefix,
	}
}

// CheckAndSet checks if a key exists; if not, it sets it with a TTL.
// Returns true if the request is unique (not seen before).
func (im *IdempotencyManager) CheckAndSet(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	fullKey := fmt.Sprintf("%s:%s", im.prefix, key)
	
	// NX: Set if Not Exists
	success, err := im.redisClient.SetNX(ctx, fullKey, "processing", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("idempotency check failed: %w", err)
	}
	
	return success, nil
}

// Complete marks the operation as finished and updates the value
func (im *IdempotencyManager) Complete(ctx context.Context, key string, result []byte, ttl time.Duration) error {
	fullKey := fmt.Sprintf("%s:%s", im.prefix, key)
	return im.redisClient.Set(ctx, fullKey, result, ttl).Err()
}

// GenerateKey creates a unique key based on the request content
func (im *IdempotencyManager) GenerateKey(prefix string, data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s:%s", prefix, hex.EncodeToString(hash[:]))
}
