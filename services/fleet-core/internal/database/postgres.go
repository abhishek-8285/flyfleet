package database

import (
	"context"
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewPostgresPool(ctx context.Context, databaseURL string, logger *zap.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 100
	config.MinConns = 10
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// Optimize for prepared statement caching to reduce query parse/plan overhead
	// Use CacheStatement mode to ensure all queries are prepared and cached on the server
	if config.ConnConfig != nil {
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database pool initialized successfully")
	return pool, nil
}
