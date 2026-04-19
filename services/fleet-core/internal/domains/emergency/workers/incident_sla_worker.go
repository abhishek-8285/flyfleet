package workers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/emergency"
	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type IncidentSLAWorker struct {
	repo repository.EmergencyRepository
	nc   *nats.Conn
	log  *zap.Logger
}

func NewIncidentSLAWorker(repo repository.EmergencyRepository, nc *nats.Conn, log *zap.Logger) *IncidentSLAWorker {
	return &IncidentSLAWorker{
		repo: repo,
		nc:   nc,
		log:  log,
	}
}

func (w *IncidentSLAWorker) Run(ctx context.Context) {
	w.log.Info("Starting Incident SLA Worker")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Stopping Incident SLA Worker")
			return
		case <-ticker.C:
			w.checkSLAs(ctx)
		}
	}
}

func (w *IncidentSLAWorker) checkSLAs(ctx context.Context) {
	incidents, err := w.repo.ListActiveIncidents(ctx)
	if err != nil {
		w.log.Error("SLA Worker: failed to list active incidents", zap.Error(err))
		return
	}

	for _, incident := range incidents {
		// Only monitor triggered (unacknowledged) incidents
		if incident.Status != models.StatusTriggered {
			continue
		}

		duration := time.Since(incident.Timestamp)
		// Phase 4: Standardize to 2 minutes for enterprise hardening alerts
		limit := 2 * time.Minute

		if duration > limit {
			w.log.Warn("SLA Breach detected",
				zap.String("incident_id", incident.ID.String()),
				zap.String("severity", string(incident.Severity)),
				zap.Duration("duration", duration),
			)
			// Phase 7: Metric increment
			emergency.SLABreachesTotal.WithLabelValues(incident.OrgID.String(), string(incident.Severity)).Inc()
			w.publishSLABreach(ctx, incident, duration)
		}
	}
}

func (w *IncidentSLAWorker) publishSLABreach(ctx context.Context, incident models.SOSIncident, duration time.Duration) {
	// 1. Log to DB
	violation := &models.SLAViolation{
		ID:         uuid.New(),
		IncidentID: incident.ID,
		Severity:   incident.Severity,
		BreachType: "SLA_BREACH_THRESHOLD_EXCEEDED",
		BreachTime: time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := w.repo.LogSLAViolation(ctx, violation); err != nil {
		w.log.Error("Failed to log SLA violation", zap.Error(err))
	}

	// 2. Publish to NATS
	if w.nc == nil {
		return
	}

	payload := map[string]interface{}{
		"violation_id": violation.ID,
		"incident_id":  incident.ID,
		"tenant_id":       incident.OrgID,
		"severity":     incident.Severity,
		"duration":     duration.Seconds(),
		"breached_at":  violation.BreachTime,
	}

	data, _ := json.Marshal(payload)
	// Phase 4: Subject standardized
	if err := w.nc.Publish("emergency.sla.violation", data); err != nil {
		w.log.Error("Failed to publish SLA violation event", zap.Error(err))
	}
}
