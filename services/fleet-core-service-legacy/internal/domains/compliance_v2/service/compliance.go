package service

import (
	"context"
	"errors"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/models"
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
	driverID, _ := uuid.Parse(req.DriverId)
	
	// 1. Fetch current status to evaluate transition legality
	currentStatus, err := s.repo.GetLatestDutyStatus(ctx, driverID)
	if err != nil {
		s.logger.Error("failed to get current status", zap.Error(err))
		return nil, err
	}

	// 2. Complex HOS Rule Evaluation (Simplified for test demonstration)
	if currentStatus != nil {
		if mapModelToProtoStatus(currentStatus.Status) == req.Status {
			// No state change
			return &pb.RecordDutyStatusResponse{Log: mapModelToProtoLog(currentStatus)}, nil
		}

		// Example FMCSA Rule: Cannot directly transition from DRIVING to YARD_MOVES without OFF_DUTY or ON_DUTY
		if currentStatus.Status == models.DutyStatusDriving && req.Status == pb.DutyStatus_YARD_MOVES {
			return nil, ErrInvalidTransition
		}
	}

	// 3. Create the new log entry for repository
	var vehicleID *uuid.UUID
	if vID, err := uuid.Parse(req.VehicleId); err == nil {
		vehicleID = &vID
	}

	newModelLog := &models.DriverDutyLog{
		ID:        uuid.New(),
		DriverID:  driverID,
		VehicleID: vehicleID,
		Status:    mapProtoToModelStatus(req.Status),
		StartTime: time.Now(),
		Location:  req.Location,
	}

	// 4. Persist
	if err := s.repo.RecordDutyStatus(ctx, newModelLog); err != nil {
		return nil, err
	}

	return &pb.RecordDutyStatusResponse{
		Log: mapModelToProtoLog(newModelLog),
	}, nil
}

func (s *complianceServiceImpl) GetCurrentDutyStatus(ctx context.Context, driverIDStr string) (*pb.DutyStatusLog, error) {
	driverID, _ := uuid.Parse(driverIDStr)
	log, err := s.repo.GetLatestDutyStatus(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		// New driver, default to OFF_DUTY
		return &pb.DutyStatusLog{
			DriverId: driverIDStr,
			Status:   pb.DutyStatus_OFF_DUTY,
		}, nil
	}
	return mapModelToProtoLog(log), nil
}

func mapModelToProtoStatus(s models.DriverDutyStatus) pb.DutyStatus {
	switch s {
	case models.DutyStatusDriving:
		return pb.DutyStatus_DRIVING
	case models.DutyStatusOnDuty:
		return pb.DutyStatus_ON_DUTY_NOT_DRIVING
	case models.DutyStatusOffDuty:
		return pb.DutyStatus_OFF_DUTY
	default:
		return pb.DutyStatus_OFF_DUTY
	}
}

func mapProtoToModelStatus(s pb.DutyStatus) models.DriverDutyStatus {
	switch s {
	case pb.DutyStatus_DRIVING:
		return models.DutyStatusDriving
	case pb.DutyStatus_ON_DUTY_NOT_DRIVING:
		return models.DutyStatusOnDuty
	case pb.DutyStatus_OFF_DUTY:
		return models.DutyStatusOffDuty
	default:
		return models.DutyStatusOffDuty
	}
}

func mapModelToProtoLog(m *models.DriverDutyLog) *pb.DutyStatusLog {
	p := &pb.DutyStatusLog{
		Id:                  m.ID.String(),
		DriverId:            m.DriverID.String(),
		Status:              mapModelToProtoStatus(m.Status),
		StartTime:           timestamppb.New(m.StartTime),
		LocationDescription: m.Location,
	}
	if m.VehicleID != nil {
		p.VehicleId = m.VehicleID.String()
	}
	return p
}
