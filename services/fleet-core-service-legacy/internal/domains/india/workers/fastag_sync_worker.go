package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/service"
	"go.uber.org/zap"
)

type FastagSyncWorker struct {
	fastagService service.FastagIntegrationService
	logger        *zap.Logger
}

func NewFastagSyncWorker(fs service.FastagIntegrationService, logger *zap.Logger) *FastagSyncWorker {
	return &FastagSyncWorker{
		fastagService: fs,
		logger:        logger,
	}
}

func (w *FastagSyncWorker) Start(ctx context.Context) {
	w.logger.Info("FastagSyncWorker started")
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Initial sync
	w.Process(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("FastagSyncWorker stopped")
			return
		case <-ticker.C:
			w.Process(ctx)
		}
	}
}

func (w *FastagSyncWorker) Process(ctx context.Context) {
	w.logger.Info("FastagSyncWorker: starting periodic sync")
	if err := w.fastagService.SyncTransactions(ctx); err != nil {
		w.logger.Error("FastagSyncWorker: sync failed", zap.Error(err))
	} else {
		w.logger.Info("FastagSyncWorker: sync completed successfully")
	}
}
