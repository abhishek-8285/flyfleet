package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/models"
	"github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/repository"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrganizationService interface {
	GetOrganizations(ctx context.Context, page, limit int, search string) ([]models.Organization, int, error)
	GetOrganizationStats(ctx context.Context) (models.OrganizationStats, error)
	GetDashboardStats(ctx context.Context) (models.DashboardStats, error)
	CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest) (models.CreateOrganizationResponse, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	ListBranches(ctx context.Context) ([]models.Branch, error)
	ListCostCenters(ctx context.Context) ([]models.CostCenter, error)
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
	return s.repo.CreateOrganization(ctx, req, s.getUserID(ctx))
}

func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteOrganization(ctx, id, s.getUserID(ctx))
}

func (s *organizationService) ListBranches(ctx context.Context) ([]models.Branch, error) {
	return s.repo.ListBranches(ctx, s.getTenantIDFromContext(ctx))
}

func (s *organizationService) ListCostCenters(ctx context.Context) ([]models.CostCenter, error) {
	return s.repo.ListCostCenters(ctx, s.getTenantIDFromContext(ctx))
}

func (s *organizationService) getUserID(ctx context.Context) uuid.UUID {
	val, ok := ctx.Value(ctxkeys.UserIDKey).(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(val)
	return uid
}

func (s *organizationService) getTenantIDFromContext(ctx context.Context) uuid.UUID {
	if val := ctx.Value(ctxkeys.TenantIDKey); val != nil {
		switch v := val.(type) {
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		case uuid.UUID:
			return v
		}
	}
	return uuid.Nil
}
