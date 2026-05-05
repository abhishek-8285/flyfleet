package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MobileSyncRepo interface {
	QueueItem(ctx context.Context, item *models.MobileSyncItem) error
	GetPendingItems(ctx context.Context, driverID uuid.UUID) ([]*models.MobileSyncItem, error)
	UpdateItemStatus(ctx context.Context, itemID uuid.UUID, status string) error
}

type MobileSyncService struct {
	repo   MobileSyncRepo
	logger *zap.Logger
}

func NewMobileSyncService(repo MobileSyncRepo, logger *zap.Logger) *MobileSyncService {
	return &MobileSyncService{
		repo:   repo,
		logger: logger,
	}
}

func (s *MobileSyncService) QueueOfflineEvent(ctx context.Context, driverID uuid.UUID, actionType string, payload map[string]interface{}) error {
	item := &models.MobileSyncItem{
		DriverID:   driverID,
		ActionType: actionType,
		Payload:    payload,
		Status:     "PENDING",
	}
	return s.repo.QueueItem(ctx, item)
}

func (s *MobileSyncService) ProcessSyncQueue(ctx context.Context, driverID uuid.UUID) error {
	items, err := s.repo.GetPendingItems(ctx, driverID)
	if err != nil {
		return err
	}

	for _, item := range items {
		// In a real implementation, we would dispatch these actions to their respective services
		// For now, we just mark them as synced
		s.logger.Info("processing sync item", zap.String("item_id", item.ID.String()), zap.String("action", item.ActionType))
		err := s.repo.UpdateItemStatus(ctx, item.ID, "SYNCED")
		if err != nil {
			s.logger.Error("failed to update sync item status", zap.Error(err), zap.String("item_id", item.ID.String()))
		}
	}
	return nil
}
