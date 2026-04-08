package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/compliance/repository"
	"github.com/google/uuid"
)

type BorderMonitorService interface {
	ProcessPosition(ctx context.Context, vehicleID uuid.UUID, lat, lon float64, currentState string) error
}

type borderMonitorServiceImpl struct {
	repo repository.ComplianceRepository
}

func NewBorderMonitorService(repo repository.ComplianceRepository) BorderMonitorService {
	return &borderMonitorServiceImpl{
		repo: repo,
	}
}

func (s *borderMonitorServiceImpl) ProcessPosition(ctx context.Context, vehicleID uuid.UUID, lat, lon float64, currentState string) error {
	// In a real implementation:
	// 1. Fetch vehicle's last state border crossing from DB
	// 2. Perform point-in-polygon check for (lat, lon) against state boundaries
	// 3. Detect state changes

	// Mock logic: randomly detects border crossing for simulation
	// s.logger.Info("Processing interstate border check", zap.String("vehicle_id", vehicleID.String()), zap.String("state", currentState))

	// Placeholder for persistent state check
	// lastState := s.repo.GetLastState(ctx, vehicleID)
	// if currentState != lastState {
	// 	crossing := &models.StateBorderCrossing{
	// 		ID:          uuid.New(),
	// 		VehicleID:   vehicleID,
	// 		FromState:   lastState,
	// 		ToState:     currentState,
	// 		Timestamp:   time.Now(),
	// 		GPSLocation: fmt.Sprintf("%f,%f", lat, lon),
	// 	}
	// 	return s.repo.RecordBorderCrossing(ctx, crossing)
	// }

	_ = currentState // Prevent unused variable error

	return nil
}
