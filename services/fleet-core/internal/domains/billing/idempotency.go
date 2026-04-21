package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IdempotencyHelper provides utilities for preventing duplicate transactions
type IdempotencyHelper struct {
	redisClient *redis.Client
}

// NewIdempotencyHelper creates a new idempotency helper
func NewIdempotencyHelper(redisClient *redis.Client) *IdempotencyHelper {
	return &IdempotencyHelper{
		redisClient: redisClient,
	}
}

// CheckAndLock attempts to acquire a lock for a specific transaction.
// Returns true if the lock was acquired (transaction is new),
// false if the transaction already exists (duplicate),
// and an error if the Redis operation failed.
func (i *IdempotencyHelper) CheckAndLock(ctx context.Context, tenantID, requestID string, ttl time.Duration) (bool, error) {
	if tenantID == "" || requestID == "" {
		return false, fmt.Errorf("tenantID and requestID are required for idempotency locking")
	}

	key := fmt.Sprintf("idempotency:%s:%s", tenantID, requestID)

	// SetNX returns true if the key was set (didn't exist before)
	acquired, err := i.redisClient.SetNX(ctx, key, "locked", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check/set idempotency lock: %w", err)
	}

	return acquired, nil
}
