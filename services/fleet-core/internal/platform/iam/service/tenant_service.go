package service

import (
	"context"
	"github.com/flyfleet/fleet-core/internal/platform/iam/models"
	"github.com/flyfleet/fleet-core/internal/platform/iam/repository"
	"github.com/google/uuid"
	"time"
)

type TenantService interface {
	Onboard(ctx context.Context, name, taxID string) (*models.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
}

type tenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) Onboard(ctx context.Context, name, taxID string) (*models.Tenant, error) {
	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      name,
		TaxID:     taxID,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *tenantService) GetTenant(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}
