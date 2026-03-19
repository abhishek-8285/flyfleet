package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
)

type CreditService struct {
	repo repository.CustomerV2Repository
}

func NewCreditService(repo repository.CustomerV2Repository) *CreditService {
	return &CreditService{repo: repo}
}

// CheckCreditAvailability checks if a customer has enough credit for a booking
func (s *CreditService) CheckCreditAvailability(ctx context.Context, customerID string, amount float64) (bool, error) {
	// Logic to query wallet, credit limit, and evaluate rules
	return true, nil
}
