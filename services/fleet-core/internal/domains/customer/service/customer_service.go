package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core/internal/domains/customer/repository"
	"github.com/google/uuid"
)

type CustomerService struct {
	repo repository.CustomerV2Repository
}

func NewCustomerService(repo repository.CustomerV2Repository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, c *models.Customer) error {
	// Add business logic validation here
	if c.Name == "" {
		return fmt.Errorf("company name is required")
	}
	return s.repo.CreateCustomer(ctx, c)
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	return s.repo.GetCustomerByID(ctx, id)
}

func (s *CustomerService) ListCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]models.Customer, error) {
	return s.repo.ListCustomers(ctx, tenantID, filters, limit, offset)
}

func (s *CustomerService) CountCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}) (int, error) {
	return s.repo.CountCustomers(ctx, tenantID, filters)
}

func (s *CustomerService) GetCustomerSummary(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	return s.repo.GetDashboardMetrics(ctx, customerID)
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, c *models.Customer) error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("customer id is required for updates")
	}
	return s.repo.UpdateCustomer(ctx, c)
}
