package repository

import (
	"context"
	"fmt"
)

func (r *AnalyticsRepositoryImpl) HealthCheck(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

func (r *AnalyticsRepositoryImpl) Ping(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	return r.db.Ping(ctx)
}
