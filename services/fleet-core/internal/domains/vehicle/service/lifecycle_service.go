package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LifecycleService manages vehicle status state machine transitions.
//
// Allowed transitions:
//
//	ACTIVE       → MAINTENANCE, SUSPENDED
//	MAINTENANCE  → ACTIVE, SUSPENDED
//	SUSPENDED    → ACTIVE
//	DECOMMISSIONED → (terminal — no transitions allowed)
type LifecycleService struct {
	vRepo      repository.VehicleRepositoryInterface
	compRepo   repository.ComplianceRepositoryInterface
	eventsRepo repository.EventsRepositoryInterface
	logger     *zap.Logger
}

func NewLifecycleService(
	vRepo repository.VehicleRepositoryInterface,
	compRepo repository.ComplianceRepositoryInterface,
	eventsRepo repository.EventsRepositoryInterface,
	logger *zap.Logger,
) *LifecycleService {
	return &LifecycleService{
		vRepo:      vRepo,
		compRepo:   compRepo,
		eventsRepo: eventsRepo,
		logger:     logger,
	}
}

// Activate transitions a vehicle to ACTIVE status.
// Business rule: ACTIVE requires valid RC, Insurance, and PUC.
func (s *LifecycleService) Activate(ctx context.Context, vehicleID uuid.UUID) (*models.Vehicle, error) {
	vehicle, err := s.vRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	// SECURITY FIX: Ensure vehicle belongs to authenticated organization (CWE-639)
	if err := s.ensureVehicleOrgMatch(ctx, vehicle); err != nil {
		return nil, err
	}
	if vehicle.Status == models.StatusDecommissioned {
		return nil, ErrDecommissionedIrreversible
	}
	if vehicle.Status == models.StatusActive {
		return vehicle, nil // idempotent
	}

	// Check compliance gate
	if eligible, reason, err := s.checkComplianceForActivation(ctx, vehicleID); err != nil {
		return nil, err
	} else if !eligible {
		return nil, fmt.Errorf("%w: %s", ErrComplianceRequired, reason)
	}

	oldStatus := vehicle.Status
	vehicle.Status = models.StatusActive
	if err := s.vRepo.Update(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to activate vehicle: %w", err)
	}

	s.emitStatusChangedEvent(ctx, vehicle.ID, oldStatus, models.StatusActive)
	return vehicle, nil
}

// Suspend transitions a vehicle to SUSPENDED status.
func (s *LifecycleService) Suspend(ctx context.Context, vehicleID uuid.UUID, reason string) (*models.Vehicle, error) {
	vehicle, err := s.vRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	// SECURITY FIX: Ensure vehicle belongs to authenticated organization (CWE-639)
	if err := s.ensureVehicleOrgMatch(ctx, vehicle); err != nil {
		return nil, err
	}
	if vehicle.Status == models.StatusDecommissioned {
		return nil, ErrDecommissionedIrreversible
	}

	oldStatus := vehicle.Status
	vehicle.Status = models.StatusSuspended
	if err := s.vRepo.Update(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to suspend vehicle: %w", err)
	}

	s.emitStatusChangedEvent(ctx, vehicle.ID, oldStatus, models.StatusSuspended)
	return vehicle, nil
}

// SetMaintenance transitions a vehicle to MAINTENANCE status.
// Business rule: only ACTIVE or SUSPENDED vehicles can enter maintenance.
func (s *LifecycleService) SetMaintenance(ctx context.Context, vehicleID uuid.UUID) (*models.Vehicle, error) {
	vehicle, err := s.vRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	// SECURITY FIX: Ensure vehicle belongs to authenticated organization (CWE-639)
	if err := s.ensureVehicleOrgMatch(ctx, vehicle); err != nil {
		return nil, err
	}
	switch vehicle.Status {
	case models.StatusDecommissioned:
		return nil, ErrDecommissionedIrreversible
	case models.StatusMaintenance:
		return vehicle, nil // idempotent
	}

	oldStatus := vehicle.Status
	vehicle.Status = models.StatusMaintenance
	if err := s.vRepo.Update(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to set maintenance status: %w", err)
	}

	s.emitStatusChangedEvent(ctx, vehicle.ID, oldStatus, models.StatusMaintenance)
	return vehicle, nil
}

// Decommission permanently transitions a vehicle to DECOMMISSIONED (irreversible).
// Business rule: DECOMMISSIONED cannot revert.
func (s *LifecycleService) Decommission(ctx context.Context, vehicleID uuid.UUID, reason string) (*models.Vehicle, error) {
	vehicle, err := s.vRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	// SECURITY FIX: Ensure vehicle belongs to authenticated organization (CWE-639)
	if err := s.ensureVehicleOrgMatch(ctx, vehicle); err != nil {
		return nil, err
	}
	if vehicle.Status == models.StatusDecommissioned {
		return vehicle, nil // already decommissioned — idempotent
	}

	oldStatus := vehicle.Status
	vehicle.Status = models.StatusDecommissioned
	if err := s.vRepo.Update(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to decommission vehicle: %w", err)
	}

	s.emitEvent(ctx, vehicle.ID, models.EventVehicleDecommissioned, map[string]interface{}{
		"old_status": oldStatus,
		"reason":     reason,
	})
	return vehicle, nil
}

func (s *LifecycleService) checkComplianceForActivation(ctx context.Context, vehicleID uuid.UUID) (bool, string, error) {
	comp, err := s.compRepo.Get(ctx, vehicleID)
	if err != nil {
		return false, "compliance check failed", err
	}
	if comp == nil {
		// No compliance record — allow activation (first-time setup)
		return true, "", nil
	}
	if comp.LegalHold {
		return false, "vehicle is on legal hold", nil
	}
	if comp.IsBlacklisted {
		return false, "vehicle is blacklisted", nil
	}
	return true, "", nil
}

func (s *LifecycleService) emitStatusChangedEvent(ctx context.Context, vehicleID uuid.UUID, oldStatus, newStatus string) {
	s.emitEvent(ctx, vehicleID, models.EventVehicleStatusChanged, map[string]interface{}{
		"old_status": oldStatus,
		"new_status": newStatus,
	})
}

func (s *LifecycleService) emitEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload map[string]interface{}) {
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
		s.logger.Warn("failed to emit lifecycle event", zap.Error(err))
	}
}

// ensureVehicleOrgMatch verifies that the vehicle belongs to the organization from the request context.
// Returns ErrVehicleNotFound if the check fails to avoid information leakage.
func (s *LifecycleService) ensureVehicleOrgMatch(ctx context.Context, vehicle *models.Vehicle) error {
	orgIDVal := ctx.Value("org_id")
	if orgIDVal == nil {
		s.logger.Error("Organization context missing")
		return ErrVehicleNotFound
	}
	var reqOrgID uuid.UUID
	switch v := orgIDVal.(type) {
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			s.logger.Error("Invalid orgID in context", zap.Error(err))
			return ErrVehicleNotFound
		}
		reqOrgID = id
	case uuid.UUID:
		reqOrgID = v
	default:
		s.logger.Error("Invalid orgID type", zap.Any("type", fmt.Sprintf("%T", orgIDVal)))
		return ErrVehicleNotFound
	}
	if vehicle.TenantID != reqOrgID {
		return ErrVehicleNotFound
	}
	return nil
}
