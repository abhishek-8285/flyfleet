package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/repository"
	"github.com/flyfleet/fleet-core-service/internal/platform/observability"
)

type SLOService struct {
	repo   repository.MonitoringRepository
	logger *zap.Logger
}

func NewSLOService(repo repository.MonitoringRepository, logger *zap.Logger) *SLOService {
	return &SLOService{repo: repo, logger: logger}
}

func (s *SLOService) CalculateSLOs(ctx context.Context) error {
	ctx, span := observability.StartSpan(ctx, "SLOService.Calculate")
	defer span.End()

	slos, err := s.repo.GetSLOs(ctx)
	if err != nil {
		return err
	}

	for _, slo := range slos {
		if err := s.processSLO(ctx, slo); err != nil {
			s.logger.Error("Failed to process SLO", zap.String("service", slo.ServiceName), zap.Error(err))
		}
	}

	return nil
}

func (s *SLOService) processSLO(ctx context.Context, slo *models.ServiceSLO) error {
	// 1. Get health history for the period
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -slo.PeriodDays)

	history, err := s.repo.GetHealthHistory(ctx, "all", startTime, endTime, 10000)
	if err != nil {
		return err
	}

	// 2. Calculate actual availability
	total := len(history)
	if total == 0 {
		return nil
	}

	downCount := 0
	for _, h := range history {
		if h.Status == models.HealthStatusDown {
			downCount++
		}
	}

	actualAvailability := float64(total-downCount) / float64(total) * 100.0
	s.logger.Info("SLO availability calculated", zap.Float64("availability", actualAvailability), zap.String("service", slo.ServiceName))

	// 3. Update Error Budget
	budget, err := s.repo.GetErrorBudget(ctx, slo.ServiceName)
	if err != nil {
		// Create new budget if not exists
		allowedDowntime := int(float64(slo.PeriodDays*24*60) * (1 - slo.AvailabilityTarget/100.0))
		budget = &models.ErrorBudget{
			ServiceName:           slo.ServiceName,
			AllowedDowntimeMinutes: allowedDowntime,
			ConsumedMinutes:        downCount, // Simplified
			PeriodStart:           startTime,
			PeriodEnd:             endTime,
		}
	} else {
		budget.ConsumedMinutes = downCount // Simplified
		budget.UpdatedAt = time.Now()
	}

	if err := s.repo.SaveErrorBudget(ctx, budget); err != nil {
		return err
	}

	// 4. Trigger alert if burn rate is high
	burnRate := float64(budget.ConsumedMinutes) / float64(budget.AllowedDowntimeMinutes)
	if burnRate > 0.9 {
		s.logger.Warn("SLO Error Budget Burn Rate Critical", 
			zap.String("service", slo.ServiceName),
			zap.Float64("burn_rate", burnRate))
		
		// Create SRE alert group/incident
		groupKey := fmt.Sprintf("%s:slo:critical", slo.ServiceName)
		_, _ = s.repo.GetOrCreateAlertGroup(ctx, groupKey, slo.ServiceName, "slo_availability", "CRITICAL")
	}

	return nil
}
