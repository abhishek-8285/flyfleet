package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/governance/service"
	"go.uber.org/zap"
)

type GovernanceExpiryWorker struct {
	svc    service.GovernanceService
	logger *zap.Logger
}

func NewGovernanceExpiryWorker(svc service.GovernanceService, logger *zap.Logger) *GovernanceExpiryWorker {
	return &GovernanceExpiryWorker{
		svc:    svc,
		logger: logger,
	}
}

func (w *GovernanceExpiryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	w.logger.Info("Starting Governance Expiry Worker")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping Governance Expiry Worker")
			return
		case <-ticker.C:
			w.runExpiry(ctx)
		}
	}
}

func (w *GovernanceExpiryWorker) runExpiry(ctx context.Context) {
	w.logger.Debug("Running governance request expiry check")
	count, err := w.svc.ExpireRequests(ctx)
	if err != nil {
		w.logger.Error("failed to expire requests in worker", zap.Error(err))
		return
	}
	if count > 0 {
		w.logger.Info("expired governance requests", zap.Int("count", count))
	}
}
