package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdvancedSubService struct {
	repo       *repository.AdvancedSubRepository
	ledgerServ *LedgerService
	subServ    *SubscriptionService
	billServ   *BillingService
	logger     *zap.Logger
}

func NewAdvancedSubService(repo *repository.AdvancedSubRepository, ledgerServ *LedgerService, subServ *SubscriptionService, billServ *BillingService, logger *zap.Logger) *AdvancedSubService {
	return &AdvancedSubService{
		repo:       repo,
		ledgerServ: ledgerServ,
		subServ:    subServ,
		billServ:   billServ,
		logger:     logger,
	}
}

// UpgradePlan applies a mid-cycle plan change, calculating proration credits and debits.
func (s *AdvancedSubService) UpgradePlan(ctx context.Context, tenantID uuid.UUID, newPlanID uuid.UUID) error {
	// 1. Get current subscription
	sub, err := s.subServ.GetSubscription(ctx, tenantID)
	if err != nil || sub == nil {
		return errors.New("active subscription not found")
	}

	// 2. Get active billing cycle
	cycle, err := s.repo.GetActiveCycle(ctx, sub.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch active cycle: %w", err)
	}
	if cycle == nil {
		return errors.New("no active billing cycle found to prorate against")
	}

	// 3. Get old plan and new plan details
	plans, err := s.subServ.ListPlans(ctx)
	if err != nil {
		return err
	}

	var oldPlan, newPlan *models.Plan
	for i := range plans {
		if plans[i].ID == sub.PlanID {
			oldPlan = &plans[i]
		}
		if plans[i].ID == newPlanID {
			newPlan = &plans[i]
		}
	}
	if oldPlan == nil || newPlan == nil {
		return errors.New("invalid plan IDs")
	}

	// Calculate base prices
	var oldPlanBasePrice, newPlanBasePrice float64
	var oldPlanCurrency, newPlanCurrency string

	for _, comp := range oldPlan.Components {
		if comp.Type == models.ComponentFlat {
			oldPlanBasePrice += comp.UnitPrice
			oldPlanCurrency = comp.Currency
		}
	}
	for _, comp := range newPlan.Components {
		if comp.Type == models.ComponentFlat {
			newPlanBasePrice += comp.UnitPrice
			newPlanCurrency = comp.Currency
		}
	}
	if oldPlanCurrency == "" {
		oldPlanCurrency = "USD"
	}
	if newPlanCurrency == "" {
		newPlanCurrency = "USD"
	}

	// 4. Calculate Proration
	now := time.Now()
	totalDuration := cycle.CycleEnd.Sub(cycle.CycleStart)
	remainingDuration := cycle.CycleEnd.Sub(now)

	if totalDuration <= 0 || remainingDuration <= 0 {
		return errors.New("cycle timing is invalid or expired")
	}

	totalDays := float64(totalDuration / (24 * time.Hour))
	remainingDays := float64(remainingDuration / (24 * time.Hour))

	// Old plan credit (Refund for unused days)
	oldDailyRate := float64(oldPlanBasePrice) / totalDays
	unusedCredit := oldDailyRate * remainingDays

	// New plan debit (Charge for remaining days on new plan)
	newDailyRate := float64(newPlanBasePrice) / totalDays
	newCharge := newDailyRate * remainingDays

	// 5. Issue Wallet Credit for the unused Old Plan
	desc := fmt.Sprintf("Proration Credit for %s to %s", oldPlan.Name, newPlan.Name)
	err = s.ledgerServ.ApplyCredit(ctx, tenantID, unusedCredit, oldPlanCurrency, models.TxTypeRefund, &sub.ID, &desc)
	if err != nil {
		return err
	}

	// 6. Generate an immediate invoice for the proration difference
	// Instead of a full invoice, we charge just the newCharge. The customer can pay using wallet credit.
	prorationItem := models.InvoiceItem{
		ID:          uuid.New(),
		Description: fmt.Sprintf("Prorated Upgrade from %s to %s", oldPlan.Name, newPlan.Name),
		Quantity:    1,
		UnitPrice:   newCharge,
		TaxPercent:  0, // Assuming 0 for now
		Total:       newCharge,
	}

	inv := &models.Invoice{
		ID:            uuid.New(),
		TenantID:         tenantID,
		InvoiceNumber: fmt.Sprintf("PRORATE-%d", time.Now().UnixNano()),
		CustomerName:  "Customer", // Placeholder, would fetch real org name normally
		InvoiceDate:   now,
		DueDate:       now,
		Status:        models.StatusPending,
		Currency:      newPlanCurrency,
		Items:         []models.InvoiceItem{prorationItem},
		TotalAmount:   newCharge,
		BalanceDue:    newCharge,
	}

	err = s.billServ.CreateInvoice(ctx, inv)
	if err != nil {
		return err
	}

	// 7. Update subscription ID to point to the new plan
	sub.PlanID = newPlanID
	err = s.subServ.UpdateSubscription(ctx, sub)
	return err
}

// RedeemCoupon validates and associates a discount tracking object
func (s *AdvancedSubService) RedeemCoupon(ctx context.Context, tenantID uuid.UUID, code string, subID, invID *uuid.UUID) error {
	coupon, err := s.repo.GetCouponByCode(ctx, code)
	if err != nil || coupon == nil {
		return errors.New("invalid or non-existent coupon")
	}

	redemption := &models.CouponRedemption{
		CouponID:       coupon.ID,
		TenantID:          tenantID,
		SubscriptionID: subID,
		InvoiceID:      invID,
	}

	return s.repo.RedeemCoupon(ctx, redemption)
}
