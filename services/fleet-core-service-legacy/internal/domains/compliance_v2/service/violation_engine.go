package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ViolationEngine interface {
	EvaluateMoRTHRules(ctx context.Context, driverID uuid.UUID) error
}

type violationEngineImpl struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewViolationEngine(logger logger.Logger, repo repository.ComplianceRepository) ViolationEngine {
	return &violationEngineImpl{
		repo:   repo,
		logger: logger,
	}
}

func (e *violationEngineImpl) EvaluateMoRTHRules(ctx context.Context, driverID uuid.UUID) error {
	// Implement MoRTH HOS rules evaluation
	// 1. Max 8 hours continuous driving
	// 2. Max 12 hours duty in 24 hours
	// 3. Mandatory 30m break after 5h driving
	
	now := time.Now()
	startOf24h := now.Add(-24 * time.Hour)
	
	logs, err := e.repo.GetDutyLogsByDriver(ctx, driverID, startOf24h, now)
	if err != nil {
		return err
	}
	
	var totalDriving time.Duration
	var totalDuty time.Duration
	var continuousDriving time.Duration

	for _, log := range logs {
		endTime := log.EndTime
		if endTime == nil {
			endTime = ptrTime(now)
		}
		duration := endTime.Sub(log.StartTime)

		if log.Status == models.DutyStatusDriving {
			totalDriving += duration
			continuousDriving += duration
			totalDuty += duration
		} else if log.Status == models.DutyStatusOnDuty {
			totalDuty += duration
			continuousDriving = 0 // Break resets continuous driving
		} else if log.Status == models.DutyStatusBreak || log.Status == models.DutyStatusRest || log.Status == models.DutyStatusOffDuty {
			continuousDriving = 0 // Break resets continuous driving
		}
	}

	// Rule 1: Max 8 hours continuous driving (MoRTH strict)
	if continuousDriving.Hours() > 8.0 {
		e.createViolation(ctx, driverID, models.ViolationDrivingLimitExceeded, models.SeverityHigh, "Exceeded 8 hours continuous driving limit")
	}

	// Rule 2: Max 12 hours total duty
	if totalDuty.Hours() > 12.0 {
		e.createViolation(ctx, driverID, models.ViolationDutyLimitExceeded, models.SeverityMedium, "Exceeded 12 hours standard duty limit")
	}
	
	return nil
}

func (e *violationEngineImpl) createViolation(ctx context.Context, driverID uuid.UUID, vType models.ViolationType, severity models.ValidationSeverity, metadata string) {
	violation := &models.DriverComplianceViolation{
		ID:            uuid.New(),
		DriverID:      driverID,
		ViolationType: vType,
		Timestamp:     time.Now(),
		Severity:      severity,
		Metadata:      []byte(`{"reason":"` + metadata + `"}`),
	}
	
	if err := e.repo.RecordViolation(ctx, violation); err != nil {
		e.logger.Error("failed to record violation", zap.Error(err))
	}
}
