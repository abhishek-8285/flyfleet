package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MeteringService struct {
	repo   *repository.MeteringRepository
	logger *zap.Logger
}

func NewMeteringService(repo *repository.MeteringRepository, logger *zap.Logger) *MeteringService {
	return &MeteringService{
		repo:   repo,
		logger: logger,
	}
}

// RecordUsage logs a usage event from an external source
func (s *MeteringService) RecordUsage(ctx context.Context, event *models.ConsumptionEvent) error {
	if event.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if event.MeterCode == "" {
		return fmt.Errorf("meter_code is required")
	}
	if event.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	err := s.repo.RecordConsumption(ctx, event)
	if err != nil {
		s.logger.Error("failed to record usage", zap.Error(err), zap.String("tenant_id", event.TenantID.String()), zap.String("meter", event.MeterCode))
		return err
	}

	return nil
}

// GetUnbilledUsageByOrg retrieves all usage waiting to be invoiced
func (s *MeteringService) GetUnbilledUsageByOrg(ctx context.Context, tenantID uuid.UUID) ([]models.UsageSummary, error) {
	return s.repo.GetUnbilledUsage(ctx, tenantID)
}

// MarkUsageAsInvoiced marks events as billed
func (s *MeteringService) MarkUsageAsInvoiced(ctx context.Context, tenantID uuid.UUID, invoiceID uuid.UUID) error {
	return s.repo.MarkUsageAsBilled(ctx, tenantID, invoiceID)
}

// ListAvailableMeters returns all billable meters
func (s *MeteringService) ListAvailableMeters(ctx context.Context) ([]models.UsageMeter, error) {
	return s.repo.ListMeters(ctx)
}
