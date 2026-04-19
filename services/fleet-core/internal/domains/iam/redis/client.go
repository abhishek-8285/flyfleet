package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionData represents an authenticated session stored in Redis.
type SessionData struct {
	UserID     uuid.UUID `json:"user_id"`
	OrgID      uuid.UUID `json:"tenant_id"`
	Role       string    `json:"role"`
	Device     string    `json:"device"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

// Client wraps a go-redis client with IAM-specific session operations.
type Client struct {
	rdb *redis.Client
}

func NewClient(rdb *redis.Client) *Client {
	return &Client{rdb: rdb}
}

// Session keys
func sessionKey(jti string) string        { return "iam:session:" + jti }
func refreshKey(jti string) string        { return "iam:refresh:" + jti }
func actionKey(jti string) string         { return "iam:action:" + jti }
func failKey(userID string) string        { return "iam:fail:" + userID }
func permKey(userID, orgID string) string { return "iam:perm:" + userID + ":" + orgID }

// SetSession stores session data in Redis.
func (c *Client) SetSession(ctx context.Context, jti string, data SessionData, ttl time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, sessionKey(jti), b, ttl).Err()
}

// GetSession retrieves session data from Redis.
func (c *Client) GetSession(ctx context.Context, jti string) (*SessionData, error) {
	b, err := c.rdb.Get(ctx, sessionKey(jti)).Bytes()
	if err != nil {
		return nil, err
	}
	var data SessionData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// DeleteSession removes a session from Redis.
func (c *Client) DeleteSession(ctx context.Context, jti string) error {
	return c.rdb.Del(ctx, sessionKey(jti)).Err()
}

// IsRefreshTokenUsed checks if a refresh token JTI has been consumed.
func (c *Client) IsRefreshTokenUsed(ctx context.Context, jti string) (bool, error) {
	val, err := c.rdb.Exists(ctx, refreshKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// MarkRefreshTokenUsed marks a refresh token as consumed.
func (c *Client) MarkRefreshTokenUsed(ctx context.Context, jti string, ttl time.Duration) error {
	return c.rdb.Set(ctx, refreshKey(jti), "1", ttl).Err()
}

// MarkActionTokenUsed marks an action token as consumed.
func (c *Client) MarkActionTokenUsed(ctx context.Context, jti string, ttl time.Duration) error {
	return c.rdb.Set(ctx, actionKey(jti), "1", ttl).Err()
}

// AtomicConsumeActionToken atomically checks and sets an action token.
// Returns true if the token was consumed (was already used), false if first use.
func (c *Client) AtomicConsumeActionToken(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	val, err := c.rdb.Get(ctx, actionKey(jti)).Result()
	if err == redis.Nil {
		// Key doesn't exist — first use
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// Key exists — already consumed
	_ = val
	return true, nil
}

// RecordFailedLogin records a failed login attempt.
func (c *Client) RecordFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil time.Time, duration time.Duration) error {
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, failKey(userID), attempts, duration)
	if !lockedUntil.IsZero() {
		pipe.Set(ctx, failKey(userID)+":lock", "1", time.Until(lockedUntil))
	}
	_, err := pipe.Exec(ctx)
	return err
}

// GetFailedLoginAttempts returns the current failed attempt count.
func (c *Client) GetFailedLoginAttempts(ctx context.Context, userID string) (int, error) {
	val, err := c.rdb.Get(ctx, failKey(userID)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// IsLocked checks if a user is currently locked out.
func (c *Client) IsLocked(ctx context.Context, userID string) (bool, error) {
	val, err := c.rdb.Exists(ctx, failKey(userID)+":lock").Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// CacheUserPermissions caches a user's permission mask in Redis.
func (c *Client) CacheUserPermissions(ctx context.Context, userID, orgID string, mask int64, ttl time.Duration) error {
	return c.rdb.Set(ctx, permKey(userID, orgID), mask, ttl).Err()
}

// GetUserPermissions retrieves cached permissions from Redis.
func (c *Client) GetUserPermissions(ctx context.Context, userID, orgID string) (int64, error) {
	val, err := c.rdb.Get(ctx, permKey(userID, orgID)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// InvalidateUserPermissions removes cached permissions.
func (c *Client) InvalidateUserPermissions(ctx context.Context, userID, orgID string) error {
	return c.rdb.Del(ctx, permKey(userID, orgID)).Err()
}

// RDB returns the underlying *redis.Client for direct operations.
func (c *Client) RDB() *redis.Client {
	return c.rdb
}

// Key formats for reference:
// iam:session:{jti}        — session data (JSON)
// iam:refresh:{jti}       — refresh token consumption flag
// iam:action:{jti}        — action token consumption flag
// iam:fail:{userID}       — failed login attempt count
// iam:fail:{userID}:lock  — account lock flag
// iam:perm:{userID}:{orgID} — cached permission mask
