package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
)

type ProcurementService struct {
	poRepo        *repository.MaintenanceRepositoryImpl
	inventoryRepo *repository.InventoryRepository
	logger        *zap.Logger
}

func NewProcurementService(
	poRepo *repository.MaintenanceRepositoryImpl,
	inventoryRepo *repository.InventoryRepository,
	logger *zap.Logger,
) *ProcurementService {
	return &ProcurementService{
		poRepo:        poRepo,
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// AutoCreatePOIfLowStock checks inventory and creates a PO if stock is below reorder level
func (s *ProcurementService) AutoCreatePOIfLowStock(ctx context.Context, orgID uuid.UUID, partID uuid.UUID) error {
	part, err := s.inventoryRepo.GetByID(ctx, partID)
	if err != nil {
		return err
	}
	if part == nil {
		return fmt.Errorf("part not found")
	}

	if part.QuantityAvailable <= part.ReorderLevel {
		// Create Draft PO
		po := &models.PurchaseOrder{
			OrganizationID: orgID,
			OrderNumber:    fmt.Sprintf("AUTO-PO-%d", time.Now().Unix()),
			Status:         models.POStatusDraft,
			TotalCost:      float64(part.ReorderLevel*2) * part.UnitCost, // Simple heuristic for order quantity
			Items: []models.PurchaseOrderItem{
				{
					PartID:    &part.ID,
					Quantity:  part.ReorderLevel * 2,
					UnitCost:  part.UnitCost,
					TotalCost: float64(part.ReorderLevel*2) * part.UnitCost,
				},
			},
		}

		if err := s.poRepo.CreatePO(ctx, po); err != nil {
			s.logger.Error("failed to auto-create PO", zap.Error(err))
			return err
		}
		s.logger.Info("auto-created PO for low stock", zap.String("order_number", po.OrderNumber))
	}

	return nil
}

// ReceivePO updates inventory stocks when a PO is marked as RECEIVED
func (s *ProcurementService) ReceivePO(ctx context.Context, poID uuid.UUID) error {
	po, err := s.poRepo.GetPOByID(ctx, poID)
	if err != nil {
		return err
	}
	if po == nil {
		return fmt.Errorf("PO not found")
	}

	if po.Status == models.POStatusReceived {
		return fmt.Errorf("PO already received")
	}

	// Update stocks
	for _, item := range po.Items {
		if item.PartID != nil {
			if err := s.inventoryRepo.RestoreStock(ctx, *item.PartID, item.Quantity); err != nil {
				s.logger.Error("failed to update stock for received PO item", zap.Error(err))
				return err
			}
		}
	}

	// Update PO status
	return s.poRepo.UpdatePOStatus(ctx, poID, models.POStatusReceived)
}
