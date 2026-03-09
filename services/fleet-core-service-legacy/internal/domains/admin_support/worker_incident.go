package admin_support

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// IncidentWorker monitors unresolved incidents and sends alerts.
type IncidentWorker struct {
	repo     Repository
	metrics  *Metrics
	logger   *zap.Logger
	interval time.Duration
	stopCh   chan struct{}
}

// NewIncidentWorker creates a new incident monitoring worker.
func NewIncidentWorker(repo Repository, metrics *Metrics, logger *zap.Logger) *IncidentWorker {
	return &IncidentWorker{
		repo:     repo,
		metrics:  metrics,
		logger:   logger,
		interval: 1 * time.Minute, // Check every minute for SEV1
		stopCh:   make(chan struct{}),
	}
}

// Start begins the incident monitoring loop.
func (w *IncidentWorker) Start(ctx context.Context) {
	w.logger.Info("Incident worker started", zap.Duration("interval", w.interval))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Incident worker stopping due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("Incident worker stopping due to stop signal")
			return
		case <-ticker.C:
			w.checkUnresolvedIncidents(ctx)
		}
	}
}

// Stop gracefully stops the worker.
func (w *IncidentWorker) Stop() {
	close(w.stopCh)
}

// checkUnresolvedIncidents checks for unresolved SEV1 incidents and sends alerts.
func (w *IncidentWorker) checkUnresolvedIncidents(ctx context.Context) {
	w.logger.Debug("checking for unresolved incidents")

	incidents, err := w.repo.ListAllIncidents(ctx)
	if err != nil {
		w.logger.Error("failed to list incidents", zap.Error(err))
		return
	}

	for _, inc := range incidents {
		if inc.Status == IncidentOpen || inc.Status == IncidentMitigating {
			w.handleUnresolvedIncident(inc)
		}
	}
}

// handleUnresolvedIncident sends alerts for high-severity unresolved incidents.
func (w *IncidentWorker) handleUnresolvedIncident(inc *Incident) {
	switch inc.Severity {
	case Sev1:
		w.logger.Error("SEV1 INCIDENT UNRESOLVED - ALERTING",
			zap.String("incident_id", inc.ID.String()),
			zap.String("description", inc.Description),
			zap.Duration("duration", time.Since(inc.DetectedAt)),
		)
		// TODO: Integrate with PagerDuty
		// notify.PagerDuty.SendAlert(...)

	case Sev2:
		w.logger.Warn("SEV2 INCIDENT UNRESOLVED",
			zap.String("incident_id", inc.ID.String()),
			zap.String("description", inc.Description),
		)
		// TODO: Send email/Slack notification
		// notify.Slack.SendMessage(...)

	default:
		w.logger.Debug("lower severity incident unresolved",
			zap.String("incident_id", inc.ID.String()),
			zap.String("severity", string(inc.Severity)),
		)
	}
}

// ListAllIncidents fetches all incidents from the repository.
// This method should be added to the Repository interface.
func (r *repository) ListAllIncidents(ctx context.Context) ([]*Incident, error) {
	query := `SELECT id, severity, status, description, detected_at, resolved_at, created_at, updated_at FROM incidents ORDER BY detected_at DESC`
	// Note: This is a simplified query - in production, you might want pagination
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		var inc Incident
		if err := rows.Scan(&inc.ID, &inc.Severity, &inc.Status, &inc.Description, &inc.DetectedAt, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, &inc)
	}
	return incidents, nil
}
