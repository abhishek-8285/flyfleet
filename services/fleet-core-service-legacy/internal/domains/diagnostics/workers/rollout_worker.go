package workers

import (
	"context"
	"log"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"
)

type RolloutWorker struct {
	rolloutService *service.FirmwareRolloutService
}

func NewRolloutWorker(rs *service.FirmwareRolloutService) *RolloutWorker {
	return &RolloutWorker{rolloutService: rs}
}

func (w *RolloutWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("RolloutWorker stopped")
			return
		case <-ticker.C:
			w.run(ctx)
		}
	}
}

func (w *RolloutWorker) run(ctx context.Context) {
	log.Println("Running firmware rollout batch monitor")
	// Logic to push new batches if previous succeeded 
}
