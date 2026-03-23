package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/crypto"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VehicleServiceInterface defines the interface for vehicle CRUD and management.
type VehicleServiceInterface interface {
	// CRUD
	CreateVehicle(ctx context.Context, req *models.CreateVehicleRequest, orgID uuid.UUID) (*models.Vehicle, error)
	GetVehicle(ctx context.Context, id uuid.UUID) (*models.Vehicle, error)
	ListVehicles(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.Vehicle, int64, error)
	UpdateVehicle(ctx context.Context, id uuid.UUID, req *models.UpdateVehicleRequest) (*models.Vehicle, error)
	DeleteVehicle(ctx context.Context, id uuid.UUID) error

	// Driver assignment
	AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) (*models.Vehicle, error)

	// Location (stub — real data from tracking service)
	GetVehicleLocation(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleLocation, error)

	// Statistics
	GetStats(ctx context.Context) (*models.VehicleStats, error)

	// Search
	SearchVehicles(ctx context.Context, params models.VehicleSearchParams) ([]models.Vehicle, int64, error)
}

type VehicleService struct {
	repo       repository.VehicleRepositoryInterface
	eventsRepo repository.EventsRepositoryInterface
	logger     *zap.Logger
}

func NewVehicleService(repo repository.VehicleRepositoryInterface, logger *zap.Logger) *VehicleService {
	return &VehicleService{repo: repo, logger: logger}
}

// NewVehicleServiceWithEvents creates a VehicleService that also records events.
func NewVehicleServiceWithEvents(
	repo repository.VehicleRepositoryInterface,
	eventsRepo repository.EventsRepositoryInterface,
	logger *zap.Logger,
) *VehicleService {
	return &VehicleService{repo: repo, eventsRepo: eventsRepo, logger: logger}
}

func (s *VehicleService) CreateVehicle(ctx context.Context, req *models.CreateVehicleRequest, orgID uuid.UUID) (*models.Vehicle, error) {
	// Normalize legacy registration_number
	regNo := req.RegistrationNumber
	if regNo == "" && req.LicensePlate != nil {
		regNo = *req.LicensePlate
	}
	if regNo == "" {
		return nil, ErrMissingRequiredField
	}

	// Business rule: unique registration per tenant
	existing, err := s.repo.GetByRegistrationNumber(ctx, orgID, regNo)
	if err != nil {
		return nil, fmt.Errorf("failed to check registration: %w", err)
	}
	if existing != nil {
		return nil, ErrVehicleAlreadyExists
	}

	vehicle := &models.Vehicle{
		ID:                 uuid.New(),
		TenantID:           orgID,
		RegistrationNumber: regNo,
		ChassisNumber:      req.ChassisNumber,
		EngineNumber:       req.EngineNumber,
		VINNumber:          req.VINNumber,
		VehicleType:        req.VehicleType,
		CapacityTonnage:    req.CapacityTonnage,
		FuelType:           req.FuelType,
		FuelTankCapacity:   req.FuelTankCapacity,
		OwnershipType:      req.OwnershipType,
		ParentVehicleID:    req.ParentVehicleID,
		BranchID:           req.BranchID,
		CostCenterID:       req.CostCenterID,
		GPSDeviceID:        req.GPSDeviceID,
		ImmobilizerEnabled: req.ImmobilizerEnabled,
		AIS140Compliant:    req.AIS140Compliant,
		Make:               req.Make,
		Model:              req.Model,
		Year:               req.Year,
		Color:              req.Color,
		Status:             models.StatusActive,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Encrypt sensitive fields before persist
	if err := crypto.EncryptPtr(vehicle.ChassisNumber); err != nil {
		return nil, fmt.Errorf("failed to encrypt chassis number: %w", err)
	}
	if err := crypto.EncryptPtr(vehicle.EngineNumber); err != nil {
		return nil, fmt.Errorf("failed to encrypt engine number: %w", err)
	}

	if err := s.repo.Create(ctx, vehicle); err != nil {
		s.logger.Error("failed to create vehicle", zap.Error(err))
		return nil, fmt.Errorf("failed to create vehicle: %w", err)
	}

	// Emit creation event — best-effort (non-transactional outbox)
	s.emitEvent(ctx, vehicle.ID, models.EventVehicleCreated, map[string]interface{}{
		"registration_number": vehicle.RegistrationNumber,
		"tenant_id":           vehicle.TenantID.String(),
		"status":              vehicle.Status,
	})

	// Decrypt for response (caller should not see ciphertext)
	_ = crypto.DecryptPtr(vehicle.ChassisNumber)
	_ = crypto.DecryptPtr(vehicle.EngineNumber)

	return vehicle, nil
}

func (s *VehicleService) GetVehicle(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Decrypt sensitive fields for caller
	_ = crypto.DecryptPtr(v.ChassisNumber)
	_ = crypto.DecryptPtr(v.EngineNumber)
	return v, nil
}

func (s *VehicleService) ListVehicles(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.Vehicle, int64, error) {
	offset := (page - 1) * limit
	vehicles, total, err := s.repo.GetByOrgID(ctx, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]models.Vehicle, len(vehicles))
	for i, v := range vehicles {
		result[i] = *v
	}
	return result, total, nil
}

