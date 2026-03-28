package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/india/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GSTIntegrationService defines the interface for GST portal integration
type GSTIntegrationService interface {
	VerifyGSTIN(ctx context.Context, orgID uuid.UUID, gstin string) (*models.GSTCompliance, error)
	SyncCompliance(ctx context.Context) error
}

type GSTIntegrationServiceImpl struct {
	repo   repository.IndiaIntegrationRepository
	logger *zap.Logger
}

func NewGSTIntegrationService(repo repository.IndiaIntegrationRepository, logger *zap.Logger) *GSTIntegrationServiceImpl {
	return &GSTIntegrationServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *GSTIntegrationServiceImpl) VerifyGSTIN(ctx context.Context, orgID uuid.UUID, gstin string) (*models.GSTCompliance, error) {
	s.logger.Info("verifying GSTIN", zap.String("gstin", gstin))

	// Mock GST Portal API Call
	compliance := &models.GSTCompliance{
		ID:                 uuid.New().String(),
		OrganizationID:     orgID,
		GSTIN:              gstin,
		Status:             "active",
		LegalName:          "Mock Legal Name",
		Address:            "Mock Address",
		DateOfRegistration: time.Now().Add(-365 * 24 * time.Hour),
		LastUpdated:        time.Now(),
		ReturnFilingStatus: "compliant",
		PaymentStatus:      "compliant",
		CreatedAt:          time.Now(),
	}

	log := &models.IntegrationLog{
		ServiceType:     "gst",
		RequestPayload:  map[string]string{"gstin": gstin},
		ResponsePayload: compliance,
		Status:          200,
		OrganizationID:  &orgID,
	}
	_ = s.repo.CreateIntegrationLog(ctx, log)

	if err := s.repo.CreateGSTCompliance(ctx, compliance); err != nil {
		return nil, err
	}

	return compliance, nil
}

func (s *GSTIntegrationServiceImpl) SyncCompliance(ctx context.Context) error {
	s.logger.Info("starting GST compliance sync")
	// Logic to poll GST portal for compliance status of all registered organizations
	return nil
}
