package workers

import (
	"context"
	"encoding/json"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/service"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type DeletionWorker struct {
	svc    service.DataDeletionService
	nats   *nats.Conn
	logger *zap.Logger
}

func NewDeletionWorker(svc service.DataDeletionService, nats *nats.Conn, logger *zap.Logger) *DeletionWorker {
	return &DeletionWorker{
		svc:    svc,
		nats:   nats,
		logger: logger,
	}
}

func (w *DeletionWorker) Start(ctx context.Context) {
	if w.nats == nil {
		w.logger.Warn("NATS not connected, DeletionWorker unable to receive deletion events")
		return
	}

	sub, err := w.nats.Subscribe("governance.request.approved", func(m *nats.Msg) {
		var req models.GovernanceRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			return
		}

		if req.RequestType == models.RequestTypeDeletion {
			w.logger.Info("received deletion request via NATS", zap.String("id", req.ID.String()))
			err := w.svc.ProcessDeletionRequest(ctx, req.ID)
			if err != nil {
				w.logger.Error("failed to process deletion", zap.Error(err))
			}
		}
	})
	if err != nil {
		w.logger.Error("failed to subscribe to deletion events", zap.Error(err))
	} else {
		defer sub.Unsubscribe()
	}

	<-ctx.Done()
}
