package service

import (
	"context"
	"errors"

	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/repository"
	pb "github.com/flyfleet/proto/compliance/v1"
	"github.com/google/uuid"
	"github.com/fleet-platform/shared/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrDriverNotFound    = errors.New("driver not found")
	ErrInvalidTransition = errors.New("invalid HOS state transition")
	ErrHOSViolation      = errors.New("FMCSA HOS violation: exceeding driving limits")
)

type ComplianceService interface {
	RecordDutyStatus(ctx context.Context, req *pb.RecordDutyStatusRequest) (*pb.RecordDutyStatusResponse, error)
	GetCurrentDutyStatus(ctx context.Context, driverID string) (*pb.DutyStatusLog, error)
}

type complianceServiceImpl struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewComplianceService(logger logger.Logger, repo repository.ComplianceRepository) ComplianceService {
	return &complianceServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// RecordDutyStatus simulates the FMCSA ELD State Machine
func (s *complianceServiceImpl) RecordDutyStatus(ctx context.Context, req *pb.RecordDutyStatusRequest) (*pb.RecordDutyStatusResponse, error) {
	// 1. Fetch current status to evaluate transition legality
	currentStatus, err := s.repo.GetLatestDutyStatus(ctx, req.DriverId)
	if err != nil {
		s.logger.Error("failed to get current status", zap.Error(err))
		return nil, err
	}

	// 2. Complex HOS Rule Evaluation (Simplified for test demonstration)
	if currentStatus != nil {
		if currentStatus.Status == req.Status {
			// No state change
			return &pb.RecordDutyStatusResponse{Log: currentStatus}, nil
		}

		// Example FMCSA Rule: Cannot directly transition from DRIVING to YARD_MOVES without OFF_DUTY or ON_DUTY
		if currentStatus.Status == pb.DutyStatus_DRIVING && req.Status == pb.DutyStatus_YARD_MOVES {
			return nil, ErrInvalidTransition
		}
	}

	// 3. Create the new log entry
	newLog := &pb.DutyStatusLog{
		Id:                  uuid.NewString(),
		DriverId:            req.DriverId,
		VehicleId:           req.VehicleId,
		Status:              req.Status,
		StartTime:           timestamppb.Now(),
		LocationDescription: req.Location,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		OdometerStart:       req.Odometer,
		EngineHoursStart:    req.EngineHours,
	}

	// 4. Persist
	if err := s.repo.RecordDutyStatus(ctx, newLog); err != nil {
		return nil, err
	}

	return &pb.RecordDutyStatusResponse{
		Log: newLog,
	}, nil
}

func (s *complianceServiceImpl) GetCurrentDutyStatus(ctx context.Context, driverID string) (*pb.DutyStatusLog, error) {
	log, err := s.repo.GetLatestDutyStatus(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		// New driver, default to OFF_DUTY
		return &pb.DutyStatusLog{
			DriverId: driverID,
			Status:   pb.DutyStatus_OFF_DUTY,
		}, nil
	}
	return log, nil
}
