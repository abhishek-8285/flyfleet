package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const queryTimeout = 30 * time.Second

// AnalyticsRepositoryImpl implements the AnalyticsRepository interface
// This file contains only the struct definition, constructor, and helper methods.
// Implementation methods are split across multiple files for maintainability:
// - predictive_models.go
// - prediction_results.go
// - analytics_jobs.go
// - fleet_analytics.go
// - driver_behavior_analytics.go
// - predictive_maintenance.go
// - operations.go
// - health.go
type AnalyticsRepositoryImpl struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewAnalyticsRepository creates a new instance of AnalyticsRepositoryImpl
func NewAnalyticsRepository(db *pgxpool.Pool, logger *zap.Logger) *AnalyticsRepositoryImpl {
	return &AnalyticsRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *AnalyticsRepositoryImpl) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, queryTimeout)
}
