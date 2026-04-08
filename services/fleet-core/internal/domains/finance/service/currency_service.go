package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
)

type CurrencyService struct {
	repo *repository.FinanceRepository
}

func NewCurrencyService(repo *repository.FinanceRepository) *CurrencyService {
	return &CurrencyService{repo: repo}
}

func (s *CurrencyService) Convert(ctx context.Context, amount float64, from, to string) (float64, error) {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	if from == to {
		return amount, nil
	}

	rate, err := s.repo.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, fmt.Errorf("exchange rate not found for %s to %s: %w", from, to, err)
	}

	return amount * rate, nil
}

func (s *CurrencyService) GetRate(ctx context.Context, from, to string) (float64, error) {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)
	if from == to {
		return 1.0, nil
	}
	return s.repo.GetExchangeRate(ctx, from, to)
}
