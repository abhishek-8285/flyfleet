package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/service"
	"go.uber.org/zap"
)

type GSTComplianceWorker struct {
	gstService service.GSTIntegrationService
	logger     *zap.Logger
}

func NewGSTComplianceWorker(gs service.GSTIntegrationService, logger *zap.Logger) *GSTComplianceWorker {
	return &GSTComplianceWorker{
		gstService: gs,
		logger:     logger,
	}
}

func (w *GSTComplianceWorker) Start(ctx context.Context) {
	w.logger.Info("GSTComplianceWorker started")
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Initial sync
	w.Process(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("GSTComplianceWorker stopped")
			return
		case <-ticker.C:
			w.Process(ctx)
		}
	}
}

func (w *GSTComplianceWorker) Process(ctx context.Context) {
	w.logger.Info("GSTComplianceWorker: starting periodic sync")
	if err := w.gstService.SyncCompliance(ctx); err != nil {
		w.logger.Error("GSTComplianceWorker: sync failed", zap.Error(err))
	} else {
		w.logger.Info("GSTComplianceWorker: sync completed successfully")
	}
}
