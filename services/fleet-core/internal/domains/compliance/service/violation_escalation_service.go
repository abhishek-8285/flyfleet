package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ViolationEscalationService interface {
	CheckAndEscalate(ctx context.Context, driverID uuid.UUID) error
}

type violationEscalationServiceImpl struct {
	repo            repository.ComplianceRepository
	alertDispatcher AlertDispatcher
	logger          logger.Logger
}

func NewViolationEscalationService(repo repository.ComplianceRepository, alertDispatcher AlertDispatcher, logger logger.Logger) ViolationEscalationService {
	return &violationEscalationServiceImpl{
		repo:            repo,
		alertDispatcher: alertDispatcher,
		logger:          logger,
	}
}

func (s *violationEscalationServiceImpl) CheckAndEscalate(ctx context.Context, driverID uuid.UUID) error {
	since := time.Now().Add(-24 * time.Hour)
	violations, err := s.repo.GetViolationsForEscalation(ctx, driverID, since)
	if err != nil {
		s.logger.Error("failed to fetch violations for escalation check", zap.Error(err), zap.String("driver_id", driverID.String()))
		return err
	}

	count := len(violations)
	if count == 0 {
		return nil
	}

	// 1. Policy: 3 violations -> fleet manager notification
	if count >= 3 && count < 5 {
		alert := &models.ComplianceAlert{
			ID:             uuid.New(),
			TenantID: uuid.Nil, // In real app, associate from driver context
			Type:           models.AlertDriverViolation,
			EntityType:     models.EntityDriver,
			EntityID:       driverID,
			Severity:       models.AlertSeverityHigh,
		}

		if err := s.alertDispatcher.Dispatch(ctx, alert); err != nil {
			s.logger.Error("failed to dispatch violation escalation alert", zap.Error(err))
		}
	}

	// 2. Policy: 5 violations -> driver suspension flag
	if count >= 5 {
		s.logger.Warn("DRIVER SUSPENSION RECOMMENDED - Chronic Violations",
			zap.String("driver_id", driverID.String()),
			zap.Int("violation_count", count),
		)

		alert := &models.ComplianceAlert{
			ID:             uuid.New(),
			TenantID: uuid.Nil,
			Type:           models.AlertDriverViolation,
			EntityType:     models.EntityDriver,
			EntityID:       driverID,
			Severity:       models.AlertSeverityCritical,
		}

		if err := s.alertDispatcher.Dispatch(ctx, alert); err != nil {
			s.logger.Error("failed to dispatch critical violation alert", zap.Error(err))
		}

		// Typically would call auth_v2 or driver service to set suspension flag
		// s.driverService.Suspend(ctx, driverID, "Persistent compliance violations")
	}

	// 3. Mark the violations as escalated so we don't count them again for escalation next time
	for _, v := range violations {
		if err := s.repo.EscalateViolation(ctx, v.ID); err != nil {
			s.logger.Error("failed to mark violation as escalated", zap.Error(err), zap.String("id", v.ID.String()))
		}
	}

	return nil
}
