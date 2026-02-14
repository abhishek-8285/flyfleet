package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func Connect(ctx context.Context, databaseURL string, logger *zap.Logger) (*pgxpool.Pool, error) {
config, err := pgxpool.ParseConfig(databaseURL)
if err != nil {
return nil, fmt.Errorf("failed to parse database config: %w", err)
}

config.MaxConns = 20
config.MinConns = 5
config.MaxConnLifetime = time.Hour
config.HealthCheckPeriod = time.Minute

pool, err := pgxpool.NewWithConfig(ctx, config)
if err != nil {
return nil, fmt.Errorf("failed to connect to database: %w", err)
}

if err := pool.Ping(ctx); err != nil {
pool.Close()
return nil, fmt.Errorf("failed to ping database: %w", err)
}

logger.Info("Successfully connected to database")
return pool, nil
}

func Close(pool *pgxpool.Pool) {
if pool != nil {
pool.Close()
}
}
