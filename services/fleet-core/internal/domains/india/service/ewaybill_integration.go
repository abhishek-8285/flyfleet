package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/india/models"
	"github.com/flyfleet/fleet-core/internal/domains/india/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EWayBillIntegrationService defines the interface for NIC E-Way Bill integration
type EWayBillIntegrationService interface {
	GenerateEWayBill(ctx context.Context, bill *models.EWayBill) (*models.EWayBill, error)
	CancelEWayBill(ctx context.Context, orgID uuid.UUID, number string, reason string) error
	FetchEWayBillStatus(ctx context.Context, orgID uuid.UUID, number string) (*models.EWayBill, error)
	SyncEWayBills(ctx context.Context) error
}

type EWayBillIntegrationServiceImpl struct {
	repo   repository.IndiaIntegrationRepository
	logger *zap.Logger
}

func NewEWayBillIntegrationService(repo repository.IndiaIntegrationRepository, logger *zap.Logger) *EWayBillIntegrationServiceImpl {
	return &EWayBillIntegrationServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *EWayBillIntegrationServiceImpl) GenerateEWayBill(ctx context.Context, bill *models.EWayBill) (*models.EWayBill, error) {
	s.logger.Info("generating e-way bill", zap.String("doc_number", bill.DocumentNumber))

	// Mock NIC API Call
	bill.EWayBillNumber = fmt.Sprintf("EWB-%d", time.Now().UnixNano())
	bill.EWayBillDate = time.Now()
	bill.EWayBillValidTill = time.Now().Add(24 * time.Hour)
	bill.CreatedAt = time.Now()

	log := &models.IntegrationLog{
		ServiceType:     "ewaybill",
		RequestPayload:  bill,
		ResponsePayload: bill,
		Status:          200,
		OrganizationID:  &bill.OrganizationID,
	}
	_ = s.repo.CreateIntegrationLog(ctx, log)

	if err := s.repo.CreateEWayBill(ctx, bill); err != nil {
		return nil, err
	}

	return bill, nil
}

func (s *EWayBillIntegrationServiceImpl) CancelEWayBill(ctx context.Context, orgID uuid.UUID, number string, reason string) error {
	s.logger.Info("cancelling e-way bill", zap.String("number", number), zap.String("reason", reason))

	log := &models.IntegrationLog{
		ServiceType: "ewaybill",
		RequestPayload: map[string]interface{}{
			"action": "cancel",
			"number": number,
			"reason": reason,
		},
		ResponsePayload: map[string]string{"status": "cancelled"},
		Status:          200,
		OrganizationID:  &orgID,
	}
	_ = s.repo.CreateIntegrationLog(ctx, log)

	// Update in repo
	return s.repo.UpdateEWayBill(ctx, number, map[string]interface{}{
		"cancelled":        true,
		"cancelled_reason": reason,
		"cancelled_on":     time.Now(),
	})
}

func (s *EWayBillIntegrationServiceImpl) FetchEWayBillStatus(ctx context.Context, orgID uuid.UUID, number string) (*models.EWayBill, error) {
	s.logger.Info("fetching e-way bill status", zap.String("number", number))

	// Mock status fetch
	bill, err := s.repo.GetEWayBillByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	log := &models.IntegrationLog{
		ServiceType:     "ewaybill",
		RequestPayload:  map[string]string{"action": "get_status", "number": number},
		ResponsePayload: bill,
		Status:          200,
		OrganizationID:  &orgID,
	}
	_ = s.repo.CreateIntegrationLog(ctx, log)

	return bill, nil
}

func (s *EWayBillIntegrationServiceImpl) SyncEWayBills(ctx context.Context) error {
	s.logger.Info("starting e-way bill sync")
	// Logic to poll NIC for updates on pending/active bills
	return nil
}
