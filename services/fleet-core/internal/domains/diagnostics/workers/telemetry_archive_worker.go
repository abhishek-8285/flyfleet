package workers

import (
	"context"
	"log"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/repository"
)

type TelemetryArchiveWorker struct {
	repo *repository.DiagnosticsRepository
}

func NewTelemetryArchiveWorker(repo *repository.DiagnosticsRepository) *TelemetryArchiveWorker {
	return &TelemetryArchiveWorker{repo: repo}
}

func (w *TelemetryArchiveWorker) Start(ctx context.Context) {
	// Runs once a day
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("TelemetryArchiveWorker stopped")
			return
		case <-ticker.C:
			w.run(ctx)
		}
	}
}

func (w *TelemetryArchiveWorker) run(ctx context.Context) {
	log.Println("Running telemetry archival job")

	// Rule 1: Move raw telemetry over 30 days to archive table
	// e.g. INSERT INTO telemetry_archive SELECT * FROM device_telemetry WHERE timestamp < NOW() - INTERVAL '30 days'

	// Rule 2: Delete raw telemetry over 30 days
	// e.g. DELETE FROM device_telemetry WHERE timestamp < NOW() - INTERVAL '30 days'

	// Rule 3: Purge aggregated telemetry over 1 year
	// e.g. DELETE FROM telemetry_archive WHERE timestamp < NOW() - INTERVAL '1 year'

	log.Println("Archival job completed")
}
