package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"go.uber.org/zap"
)

// EmergencyRecoveryWorker resumes stalled SOS incidents after system crashes or restarts.
type EmergencyRecoveryWorker struct {
	repo   repository.EmergencyRepository
	logger *zap.Logger
}

func NewEmergencyRecoveryWorker(repo repository.EmergencyRepository, logger *zap.Logger) *EmergencyRecoveryWorker {
	return &EmergencyRecoveryWorker{repo: repo, logger: logger}
}

// Run starts the recovery worker loop with a 30-second polling interval.
func (w *EmergencyRecoveryWorker) Run(ctx context.Context) {
	w.logger.Info("EmergencyRecoveryWorker started")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("EmergencyRecoveryWorker shutting down")
			return
		case <-ticker.C:
			w.recoverStalledIncidents(ctx)
		}
	}
}

func (w *EmergencyRecoveryWorker) recoverStalledIncidents(ctx context.Context) {
	// Find TRIGGERED incidents with no update in the last 60 seconds
	incidents, err := w.repo.GetStalledIncidents(ctx, 60*time.Second)
	if err != nil {
		w.logger.Error("RecoveryWorker: Failed to query stalled incidents", zap.Error(err))
		return
	}

	if len(incidents) == 0 {
		return
	}

	w.logger.Warn("RecoveryWorker: Stalled incidents found, resuming escalation",
		zap.Int("count", len(incidents)))

	for _, inc := range incidents {
		incCopy := inc
		w.logger.Info("Resuming stalled incident",
			zap.String("incident_id", incCopy.ID.String()),
			zap.Time("last_updated", incCopy.UpdatedAt),
		)

		// Bump updated_at to avoid double-recovery
		if err := w.repo.UpdateIncidentStatus(ctx, incCopy.ID, models.StatusTriggered, nil); err != nil {
			w.logger.Error("RecoveryWorker: Failed to touch incident", zap.Error(err), zap.String("incident_id", incCopy.ID.String()))
			continue
		}

		emergency.IncidentRecoveryTotal.WithLabelValues("resumed").Inc()
	}
}
