package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FuelVendorService struct {
	repo   repository.FuelLogRepository
	logger *zap.Logger
}

func NewFuelVendorService(repo repository.FuelLogRepository, logger *zap.Logger) *FuelVendorService {
	return &FuelVendorService{repo: repo, logger: logger}
}

func (s *FuelVendorService) CreateVendor(ctx context.Context, orgID uuid.UUID, req models.CreateFuelVendorRequest) (*models.FuelVendor, error) {
	vendor := &models.FuelVendor{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           req.Name,
		GSTIN:          req.GSTIN,
		Address:        req.Address,
		ContactPhone:   req.ContactPhone,
		Email:          req.Email,
		Status:         "ACTIVE",
	}

	if err := s.repo.CreateVendor(ctx, vendor); err != nil {
		return nil, err
	}

	return vendor, nil
}

func (s *FuelVendorService) ListVendors(ctx context.Context, orgID uuid.UUID) ([]models.FuelVendor, error) {
	return s.repo.ListVendors(ctx, orgID)
}
