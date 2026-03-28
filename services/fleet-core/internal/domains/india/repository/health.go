package repository

import (
	"context"
	"fmt"
)

func (r *IndiaIntegrationRepositoryImpl) HealthCheck(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Ping checks if the database connection is alive
func (r *IndiaIntegrationRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// CreateIntegrationLog creates a new integration audit log
