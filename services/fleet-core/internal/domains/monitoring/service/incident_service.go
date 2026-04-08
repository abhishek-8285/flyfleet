package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/monitoring/models"
	"github.com/flyfleet/fleet-core/internal/domains/monitoring/repository"
	"go.uber.org/zap"
)

type IncidentService struct {
	repo   repository.MonitoringRepository
	logger *zap.Logger
}

func NewIncidentService(repo repository.MonitoringRepository, logger *zap.Logger) *IncidentService {
	return &IncidentService{repo: repo, logger: logger}
}

func (s *IncidentService) CreateFromAlertGroup(ctx context.Context, group *models.AlertGroup) error {
	incident := &models.Incident{
		Title:       "Incident triggered by " + group.MetricName,
		Severity:    group.Severity,
		Status:      "DETECTED",
		GroupKey:    group.GroupKey,
		ServiceName: group.Component,
		DetectedAt:  time.Now(),
	}

	return s.repo.CreateIncident(ctx, incident)
}
