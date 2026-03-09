package workers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/service"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type ExportProcessingWorker struct {
	svc    service.ExportService
	nats   *nats.Conn
	logger *zap.Logger
}

func NewExportProcessingWorker(svc service.ExportService, nats *nats.Conn, logger *zap.Logger) *ExportProcessingWorker {
	return &ExportProcessingWorker{
		svc:    svc,
		nats:   nats,
		logger: logger,
	}
}

func (w *ExportProcessingWorker) Start(ctx context.Context) {
	if w.nats == nil {
		w.logger.Warn("NATS not connected, ExportProcessingWorker running in fallback polling mode")
		go w.pollApprovedRequests(ctx)
		return
	}

	sub, err := w.nats.Subscribe("governance.request.approved", func(m *nats.Msg) {
		var req models.GovernanceRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			w.logger.Error("failed to unmarshal request for export", zap.Error(err))
			return
		}
		
		if req.RequestType == models.RequestTypeExport {
			w.logger.Info("received export request via NATS", zap.String("id", req.ID.String()))
			_, err := w.svc.ProcessExportRequest(ctx, req.ID)
			if err != nil {
				w.logger.Error("failed to process export", zap.Error(err))
			}
		}
	})
	if err != nil {
		w.logger.Error("failed to subscribe to export events", zap.Error(err))
	} else {
		defer sub.Unsubscribe()
	}

	<-ctx.Done()
}

func (w *ExportProcessingWorker) pollApprovedRequests(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Fallback: poll DB for approved but not processed exports
			// (Implementation skipped for now, focusing on NATS as primary)
		}
	}
}
