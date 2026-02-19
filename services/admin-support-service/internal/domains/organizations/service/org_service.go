package service

import (
	"context"

	"github.com/flyfleet/admin-support-service/internal/domains/organizations/models"
	"github.com/flyfleet/admin-support-service/internal/domains/organizations/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrganizationService interface {
	GetOrganizations(ctx context.Context, page, limit int, search string) ([]models.Organization, int, error)
	GetOrganizationStats(ctx context.Context) (models.OrganizationStats, error)
	GetDashboardStats(ctx context.Context) (models.DashboardStats, error)
	CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest) (models.CreateOrganizationResponse, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
}

type organizationService struct {
	repo   *repository.OrganizationRepository
	logger *zap.Logger
}

func NewOrganizationService(repo *repository.OrganizationRepository, logger *zap.Logger) OrganizationService {
	return &organizationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *organizationService) GetOrganizations(ctx context.Context, page, limit int, search string) ([]models.Organization, int, error) {
	return s.repo.ListOrganizations(ctx, page, limit, search)
}

func (s *organizationService) GetOrganizationStats(ctx context.Context) (models.OrganizationStats, error) {
	return s.repo.GetOrganizationStats(ctx)
}

func (s *organizationService) GetDashboardStats(ctx context.Context) (models.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}

func (s *organizationService) CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest) (models.CreateOrganizationResponse, error) {
	return s.repo.CreateOrganization(ctx, req)
}

func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteOrganization(ctx, id)
}
