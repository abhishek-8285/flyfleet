package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/service"
	"go.uber.org/zap"
)

type EWayBillSyncWorker struct {
	ewayBillService service.EWayBillIntegrationService
	logger          *zap.Logger
}

func NewEWayBillSyncWorker(es service.EWayBillIntegrationService, logger *zap.Logger) *EWayBillSyncWorker {
	return &EWayBillSyncWorker{
		ewayBillService: es,
		logger:          logger,
	}
}

func (w *EWayBillSyncWorker) Start(ctx context.Context) {
	w.logger.Info("EWayBillSyncWorker started")
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Initial sync
	w.Process(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("EWayBillSyncWorker stopped")
			return
		case <-ticker.C:
			w.Process(ctx)
		}
	}
}

func (w *EWayBillSyncWorker) Process(ctx context.Context) {
	w.logger.Info("EWayBillSyncWorker: starting periodic sync")
	if err := w.ewayBillService.SyncEWayBills(ctx); err != nil {
		w.logger.Error("EWayBillSyncWorker: sync failed", zap.Error(err))
	} else {
		w.logger.Info("EWayBillSyncWorker: sync completed successfully")
	}
}
