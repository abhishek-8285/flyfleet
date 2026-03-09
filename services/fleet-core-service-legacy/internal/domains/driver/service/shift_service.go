package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"go.uber.org/zap"
)

type ShiftRepo interface {
	CreateShift(ctx context.Context, shift *models.Shift) error
	AssignVehicle(ctx context.Context, shiftID string, vehicleID string) error
	UpdateShiftStatus(ctx context.Context, shiftID string, status string) error
	GetShiftsByDriverID(ctx context.Context, driverID string, start, end time.Time) ([]*models.Shift, error)
}

type ShiftService struct {
	repo   ShiftRepo
	logger *zap.Logger
}

func NewShiftService(repo ShiftRepo, logger *zap.Logger) *ShiftService {
	return &ShiftService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ShiftService) CreateShift(ctx context.Context, shift *models.Shift) error {
	shift.Status = "SCHEDULED"
	return s.repo.CreateShift(ctx, shift)
}

func (s *ShiftService) AssignToVehicle(ctx context.Context, shiftID string, vehicleID string) error {
	return s.repo.AssignVehicle(ctx, shiftID, vehicleID)
}

func (s *ShiftService) StartShift(ctx context.Context, shiftID string) error {
	return s.repo.UpdateShiftStatus(ctx, shiftID, "ACTIVE")
}

func (s *ShiftService) EndShift(ctx context.Context, shiftID string) error {
	return s.repo.UpdateShiftStatus(ctx, shiftID, "COMPLETED")
}

func (s *ShiftService) GetShiftsByDriverID(ctx context.Context, driverID string, start, end time.Time) ([]*models.Shift, error) {
	return s.repo.GetShiftsByDriverID(ctx, driverID, start, end)
}
