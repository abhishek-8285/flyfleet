package warmup

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Warmer handles pre-warming of connections and caches
type Warmer struct {
	db          *sql.DB
	redisClient *redis.Client
}

// NewWarmer creates a new warmer
func NewWarmer(db *sql.DB, redisClient *redis.Client) *Warmer {
	return &Warmer{db: db, redisClient: redisClient}
}

// WarmDB pings the database to establish connections
func (w *Warmer) WarmDB(ctx context.Context, count int) error {
	for i := 0; i < count; i++ {
		if err := w.db.PingContext(ctx); err != nil {
			return fmt.Errorf("warmup: DB ping failed: %w", err)
		}
	}
	return nil
}

// WarmRedis pings redis to establish connections
func (w *Warmer) WarmRedis(ctx context.Context) error {
	if _, err := w.redisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("warmup: Redis ping failed: %w", err)
	}
	return nil
}

// WarmAll warms both DB and Redis
func (w *Warmer) WarmAll(ctx context.Context) error {
	if w.db != nil {
		if err := w.WarmDB(ctx, 5); err != nil {
			return err
		}
	}
	if w.redisClient != nil {
		if err := w.WarmRedis(ctx); err != nil {
			return err
		}
	}
	return nil
}
