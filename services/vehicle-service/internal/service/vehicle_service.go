package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/flyfleet/vehicle-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VehicleService struct {
	repo   *repository.VehicleRepository
	logger *zap.Logger
}

func NewVehicleService(repo *repository.VehicleRepository, logger *zap.Logger) *VehicleService {
	return &VehicleService{
		repo:   repo,
		logger: logger,
	}
}

func (s *VehicleService) CreateVehicle(ctx context.Context, req *models.CreateVehicleRequest, orgID uuid.UUID) (*models.Vehicle, error) {
	vehicle := &models.Vehicle{
		ID:           uuid.New(),
		OrgID:        orgID,
		LicensePlate: req.LicensePlate,
		Model:        req.Model,
		Make:         req.Make,
		Year:         req.Year,
		Color:        req.Color,
		Capacity:     req.Capacity,
		FuelType:     req.FuelType,
		VIN:          req.VIN,
		Registration: req.Registration,
		Insurance:    req.Insurance,
		Status:       "active",
		Odometer:     0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if vehicle.Status == "" {
		vehicle.Status = "active"
	}

	if err := s.repo.Create(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to create vehicle: %w", err)
	}

	return vehicle, nil
}

func (s *VehicleService) GetVehicle(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	vehicle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *VehicleService) ListVehicles(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.Vehicle, int64, error) {
	offset := (page - 1) * limit
	vehicles, total, err := s.repo.GetByOrgID(ctx, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convert []*Vehicle to []Vehicle
	result := make([]models.Vehicle, len(vehicles))
	for i, v := range vehicles {
		result[i] = *v
	}

	return result, total, nil
}

func (s *VehicleService) UpdateVehicle(ctx context.Context, id uuid.UUID, req *models.UpdateVehicleRequest) (*models.Vehicle, error) {
	// Get existing vehicle
	vehicle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Model != nil {
		vehicle.Model = *req.Model
	}
	if req.Make != nil {
		vehicle.Make = *req.Make
	}
	if req.Year != nil {
		vehicle.Year = *req.Year
	}
	if req.Color != nil {
		vehicle.Color = *req.Color
	}
	if req.Status != nil {
		vehicle.Status = *req.Status
	}
	if req.Capacity != nil {
		vehicle.Capacity = *req.Capacity
	}
	if req.FuelType != nil {
		vehicle.FuelType = *req.FuelType
	}
	if req.Registration != nil {
		vehicle.Registration = *req.Registration
	}
	if req.Insurance != nil {
		vehicle.Insurance = *req.Insurance
	}
	if req.Odometer != nil {
		vehicle.Odometer = *req.Odometer
	}
	if req.Status != nil {
		vehicle.Status = *req.Status
	}

	if err := s.repo.Update(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to update vehicle: %w", err)
	}

	return vehicle, nil
}

func (s *VehicleService) DeleteVehicle(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}
	return nil
}

func (s *VehicleService) AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) (*models.Vehicle, error) {
	if err := s.repo.AssignDriver(ctx, vehicleID, driverID); err != nil {
		return nil, fmt.Errorf("failed to assign driver: %w", err)
	}

	// Return updated vehicle
	vehicle, err := s.repo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	return vehicle, nil
}

func (s *VehicleService) GetVehicleLocation(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleLocation, error) {
	// Vehicle model doesn't currently store location data
	// This would typically be handled by a separate location tracking service
	return nil, fmt.Errorf("location tracking not implemented in vehicle service")
}
