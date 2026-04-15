package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
)

type DriverRepository interface {
	GetByID(ctx context.Context, id string) (*models.Driver, error)
}

type AvailabilityService struct {
	driverRepo DriverRepository
}

func (s *AvailabilityService) IsDriverAvailable(ctx context.Context, driverID string) (bool, error) {
	driver, err := s.driverRepo.GetByID(ctx, driverID)
	if err != nil {
		return false, err
	}

	if driver == nil {
		return false, nil
	}

	// Logic from tests:
	// - Status must be "active"
	// - CurrentTripID must be empty or nil
	if driver.Status != "active" {
		return false, nil
	}

	if driver.CurrentTripID != nil && *driver.CurrentTripID != "" {
		return false, nil
	}

	return true, nil
}
