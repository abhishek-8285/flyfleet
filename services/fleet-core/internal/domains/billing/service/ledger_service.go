package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LedgerService struct {
	repo   *repository.LedgerRepository
	logger *zap.Logger
}

func NewLedgerService(repo *repository.LedgerRepository, logger *zap.Logger) *LedgerService {
	return &LedgerService{
		repo:   repo,
		logger: logger,
	}
}

// GetBalance retrieves the current prepaid balance for an organization
func (s *LedgerService) GetBalance(ctx context.Context, tenantID uuid.UUID, currency string) (*models.CustomerBalance, error) {
	return s.repo.GetOrCreateBalance(ctx, tenantID, currency)
}

// ApplyCredit adds promotional or prepaid credit to a customer's wallet
func (s *LedgerService) ApplyCredit(ctx context.Context, tenantID uuid.UUID, amount float64, currency, refType string, refID *uuid.UUID, desc *string) error {
	balance, err := s.repo.GetOrCreateBalance(ctx, tenantID, currency)
	if err != nil {
		return err
	}

	tx := &models.CreditTransaction{
		BalanceID:       balance.ID,
		TransactionType: models.TxTypeManualCredit,
		Amount:          amount,
		Currency:        currency,
		ReferenceType:   &refType,
		ReferenceID:     refID,
		Description:     desc,
	}

	if err := s.repo.ApplyCreditTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to apply credit", zap.Error(err), zap.String("tenant_id", tenantID.String()))
		return fmt.Errorf("failed to apply credit: %w", err)
	}

	// 2. Ledger Recording
	ledgerEntry := &models.LedgerEntry{
		TenantID:         tenantID,
		EntryType:     models.LedgerTypeCreditApplied,
		Amount:        amount, // Money goes into Customer Credits list
		Currency:      currency,
		LedgerAccount: models.AccountCustomerCredits,
		Description:   desc,
	}
	_ = s.repo.RecordLedgerEntry(ctx, ledgerEntry)

	return nil
}

// RecordInvoiceIssue logs an immutable general ledger entry when a final invoice is generated
func (s *LedgerService) RecordInvoiceIssue(ctx context.Context, tenantID uuid.UUID, invoiceID uuid.UUID, amount float64, currency string) error {
	desc := "Invoice Issued"
	entry := &models.LedgerEntry{
		TenantID:         tenantID,
		EntryType:     models.LedgerTypeInvoiceIssued,
		Amount:        amount,
		Currency:      currency,
		InvoiceID:     &invoiceID,
		LedgerAccount: models.AccountAccountsReceivable,
		Description:   &desc,
	}
	return s.repo.RecordLedgerEntry(ctx, entry)
}

// RecordPaymentReceived logs successful payments mapping AR to Cash
func (s *LedgerService) RecordPaymentReceived(ctx context.Context, tenantID uuid.UUID, invoiceID, paymentID uuid.UUID, amount float64, currency string) error {
	desc := "Payment Received"
	entry := &models.LedgerEntry{
		TenantID:         tenantID,
		EntryType:     models.LedgerTypePaymentReceived,
		Amount:        amount,
		Currency:      currency,
		InvoiceID:     &invoiceID,
		PaymentID:     &paymentID,
		LedgerAccount: models.AccountCash,
		Description:   &desc,
	}
	return s.repo.RecordLedgerEntry(ctx, entry)
}

// RecordCreditApplied logs the application of a credit note against an invoice
func (s *LedgerService) RecordCreditApplied(ctx context.Context, tenantID uuid.UUID, invoiceID, creditNoteID uuid.UUID, amount float64, currency string) error {
	desc := "Credit Note Applied"
	entry := &models.LedgerEntry{
		TenantID:         tenantID,
		EntryType:     models.LedgerTypeCreditApplied,
		Amount:        -amount, // applying a credit reduces accounts receivable
		Currency:      currency,
		InvoiceID:     &invoiceID,
		CreditNoteID:  &creditNoteID,
		LedgerAccount: models.AccountAccountsReceivable,
		Description:   &desc,
	}
	return s.repo.RecordLedgerEntry(ctx, entry)
}

// ScheduleRevenue ensures that upfront payments for subscriptions (e.g., Annual plans) are set to be recognized over time rather than all at once.
func (s *LedgerService) ScheduleRevenue(ctx context.Context, tenantID uuid.UUID, invoiceLineID uuid.UUID, amount float64, currency string, start, end time.Time) error {
	schedule := &models.RevenueSchedule{
		TenantID:             tenantID,
		InvoiceLineItemID: invoiceLineID,
		TotalAmount:       amount,
		Currency:          currency,
		StartDate:         start,
		EndDate:           end,
		Status:            "PENDING",
	}
	return s.repo.CreateRevenueSchedule(ctx, schedule)
}
