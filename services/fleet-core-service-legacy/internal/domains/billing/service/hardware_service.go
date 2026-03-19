package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type HardwareService struct {
	repo   *repository.HardwareRepository
	logger *zap.Logger
}

func NewHardwareService(repo *repository.HardwareRepository, logger *zap.Logger) *HardwareService {
	return &HardwareService{
		repo:   repo,
		logger: logger,
	}
}

func (s *HardwareService) ListProducts(ctx context.Context) ([]models.HardwareProduct, error) {
	return s.repo.ListProducts(ctx)
}

func (s *HardwareService) AssignHardware(ctx context.Context, a *models.DeviceAssignment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = "PENDING_ACTIVATION"
	}
	return s.repo.CreateAssignment(ctx, a)
}

func (s *HardwareService) ProcessActivation(ctx context.Context, deviceID string) error {
	assignment, err := s.repo.GetAssignmentByDevice(ctx, deviceID)
	if err != nil {
		// If no assignment exists, it's a BYOD (Bring Your Own Device) or untracked hardware
		return nil
	}

	if assignment.Status == "ACTIVE" {
		return nil // Already active
	}

	err = s.repo.ActivateAssignment(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to activate hardware: %w", err)
	}

	s.logger.Info("Hardware activated successfully", zap.String("device_id", deviceID))
	return nil
}

func (s *HardwareService) GetAssignment(ctx context.Context, deviceID string) (*models.DeviceAssignment, error) {
	return s.repo.GetAssignmentByDevice(ctx, deviceID)
}
