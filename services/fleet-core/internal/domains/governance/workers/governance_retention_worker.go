package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/governance/service"
	"go.uber.org/zap"
)

type GovernanceRetentionWorker struct {
	svc    service.GovernanceService
	logger *zap.Logger
}

func NewGovernanceRetentionWorker(svc service.GovernanceService, logger *zap.Logger) *GovernanceRetentionWorker {
	return &GovernanceRetentionWorker{
		svc:    svc,
		logger: logger,
	}
}

func (w *GovernanceRetentionWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	w.logger.Info("Starting Governance Retention Worker (Daily schedule)")

	// Run once on startup
	w.runRetention(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping Governance Retention Worker")
			return
		case <-ticker.C:
			w.runRetention(ctx)
		}
	}
}

func (w *GovernanceRetentionWorker) runRetention(ctx context.Context) {
	w.logger.Info("Running daily governance retention enforcement")
	if err := w.svc.EnforceRetention(ctx); err != nil {
		w.logger.Error("failed to enforce retention in worker", zap.Error(err))
	}
}
