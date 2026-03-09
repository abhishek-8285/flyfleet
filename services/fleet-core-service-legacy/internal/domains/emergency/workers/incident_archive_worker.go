package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"go.uber.org/zap"
)

type IncidentArchiveWorker struct {
	repo repository.EmergencyRepository
	log  *zap.Logger
}

func NewIncidentArchiveWorker(repo repository.EmergencyRepository, log *zap.Logger) *IncidentArchiveWorker {
	return &IncidentArchiveWorker{
		repo: repo,
		log:  log,
	}
}

func (w *IncidentArchiveWorker) Run(ctx context.Context) {
	w.log.Info("Starting Incident Archive Worker")
	// Run once immediately on start, then every 24 hours
	w.archiveOldData(ctx)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Stopping Incident Archive Worker")
			return
		case <-ticker.C:
			w.archiveOldData(ctx)
		}
	}
}

func (w *IncidentArchiveWorker) archiveOldData(ctx context.Context) {
	w.log.Info("Archive Worker: running data retention policy")

	// 1. Snapshot retention: 90 days
	snapshotLimit := time.Now().AddDate(0, 0, -90)
	rows, err := w.repo.ArchiveOldSnapshots(ctx, snapshotLimit)
	if err != nil {
		w.log.Error("Failed to archive old snapshots", zap.Error(err))
	} else {
		w.log.Info("Archived old snapshots", zap.Int64("count", rows))
	}

	// 2. Incident archive: 2 years (marking as ARCHIVED)
	incidentLimit := time.Now().AddDate(-2, 0, 0)
	rows, err = w.repo.ArchiveOldIncidents(ctx, incidentLimit)
	if err != nil {
		w.log.Error("Failed to archive old incidents", zap.Error(err))
	} else {
		w.log.Info("Marked old incidents as ARCHIVED", zap.Int64("count", rows))
	}

	// 3. Media retention: 1 year (Enterprise requirement)
	mediaLimit := time.Now().AddDate(-1, 0, 0)
	rows, err = w.repo.ArchiveOldMedia(ctx, mediaLimit)
	if err != nil {
		w.log.Error("Failed to archive old media", zap.Error(err))
	} else {
		w.log.Info("Deleted old incident media (1yr retention)", zap.Int64("count", rows))
	}
}
