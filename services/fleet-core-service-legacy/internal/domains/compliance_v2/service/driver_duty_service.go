package service

import (
	"context"
	"errors"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrInvalidDutyTransition = errors.New("invalid MoRTH duty transition")
)

type DriverDutyService interface {
	RecordDutyStatus(ctx context.Context, driverID, vehicleID uuid.UUID, status models.DriverDutyStatus, location, notes string) (*models.DriverDutyLog, error)
	GetLatestDutyStatus(ctx context.Context, driverID uuid.UUID) (*models.DriverDutyLog, error)
	
	RecordBreak(ctx context.Context, driverID uuid.UUID, startTime time.Time, details models.BreakType) (*models.DriverBreakLog, error)
}

type driverDutyServiceImpl struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewDriverDutyService(logger logger.Logger, repo repository.ComplianceRepository) DriverDutyService {
	return &driverDutyServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *driverDutyServiceImpl) RecordDutyStatus(ctx context.Context, driverID, vehicleID uuid.UUID, status models.DriverDutyStatus, location, notes string) (*models.DriverDutyLog, error) {
	currentStatus, err := s.repo.GetLatestDutyStatus(ctx, driverID)
	if err != nil {
		s.logger.Error("failed to get current duty status", zap.Error(err))
		return nil, err
	}

	if currentStatus != nil {
		if currentStatus.Status == status {
			return currentStatus, nil // No change
		}
	}

	newLog := &models.DriverDutyLog{
		ID:        uuid.New(),
		DriverID:  driverID,
		VehicleID: &vehicleID,
		Status:    status,
		StartTime: time.Now(),
		Location:  location,
		Notes:     notes,
	}

	if err := s.repo.RecordDutyStatus(ctx, newLog); err != nil {
		return nil, err
	}

	// 1. Record Audit Log
	audit := &models.ComplianceAuditLog{
		ID:         uuid.New(),
		ActorID:    driverID, // Assume driver is the actor for now
		Action:     "driver_duty_updated",
		EntityType: "DRIVER",
		EntityID:   driverID,
		Timestamp:  time.Now(),
	}
	_ = s.repo.RecordAuditLog(ctx, audit)

	// 2. Publish Event
	// s.eventBus.Publish(ctx, ComplianceEvent{...})

	return newLog, nil
}

func (s *driverDutyServiceImpl) GetLatestDutyStatus(ctx context.Context, driverID uuid.UUID) (*models.DriverDutyLog, error) {
	log, err := s.repo.GetLatestDutyStatus(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		// Default to Off Duty 
		return &models.DriverDutyLog{
			DriverID: driverID,
			Status:   models.DutyStatusOffDuty,
		}, nil
	}
	return log, nil
}

func (s *driverDutyServiceImpl) RecordBreak(ctx context.Context, driverID uuid.UUID, startTime time.Time, breakType models.BreakType) (*models.DriverBreakLog, error) {
	// Create Break Log
	newBreak := &models.DriverBreakLog{
		ID:        uuid.New(),
		DriverID:  driverID,
		StartTime: startTime,
		BreakType: breakType,
	}

	if err := s.repo.RecordBreakLog(ctx, newBreak); err != nil {
		return nil, err
	}
	
	return newBreak, nil
}
