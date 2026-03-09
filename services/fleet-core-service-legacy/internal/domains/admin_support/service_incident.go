package admin_support

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *service) DeclareIncident(ctx context.Context, severity IncidentSeverity, description string) (*Incident, error) {
	inc := &Incident{
		ID:          uuid.New(),
		Severity:    severity,
		Status:      IncidentOpen,
		Description: description,
		DetectedAt:  time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateIncident(ctx, inc); err != nil {
		s.logger.Error("failed to declare incident", zap.Error(err))
		return nil, err
	}

	s.metrics.IncidentsTotal.WithLabelValues(string(severity)).Inc()
	s.logger.Warn("INCIDENT DECLARED", 
		zap.String("incident_id", inc.ID.String()), 
		zap.String("severity", string(severity)),
		zap.String("description", description),
	)

	// Severity-based alerting integration
	s.triggerAlerts(inc)

	return inc, nil
}

func (s *service) UpdateIncident(ctx context.Context, incidentID uuid.UUID, updateText string) error {
	u := &IncidentUpdate{
		ID:            uuid.New(),
		IncidentID:    incidentID,
		UpdateMessage: updateText,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateIncidentUpdate(ctx, u); err != nil {
		return err
	}

	// Update the incident's updated_at timestamp
	inc, err := s.repo.GetIncident(ctx, incidentID)
	if err == nil {
		inc.UpdatedAt = time.Now()
		_ = s.repo.UpdateIncident(ctx, inc)
	}

	s.logger.Info("incident update posted", zap.String("incident_id", incidentID.String()))
	return nil
}

func (s *service) ResolveIncident(ctx context.Context, incidentID uuid.UUID) error {
	inc, err := s.repo.GetIncident(ctx, incidentID)
	if err != nil {
		return err
	}

	now := time.Now()
	inc.Status = IncidentResolved
	inc.ResolvedAt = &now
	inc.UpdatedAt = now

	if err := s.repo.UpdateIncident(ctx, inc); err != nil {
		return err
	}

	s.logger.Info("incident resolved", zap.String("incident_id", incidentID.String()))
	return nil
}

func (s *service) ListIncidents(ctx context.Context) ([]*Incident, error) {
	return s.repo.ListAllIncidents(ctx)
}

func (s *service) triggerAlerts(inc *Incident) {
	switch inc.Severity {
	case Sev1:
		s.logger.Info("SYSTEM ALERT: SEV1 - Triggering PagerDuty", zap.String("incident_id", inc.ID.String()))
		// TODO: Integrate with fleet-platform/shared/notify/pagerduty
	case Sev2:
		s.logger.Info("SYSTEM ALERT: SEV2 - Triggering Email Alerts", zap.String("incident_id", inc.ID.String()))
		// TODO: Integrate with fleet-platform/shared/notify/email
	case Sev3:
		s.logger.Info("SYSTEM ALERT: SEV3 - Log Warning", zap.String("incident_id", inc.ID.String()))
	default:
		s.logger.Info("SYSTEM ALERT: SEV4 - Log Info", zap.String("incident_id", inc.ID.String()))
	}
}
