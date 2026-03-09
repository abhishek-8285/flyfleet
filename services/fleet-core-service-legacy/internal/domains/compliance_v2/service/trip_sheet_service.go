package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/repository"
	"github.com/google/uuid"
	"github.com/fleet-platform/shared/pkg/logger"
)

type TripSheetService interface {
	CreateTrip(ctx context.Context, driverID, vehicleID uuid.UUID, origin, destination string) (*models.TripSheet, error)
	StartTrip(ctx context.Context, tripID uuid.UUID) error
	EndTrip(ctx context.Context, tripID uuid.UUID, distanceKM float64) error
	GetDriverTrips(ctx context.Context, driverID uuid.UUID) ([]models.TripSheet, error)
}

type tripSheetService struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewTripSheetService(repo repository.ComplianceRepository, logger logger.Logger) TripSheetService {
	return &tripSheetService{
		repo:   repo,
		logger: logger,
	}
}

func (s *tripSheetService) CreateTrip(ctx context.Context, driverID, vehicleID uuid.UUID, origin, destination string) (*models.TripSheet, error) {
	trip := &models.TripSheet{
		ID:          uuid.New(),
		DriverID:    driverID,
		VehicleID:   vehicleID,
		Origin:      origin,
		Destination: destination,
		Status:      models.TripStatusCreated,
		CreatedAt:   time.Now(),
	}

	err := s.repo.CreateTripSheet(ctx, trip)
	if err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *tripSheetService) StartTrip(ctx context.Context, tripID uuid.UUID) error {
	trip, err := s.repo.GetTripSheet(ctx, tripID)
	if err != nil {
		return err
	}

	now := time.Now()
	trip.TripStart = &now
	trip.Status = models.TripStatusInProgress

	return s.repo.UpdateTripSheet(ctx, trip)
}

func (s *tripSheetService) EndTrip(ctx context.Context, tripID uuid.UUID, distanceKM float64) error {
	trip, err := s.repo.GetTripSheet(ctx, tripID)
	if err != nil {
		return err
	}

	now := time.Now()
	trip.TripEnd = &now
	trip.DistanceKM = distanceKM
	trip.Status = models.TripStatusCompleted

	return s.repo.UpdateTripSheet(ctx, trip)
}

func (s *tripSheetService) GetDriverTrips(ctx context.Context, driverID uuid.UUID) ([]models.TripSheet, error) {
	return s.repo.GetTripSheetsByDriver(ctx, driverID)
}
