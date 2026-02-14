package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/driver-service/internal/models"
	"github.com/flyfleet/driver-service/internal/repository"
)

type DriverService struct {
	repo   *repository.DriverRepository
	logger *zap.Logger
}

func NewDriverService(repo *repository.DriverRepository, logger *zap.Logger) *DriverService {
	return &DriverService{repo: repo, logger: logger}
}

// Create creates a new driver
func (s *DriverService) Create(ctx context.Context, orgID string, req *models.CreateDriverRequest) (*models.Driver, error) {
	// Validate license expiry date
	if req.LicenseExpiry.Before(time.Now()) {
		return nil, fmt.Errorf("license has expired")
	}

	// Check if driver with same license number already exists
	existingDriver, _ := s.GetByLicenseNumber(ctx, req.LicenseNumber)
	if existingDriver != nil {
		return nil, fmt.Errorf("driver with license number %s already exists", req.LicenseNumber)
	}

	driver := &models.Driver{
		ID:                uuid.New().String(),
		OrgID:             orgID,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		FullName:          fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		Email:             req.Email,
		Phone:             req.Phone,
		LicenseNumber:     req.LicenseNumber,
		LicenseType:       &req.LicenseType,
		LicenseExpiry:     req.LicenseExpiry,
		LicenseAuthority:  req.LicenseAuthority,
		DateOfBirth:       req.DateOfBirth,
		Address:           req.Address,
		EmergencyContact:  req.EmergencyContact,
		Status:            "active",
		JoiningDate:       time.Now(),
		ExperienceYears:   req.ExperienceYears,
		Rating:            0.0,
		TotalTrips:        0,
		TotalDistanceKm:   0.0,
		TotalHoursDriven:  0.0,
		BehaviorScore:     100, // Start with perfect score
		Metadata:          req.Metadata,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.Create(ctx, driver); err != nil {
		s.logger.Error("failed to create driver", zap.Error(err))
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	s.logger.Info("driver created", zap.String("driver_id", driver.ID))
	return driver, nil
}

// GetByID retrieves a driver by ID
func (s *DriverService) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	driver, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get driver", zap.String("driver_id", id), zap.Error(err))
		return nil, fmt.Errorf("driver not found: %w", err)
	}
	return driver, nil
}

// List retrieves drivers with filtering and pagination
func (s *DriverService) List(ctx context.Context, filter models.DriverListFilter) ([]*models.Driver, int, error) {
	drivers, totalCount, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list drivers", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list drivers: %w", err)
	}
	return drivers, totalCount, nil
}

