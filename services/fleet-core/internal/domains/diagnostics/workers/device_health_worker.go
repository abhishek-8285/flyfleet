package workers

import (
	"context"
	"log"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/service"
)

type DeviceHealthWorker struct {
	healthService *service.DeviceHealthService
}

func NewDeviceHealthWorker(hs *service.DeviceHealthService) *DeviceHealthWorker {
	return &DeviceHealthWorker{healthService: hs}
}

func (w *DeviceHealthWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("DeviceHealthWorker stopped")
			return
		case <-ticker.C:
			w.Process(ctx)
		}
	}
}

func (w *DeviceHealthWorker) Process(ctx context.Context) {
	log.Println("Scanning for offline and critical devices")
	// Sweep all devices and invoke AssessHealth for drift detection
}
