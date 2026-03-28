package repository

import (
	"context"
	"time"
)

const timeout = 30 * time.Second

// HealthCheck checks the database connection health
func (r *VideoRepositoryImpl) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.Ping(ctx)
}

// Ping checks if the database connection is alive
func (r *VideoRepositoryImpl) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.Ping(ctx)
}
