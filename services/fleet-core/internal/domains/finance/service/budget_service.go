package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BudgetService struct {
	repo            *repository.FinanceRepository
	currencyService *CurrencyService
	alertingService *AlertingService
	logger          *zap.Logger
}

func NewBudgetService(repo *repository.FinanceRepository, currencyService *CurrencyService, alertingService *AlertingService, logger *zap.Logger) *BudgetService {
	return &BudgetService{
		repo:            repo,
		currencyService: currencyService,
		alertingService: alertingService,
		logger:          logger,
	}
}

// TrackExpense updates the budget for a given category and org
func (s *BudgetService) TrackExpense(ctx context.Context, tenantID, category string, amount float64, currency string) error {
	budget, err := s.repo.GetBudget(ctx, tenantID, category, time.Now())
	if err != nil {
		return err
	}
	if budget == nil {
		return nil // No budget set for this category
	}

	// Convert amount if currency doesn't match
	convertedAmount := amount
	if currency != budget.Currency {
		if s.currencyService == nil {
			return nil
		}
		convertedAmount, err = s.currencyService.Convert(ctx, amount, currency, budget.Currency)
		if err != nil {
			s.logger.Error("failed to convert currency for budget tracking", zap.Error(err))
			// Continue with original amount as fallback or return error?
			// For fleet safety, we'll return error to ensure accuracy
			return err
		}
	}

	if err := s.repo.UpdateBudgetSpent(ctx, budget.ID, convertedAmount); err != nil {
		return err
	}

	// Check threshold
	newSpent := budget.SpentAmount + convertedAmount
	usagePct := (newSpent / budget.LimitAmount) * 100
	if usagePct >= float64(budget.AlertThresholdPct) {
		s.logger.Warn("Budget threshold reached",
			zap.String("tenant_id", tenantID),
			zap.String("category", category),
			zap.Float64("usage_pct", usagePct),
			zap.Float64("limit", budget.LimitAmount),
			zap.Float64("spent", newSpent))
		orgUUID, parseErr := uuid.Parse(tenantID)
		if parseErr == nil && s.alertingService != nil {
			_ = s.alertingService.Publish(ctx, eventBudgetLimitExceeded, orgUUID, map[string]interface{}{
				"category":        category,
				"usage_pct":       usagePct,
				"limit_amount":    budget.LimitAmount,
				"spent_amount":    newSpent,
				"threshold_pct":   budget.AlertThresholdPct,
				"period_start":    budget.PeriodStart,
				"period_end":      budget.PeriodEnd,
				"budget_currency": budget.Currency,
				"source_currency": currency,
			})
		}
	}

	return nil
}
