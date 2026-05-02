package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/models"
	"github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/repository"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TenantService interface {
	GetTenants(ctx context.Context, page, limit int, search string) ([]models.Tenant, int, error)
	GetTenantStats(ctx context.Context) (models.TenantStats, error)
	GetDashboardStats(ctx context.Context) (models.DashboardStats, error)
	CreateTenant(ctx context.Context, req models.CreateTenantRequest) (models.CreateTenantResponse, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	ListBranches(ctx context.Context) ([]models.Branch, error)
	ListCostCenters(ctx context.Context) ([]models.CostCenter, error)
}

type tenantService struct {
	repo   *repository.TenantRepository
	logger *zap.Logger
}

func NewTenantService(repo *repository.TenantRepository, logger *zap.Logger) TenantService {
	return &tenantService{
		repo:   repo,
		logger: logger,
	}
}

func (s *tenantService) GetTenants(ctx context.Context, page, limit int, search string) ([]models.Tenant, int, error) {
	return s.repo.ListTenants(ctx, page, limit, search)
}

func (s *tenantService) GetTenantStats(ctx context.Context) (models.TenantStats, error) {
	return s.repo.GetTenantStats(ctx)
}

func (s *tenantService) GetDashboardStats(ctx context.Context) (models.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}

func (s *tenantService) CreateTenant(ctx context.Context, req models.CreateTenantRequest) (models.CreateTenantResponse, error) {
	return s.repo.CreateTenant(ctx, req, s.getUserID(ctx))
}

func (s *tenantService) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteTenant(ctx, id, s.getUserID(ctx))
}

func (s *tenantService) ListBranches(ctx context.Context) ([]models.Branch, error) {
	return s.repo.ListBranches(ctx, s.getTenantIDFromContext(ctx))
}

func (s *tenantService) ListCostCenters(ctx context.Context) ([]models.CostCenter, error) {
	return s.repo.ListCostCenters(ctx, s.getTenantIDFromContext(ctx))
}

func (s *tenantService) getUserID(ctx context.Context) uuid.UUID {
	id, _ := ctxkeys.GetUserID(ctx)
	return id
}

func (s *tenantService) getTenantIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctxkeys.GetTenantID(ctx)
	return id
}