// Update modifies an existing driver
func (s *DriverService) Update(ctx context.Context, id string, req *models.UpdateDriverRequest) (*models.Driver, error) {
	// Validate that driver exists
	driver, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.FirstName != nil || req.LastName != nil {
		firstName := driver.FirstName
		lastName := driver.LastName
		if req.FirstName != nil {
			firstName = *req.FirstName
		}
		if req.LastName != nil {
			lastName = *req.LastName
		}
		updates["full_name"] = fmt.Sprintf("%s %s", firstName, lastName)
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.LicenseNumber != nil {
		updates["license_number"] = *req.LicenseNumber
	}
	if req.LicenseType != nil {
		updates["license_type"] = *req.LicenseType
	}
	if req.LicenseExpiry != nil {
		updates["license_expiry"] = *req.LicenseExpiry
	}
	if req.LicenseAuthority != nil {
		updates["license_authority"] = *req.LicenseAuthority
	}
	if req.DateOfBirth != nil {
		updates["date_of_birth"] = *req.DateOfBirth
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.EmergencyContact != nil {
		updates["emergency_contact"] = *req.EmergencyContact
	}
	if req.ExperienceYears != nil {
		updates["experience_years"] = *req.ExperienceYears
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		
		if err := s.repo.Update(ctx, id, updates); err != nil {
			s.logger.Error("failed to update driver", zap.String("driver_id", id), zap.Error(err))
			return nil, fmt.Errorf("failed to update driver: %w", err)
		}
	}

	return s.GetByID(ctx, id)
}

// Delete performs soft delete of a driver
func (s *DriverService) Delete(ctx context.Context, id string) error {
	// Check if driver has active assignments
	driver, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if driver.CurrentVehicleID != nil {
		return fmt.Errorf("cannot delete driver with active vehicle assignment")
	}

	if driver.CurrentTripID != nil {
		return fmt.Errorf("cannot delete driver with active trip")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete driver", zap.String("driver_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete driver: %w", err)
	}

	s.logger.Info("driver deleted", zap.String("driver_id", id))
	return nil
}

// AssignVehicle assigns a vehicle to a driver
func (s *DriverService) AssignVehicle(ctx context.Context, driverID, vehicleID string) error {
	// Check if driver exists and is active
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return err
	}

	if driver.Status != "active" {
		return fmt.Errorf("driver is not active")
	}

	// Check if driver already has a vehicle assigned
	if driver.CurrentVehicleID != nil {
		return fmt.Errorf("driver already has a vehicle assigned")
	}

	if err := s.repo.AssignVehicle(ctx, driverID, vehicleID); err != nil {
		s.logger.Error("failed to assign vehicle to driver", 
			zap.String("driver_id", driverID), 
			zap.String("vehicle_id", vehicleID), 
			zap.Error(err))
		return fmt.Errorf("failed to assign vehicle: %w", err)
	}

	s.logger.Info("vehicle assigned to driver", 
		zap.String("driver_id", driverID), 
		zap.String("vehicle_id", vehicleID))
	return nil
}

// UnassignVehicle removes vehicle assignment from a driver
func (s *DriverService) UnassignVehicle(ctx context.Context, driverID string) error {
	// Check if driver exists
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return err
	}

	// Check if driver has a vehicle assigned
	if driver.CurrentVehicleID == nil {
		return fmt.Errorf("driver has no vehicle assigned")
	}

	if err := s.repo.UnassignVehicle(ctx, driverID); err != nil {
		s.logger.Error("failed to unassign vehicle from driver", 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		return fmt.Errorf("failed to unassign vehicle: %w", err)
	}

	s.logger.Info("vehicle unassigned from driver", zap.String("driver_id", driverID))
	return nil
}

// UpdateStatus updates driver status
func (s *DriverService) UpdateStatus(ctx context.Context, driverID string, status string) error {
	validStatuses := map[string]bool{
		"active": true, "inactive": true, "suspended": true, "resigned": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// If resigning, set resignation date
	if status == "resigned" {
		updates["resignation_date"] = time.Now()
		// Also unassign vehicle and trip if any
		driver, err := s.GetByID(ctx, driverID)
		if err == nil {
			if driver.CurrentVehicleID != nil {
				s.repo.UnassignVehicle(ctx, driverID)
			}
			if driver.CurrentTripID != nil {
				// Would need to integrate with trip service to properly end trip
				updates["current_trip_id"] = nil
			}
		}
	}

	if err := s.repo.Update(ctx, driverID, updates); err != nil {
		s.logger.Error("failed to update driver status", 
			zap.String("driver_id", driverID), 
			zap.String("status", status), 
			zap.Error(err))
		return fmt.Errorf("failed to update driver status: %w", err)
	}

	s.logger.Info("driver status updated", 
		zap.String("driver_id", driverID), 
		zap.String("status", status))
	return nil
}

// UpdateBehaviorScore updates driver's behavior score
func (s *DriverService) UpdateBehaviorScore(ctx context.Context, driverID string, scoreChange int, reason string) error {
	// Validate score change range
	if scoreChange < -20 || scoreChange > 20 {
		return fmt.Errorf("score change must be between -20 and 20")
	}

	// Check if driver exists
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return err
	}

	newScore := driver.BehaviorScore + scoreChange
	
	// Ensure score stays within bounds
	if newScore < 0 {
		newScore = 0
	} else if newScore > 100 {
		newScore = 100
	}

	scoreChangeActual := newScore - driver.BehaviorScore
	
	if err := s.repo.UpdateBehaviorScore(ctx, driverID, scoreChangeActual); err != nil {
		s.logger.Error("failed to update behavior score", 
			zap.String("driver_id", driverID), 
			zap.Int("score_change", scoreChange), 
			zap.String("reason", reason), 
			zap.Error(err))
		return fmt.Errorf("failed to update behavior score: %w", err)
	}

	s.logger.Info("driver behavior score updated", 
		zap.String("driver_id", driverID), 
		zap.Int("old_score", driver.BehaviorScore),
		zap.Int("new_score", newScore),
		zap.Int("score_change", scoreChangeActual),
		zap.String("reason", reason))
	
	return nil
}

// GetBehaviorScore retrieves driver's current behavior score
func (s *DriverService) GetBehaviorScore(ctx context.Context, driverID string) (int, error) {
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return 0, err
	}
	return driver.BehaviorScore, nil
}

// GetDriverAnalytics retrieves driver analytics
func (s *DriverService) GetDriverAnalytics(ctx context.Context, driverID string) (*models.DriverAnalytics, error) {
	analytics, err := s.repo.GetDriverAnalytics(ctx, driverID)
	if err != nil {
		s.logger.Error("failed to get driver analytics", 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to get driver analytics: %w", err)
	}
	return analytics, nil
}

// IncrementDriverStats increments driver statistics after trip completion
func (s *DriverService) IncrementDriverStats(ctx context.Context, driverID string, distanceKm, hoursDriven float64) error {
	if err := s.repo.IncrementStats(ctx, driverID, distanceKm, hoursDriven); err != nil {
		s.logger.Error("failed to increment driver stats", 
			zap.String("driver_id", driverID), 
			zap.Float64("distance_km", distanceKm),
			zap.Float64("hours_driven", hoursDriven),
			zap.Error(err))
		return fmt.Errorf("failed to increment driver stats: %w", err)
	}
	return nil
}

// GetByLicenseNumber retrieves driver by license number
func (s *DriverService) GetByLicenseNumber(ctx context.Context, licenseNumber string) (*models.Driver, error) {
	// This would require adding a method to the repository
	// For now, we'll implement a simple approach
	filter := models.DriverListFilter{
		Page: 1,
		PageSize: 1,
	}
	
	drivers, _, err := s.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// Filter by license number in memory (temporary solution)
	for _, driver := range drivers {
		if driver.LicenseNumber == licenseNumber {
			return driver, nil
		}
	}
	
	return nil, fmt.Errorf("driver with license number %s not found", licenseNumber)
}

// GetActiveDriversForVehicle retrieves active drivers for a vehicle
func (s *DriverService) GetActiveDriversForVehicle(ctx context.Context, vehicleID string) ([]*models.Driver, error) {
	// This would require adding a method to the repository
	// For now, return empty slice
	return []*models.Driver{}, nil
}
