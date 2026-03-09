package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ComplianceService enforces India-specific compliance rules.
type ComplianceService struct {
	vRepo      repository.VehicleRepositoryInterface
	compRepo   repository.ComplianceRepositoryInterface
	eventsRepo repository.EventsRepositoryInterface
	logger     *zap.Logger
}

func NewComplianceService(
	vRepo repository.VehicleRepositoryInterface,
	compRepo repository.ComplianceRepositoryInterface,
	eventsRepo repository.EventsRepositoryInterface,
	logger *zap.Logger,
) *ComplianceService {
	return &ComplianceService{
		vRepo:      vRepo,
		compRepo:   compRepo,
		eventsRepo: eventsRepo,
		logger:     logger,
	}
}

// VerifyCompliance upserts compliance records and derives compliance status.
func (s *ComplianceService) VerifyCompliance(ctx context.Context, vehicleID uuid.UUID, req *models.VerifyComplianceRequest) (*models.VehicleCompliance, *models.ComplianceStatus, error) {
	// Ensure vehicle exists
	if _, err := s.vRepo.GetByID(ctx, vehicleID); err != nil {
		return nil, nil, err
	}

	now := time.Now()
	compliance := &models.VehicleCompliance{
		VehicleID:           vehicleID,
		RCExpiryDate:        req.RCExpiryDate,
		InsuranceExpiryDate: req.InsuranceExpiryDate,
		PUCExpiryDate:       req.PUCExpiryDate,
		FitnessExpiryDate:   req.FitnessExpiryDate,
		PermitExpiryDate:    req.PermitExpiryDate,
		PermitType:          req.PermitType,
		InsurancePolicyNo:   req.InsurancePolicyNo,
		LastVerifiedAt:      &now,
	}

	if err := s.compRepo.Upsert(ctx, compliance); err != nil {
		s.logger.Error("failed to upsert compliance", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to save compliance: %w", err)
	}

	status := s.deriveStatus(compliance)

	// Emit compliance expiry event if any document is expired
	if len(status.ExpiredDocuments) > 0 {
		s.emitEvent(ctx, vehicleID, models.EventVehicleComplianceExpired, map[string]interface{}{
			"expired_documents": status.ExpiredDocuments,
		})
	}

	return compliance, status, nil
}

// GetCompliance returns the compliance record for a vehicle.
func (s *ComplianceService) GetCompliance(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleCompliance, *models.ComplianceStatus, error) {
	comp, err := s.compRepo.Get(ctx, vehicleID)
	if err != nil {
		return nil, nil, err
	}
	if comp == nil {
		return nil, &models.ComplianceStatus{IsCompliant: false}, nil
	}
	return comp, s.deriveStatus(comp), nil
}

// GetExpiringVehicles returns vehicles with compliance expiring within `days`.
func (s *ComplianceService) GetExpiringVehicles(ctx context.Context, tenantID uuid.UUID, days int) ([]models.ComplianceExpiryAlert, error) {
	if days <= 0 {
		days = 30
	}
	return s.compRepo.GetExpiringVehicles(ctx, tenantID, days)
}

// IsEligibleForTrip returns true when the vehicle has no expired compliance and is not on legal hold.
func (s *ComplianceService) IsEligibleForTrip(ctx context.Context, vehicleID uuid.UUID) (bool, string, error) {
	comp, err := s.compRepo.Get(ctx, vehicleID)
	if err != nil {
		return false, "compliance check failed", err
	}
	if comp == nil {
		return false, "no compliance record found", nil
	}
	if comp.LegalHold {
		return false, "vehicle is on legal hold", nil
	}
	if comp.IsBlacklisted {
		return false, "vehicle is blacklisted", nil
	}
	status := s.deriveStatus(comp)
	if !status.IsCompliant {
		return false, fmt.Sprintf("expired documents: %v", status.ExpiredDocuments), nil
	}
	return true, "", nil
}

func (s *ComplianceService) deriveStatus(comp *models.VehicleCompliance) *models.ComplianceStatus {
	now := time.Now()
	soonThreshold := now.AddDate(0, 0, 30)

	status := &models.ComplianceStatus{
		IsCompliant:   true,
		IsBlacklisted: comp.IsBlacklisted,
		LegalHold:     comp.LegalHold,
	}

	if comp.IsBlacklisted || comp.LegalHold {
		status.IsCompliant = false
	}

	checkDate := func(date *time.Time, label string) {
		if date == nil {
			return
		}
		if date.Before(now) {
			status.ExpiredDocuments = append(status.ExpiredDocuments, label)
			status.IsCompliant = false
		} else if date.Before(soonThreshold) {
			status.ExpiringDocuments = append(status.ExpiringDocuments, label)
		}
	}

	checkDate(comp.RCExpiryDate, "RC")
	checkDate(comp.InsuranceExpiryDate, "INSURANCE")
	checkDate(comp.PUCExpiryDate, "PUC")
	checkDate(comp.FitnessExpiryDate, "FITNESS")
	checkDate(comp.PermitExpiryDate, "PERMIT")

	return status
}

func (s *ComplianceService) emitEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload map[string]interface{}) {
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
		s.logger.Warn("failed to emit compliance event", zap.Error(err))
	}
}
