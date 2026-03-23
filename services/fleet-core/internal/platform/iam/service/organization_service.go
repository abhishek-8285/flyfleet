package service

import (
	"context"
	"time"
	"github.com/flyfleet/fleet-core-service/internal/platform/iam/models"
	"github.com/flyfleet/fleet-core-service/internal/platform/iam/repository"
	"github.com/google/uuid"
)

type OrganizationService interface {
	Onboard(ctx context.Context, name, taxID string) (*models.Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error)
}

type organizationService struct {
	repo repository.OrganizationRepository
}

func NewOrganizationService(repo repository.OrganizationRepository) OrganizationService {
	return &organizationService{repo: repo}
}

func (s *organizationService) Onboard(ctx context.Context, name, taxID string) (*models.Organization, error) {
	org := &models.Organization{
		ID:        uuid.New(),
		Name:      name,
		TaxID:     taxID,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *organizationService) GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	return s.repo.GetByID(ctx, id)
}
