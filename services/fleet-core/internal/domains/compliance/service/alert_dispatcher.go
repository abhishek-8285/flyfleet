package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AlertDispatcher interface {
	Dispatch(ctx context.Context, alert *models.ComplianceAlert) error
}

type alertDispatcherImpl struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewAlertDispatcher(repo repository.ComplianceRepository, logger logger.Logger) AlertDispatcher {
	return &alertDispatcherImpl{
		repo:   repo,
		logger: logger,
	}
}

func (d *alertDispatcherImpl) Dispatch(ctx context.Context, alert *models.ComplianceAlert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now()
	}

	// 1. Persist in database
	if err := d.repo.CreateComplianceAlert(ctx, alert); err != nil {
		d.logger.Error("failed to persist compliance alert", zap.Error(err), zap.String("alert_id", alert.ID.String()))
		return err
	}

	// 2. Log for observability
	d.logger.Info("Compliance Alert Dispatched",
		zap.String("id", alert.ID.String()),
		zap.String("type", string(alert.Type)),
		zap.String("severity", string(alert.Severity)),
		zap.String("entity_id", alert.EntityID.String()),
	)

	// 3. Integration with alerts_v2 (Placeholder - would call internal service or event bus)
	// In a real implementation:
	// d.alertsV2.CreateAlert(ctx, &alertsV2.AlertRequest{...})

	// 4. Publish to NATS (Placeholder for Compliance Event Bus)
	// d.eventBus.Publish("compliance.alert.created", alert)

	return nil
}
