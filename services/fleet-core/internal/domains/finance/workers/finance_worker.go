package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"go.uber.org/zap"
)

type FinanceWorker struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewFinanceWorker(repo *repository.FinanceRepository, logger *zap.Logger) *FinanceWorker {
	return &FinanceWorker{
		repo:   repo,
		logger: logger,
	}
}

func (w *FinanceWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Daily maintenance
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runDailyTasks(ctx)
		}
	}
}

func (w *FinanceWorker) runDailyTasks(ctx context.Context) {
	w.logger.Info("Running daily finance background tasks")
	// Task 1: Check for expired budgets and notify (Stub)
	// Task 2: Data retention cleanup (Phase 18)
}