func (s *VehicleService) UpdateVehicle(ctx context.Context, id uuid.UUID, req *models.UpdateVehicleRequest) (*models.Vehicle, error) {
	vehicle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Business rule: legal hold prevents modification
	// (checked by caller via compliance service; here we just update)

	// Business rule: DECOMMISSIONED cannot be updated
	if vehicle.Status == models.StatusDecommissioned {
		return nil, ErrVehicleCannotBeUpdated
	}

	if req.VehicleType != nil {
		vehicle.VehicleType = *req.VehicleType
	}
	if req.FuelType != nil {
		vehicle.FuelType = *req.FuelType
	}
	if req.OwnershipType != nil {
		vehicle.OwnershipType = *req.OwnershipType
	}
	if req.CapacityTonnage != nil {
		vehicle.CapacityTonnage = req.CapacityTonnage
	}
	if req.FuelTankCapacity != nil {
		vehicle.FuelTankCapacity = req.FuelTankCapacity
	}
	if req.ChassisNumber != nil {
		vehicle.ChassisNumber = req.ChassisNumber
	}
	if req.EngineNumber != nil {
		vehicle.EngineNumber = req.EngineNumber
	}
	if req.VINNumber != nil {
		vehicle.VINNumber = req.VINNumber
	}
	if req.BranchID != nil {
		vehicle.BranchID = req.BranchID
	}
	if req.CostCenterID != nil {
		vehicle.CostCenterID = req.CostCenterID
	}
	if req.GPSDeviceID != nil {
		vehicle.GPSDeviceID = req.GPSDeviceID
	}
	if req.ImmobilizerEnabled != nil {
		vehicle.ImmobilizerEnabled = *req.ImmobilizerEnabled
	}
	if req.AIS140Compliant != nil {
		vehicle.AIS140Compliant = *req.AIS140Compliant
	}
	if req.Make != nil {
		vehicle.Make = req.Make
	}
	if req.Model != nil {
		vehicle.Model = req.Model
	}
	if req.Year != nil {
		vehicle.Year = req.Year
	}
	if req.Color != nil {
		vehicle.Color = req.Color
	}
	// Legacy status update (normalize case)
	if req.Status != nil {
		s := normalizeStatus(*req.Status)
		vehicle.Status = s
	}

	if err := s.repo.Update(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to update vehicle: %w", err)
	}

	s.emitEvent(ctx, vehicle.ID, models.EventVehicleUpdated, map[string]interface{}{
		"status": vehicle.Status,
	})

	return vehicle, nil
}

func (s *VehicleService) DeleteVehicle(ctx context.Context, id uuid.UUID) error {
	vehicle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vehicle.Status == models.StatusDecommissioned {
		return ErrVehicleCannotBeDeleted
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}
	return nil
}

func (s *VehicleService) AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) (*models.Vehicle, error) {
	if err := s.repo.AssignDriver(ctx, vehicleID, driverID); err != nil {
		return nil, fmt.Errorf("failed to assign driver: %w", err)
	}
	return s.repo.GetByID(ctx, vehicleID)
}

func (s *VehicleService) GetVehicleLocation(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleLocation, error) {
	vehicle, err := s.repo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	return &models.VehicleLocation{
		ID:        uuid.New(),
		VehicleID: vehicleID,
		Timestamp: vehicle.UpdatedAt,
	}, nil
}

func (s *VehicleService) GetStats(ctx context.Context) (*models.VehicleStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *VehicleService) SearchVehicles(ctx context.Context, params models.VehicleSearchParams) ([]models.Vehicle, int64, error) {
	vehicles, total, err := s.repo.Search(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	result := make([]models.Vehicle, len(vehicles))
	for i, v := range vehicles {
		result[i] = *v
	}
	return result, total, nil
}

// emitEvent is best-effort; errors are logged but do not fail the main operation.
func (s *VehicleService) emitEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload map[string]interface{}) {
	if s.eventsRepo == nil {
		return
	}
	event := &models.VehicleEvent{
		ID:        uuid.New(),
		VehicleID: vehicleID,
		EventType: eventType,
		Payload:   payload,
	}
	if err := s.eventsRepo.Append(ctx, event); err != nil {
		s.logger.Warn("failed to append vehicle event", zap.String("event_type", eventType), zap.Error(err))
	}
}

// normalizeStatus converts legacy lowercase status values to ENUM values.
func normalizeStatus(s string) string {
	switch s {
	case "active":
		return models.StatusActive
	case "maintenance":
		return models.StatusMaintenance
	case "inactive", "suspended":
		return models.StatusSuspended
	case "retired", "decommissioned":
		return models.StatusDecommissioned
	}
	return s
}
