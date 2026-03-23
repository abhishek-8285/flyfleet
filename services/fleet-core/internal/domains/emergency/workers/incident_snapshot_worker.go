package workers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EmergencySnapshotWorker takes daily snapshots of critical incidents for disaster recovery.
type EmergencySnapshotWorker struct {
	repo repository.EmergencyRepository
	log  *zap.Logger
}

func NewEmergencySnapshotWorker(repo repository.EmergencyRepository, log *zap.Logger) *EmergencySnapshotWorker {
	return &EmergencySnapshotWorker{
		repo: repo,
		log:  log,
	}
}

func (w *EmergencySnapshotWorker) Run(ctx context.Context) {
	w.log.Info("Starting Emergency Snapshot Worker")
	
	// Initial run
	w.takeSnapshots(ctx)

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Stopping Emergency Snapshot Worker")
			return
		case <-ticker.C:
			w.takeSnapshots(ctx)
		}
	}
}

func (w *EmergencySnapshotWorker) takeSnapshots(ctx context.Context) {
	w.log.Info("Snapshot Worker: performing daily critical incident backup")

	// Get all critical incidents from the last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	incidents, err := w.repo.GetCriticalIncidentsForSnapshot(ctx, since)
	if err != nil {
		w.log.Error("Failed to fetch critical incidents for snapshot", zap.Error(err))
		return
	}

	for _, inc := range incidents {
		data, _ := json.Marshal(inc)
		snapshot := &models.IncidentSnapshot{
			ID:           uuid.New(),
			IncidentID:   inc.ID,
			VehicleID:    inc.VehicleID,
			OrgID:        inc.OrgID,
			IncidentType: inc.IncidentType,
			Severity:     inc.Severity,
			Status:       inc.Status,
			Latitude:     inc.Latitude,
			Longitude:    inc.Longitude,
			SnapshotData: data,
			CreatedAt:    time.Now(),
		}

		if err := w.repo.CreateIncidentSnapshot(ctx, snapshot); err != nil {
			w.log.Error("Failed to create incident snapshot", 
				zap.Error(err), 
				zap.String("incident_id", inc.ID.String()))
		} else {
			emergency.BackupSnapshotsTotal.Inc()
		}
	}

	w.log.Info("Completed daily snapshots", zap.Int("count", len(incidents)))
}
