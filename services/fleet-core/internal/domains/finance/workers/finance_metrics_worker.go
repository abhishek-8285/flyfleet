package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"github.com/flyfleet/fleet-core/internal/domains/finance/service"
	"go.uber.org/zap"
)

// FinanceMetricsWorker periodically refreshes materialized views and updates Prometheus gauges.
type FinanceMetricsWorker struct {
	repo    *repository.FinanceRepository
	metrics *service.FinanceMetrics
	logger  *zap.Logger
}

func NewFinanceMetricsWorker(repo *repository.FinanceRepository, metrics *service.FinanceMetrics, logger *zap.Logger) *FinanceMetricsWorker {
	return &FinanceMetricsWorker{
		repo:    repo,
		metrics: metrics,
		logger:  logger,
	}
}

func (w *FinanceMetricsWorker) Start(ctx context.Context) {
	// Refresh every 15 minutes
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Initial run
	w.refresh(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.refresh(ctx)
		}
	}
}

func (w *FinanceMetricsWorker) refresh(ctx context.Context) {
	w.logger.Info("FinanceMetricsWorker: refreshing dashboard materialized views")

	startTime := time.Now()
	if err := w.repo.RefreshFinanceDashboardSummary(ctx); err != nil {
		w.logger.Error("FinanceMetricsWorker: failed to refresh summary view", zap.Error(err))
	} else {
		w.logger.Info("FinanceMetricsWorker: dashboard summary view refreshed", zap.Duration("took", time.Since(startTime)))
	}

	// Update Prometheus metrics from the aggregation tables if needed.
	// In production we might pull the latest profit margins here.
}
