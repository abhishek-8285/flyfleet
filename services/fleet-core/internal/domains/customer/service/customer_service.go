package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
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
	if c.CompanyName == "" {
		return fmt.Errorf("company name is required")
	}
	return s.repo.CreateCustomer(ctx, c)
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	return s.repo.GetCustomerByID(ctx, id)
}
