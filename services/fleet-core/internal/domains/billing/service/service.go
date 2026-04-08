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

// QueueProvider defines the interface for enqueuing billing tasks
type QueueProvider interface {
	EnqueueGenerateInvoice(ctx context.Context, payload models.GenerateInvoicePayload) error
	EnqueueDunningTask(ctx context.Context, payload models.ProcessDunningPayload) error
	EnqueueRenderPDF(ctx context.Context, invoiceID uuid.UUID) error
}

type Repository interface {
	ListInvoices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Invoice, int, error)
	CreateInvoice(ctx context.Context, inv *models.Invoice) error
	GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	UpdateInvoice(ctx context.Context, inv *models.Invoice) error
	DeleteInvoice(ctx context.Context, id uuid.UUID) error
	GetSettings(ctx context.Context, orgID uuid.UUID) (*models.BillingSettings, error)
	UpdateSettings(ctx context.Context, s *models.BillingSettings) error
	GetMetrics(ctx context.Context, orgID uuid.UUID) (*models.BillingMetrics, error)
	RecordPayment(ctx context.Context, id uuid.UUID, amount float64) error
	CreateTaxLines(ctx context.Context, lines []models.InvoiceTaxLine) error
}

type BillingService struct {
	repo        Repository
	auditRepo   *repository.AuditLogRepository
	subServ     *SubscriptionService
	meterServ   *MeteringService
	hwServ      *HardwareService
	ledgerServ  *LedgerService
	taxServ     *TaxService
	seqServ     *SequenceService
	notifServ   *NotificationService
	dunningS    *DunningService
	queueClient QueueProvider
	logger      *zap.Logger
}

func NewBillingService(repo Repository, auditRepo *repository.AuditLogRepository, subServ *SubscriptionService, meterServ *MeteringService, hwServ *HardwareService, ledgerServ *LedgerService, taxServ *TaxService, seqServ *SequenceService, notifServ *NotificationService, dunningS *DunningService, qClient QueueProvider, logger *zap.Logger) *BillingService {
	return &BillingService{
		repo:        repo,
		auditRepo:   auditRepo,
		subServ:     subServ,
		meterServ:   meterServ,
		hwServ:      hwServ,
		ledgerServ:  ledgerServ,
		taxServ:     taxServ,
		seqServ:     seqServ,
		notifServ:   notifServ,
		dunningS:    dunningS,
		queueClient: qClient,
		logger:      logger,
	}
}

func (s *BillingService) ListInvoices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Invoice, int, error) {
	return s.repo.ListInvoices(ctx, orgID, page, size)
}

func (s *BillingService) CreateInvoice(ctx context.Context, inv *models.Invoice) error {
	// Validate due date
	if inv.DueDate.Before(inv.InvoiceDate) || inv.DueDate.Equal(inv.InvoiceDate) {
		return ErrDueDateBeforeInvoiceDate
	}

	// Validate discount
	if err := s.validateDiscount(inv); err != nil {
		return err
	}

	// Validate amounts are positive
	if inv.TotalAmount < 0 {
		return ErrNegativeAmount
	}

	// Calculate line item totals
	var subtotal float64
	for i := range inv.Items {
		item := &inv.Items[i]

		// Validate item quantity and price
		if item.Quantity <= 0 {
			return errors.New("item quantity must be positive")
		}
		if item.UnitPrice < 0 {
			return errors.New("item unit price cannot be negative")
		}

		itemTotal := float64(item.Quantity) * item.UnitPrice
		item.Total = itemTotal + (itemTotal * (item.TaxPercent / 100))
		subtotal += itemTotal
	}

	// Calculate discount
	inv.Subtotal = subtotal
	inv.DiscountAmount = models.CalculateDiscount(subtotal, models.DiscountType(inv.DiscountType), inv.DiscountValue)

	// Calculate tax on discounted amount
	var totalTax float64
	for i := range inv.Items {
		item := &inv.Items[i]
		itemTotal := float64(item.Quantity) * item.UnitPrice
		totalTax += itemTotal * (item.TaxPercent / 100)
	}

	// Total = Subtotal - Discount + Tax
	inv.TotalAmount = subtotal - inv.DiscountAmount + totalTax
	inv.BalanceDue = inv.TotalAmount - inv.AmountPaid

	if inv.Status == "" {
		inv.Status = models.StatusPending
	}

	// Validate status
	validStatuses := map[string]bool{
		"DRAFT": true, "PENDING": true, "PAID": true, "OVERDUE": true, "PARTIAL": true,
	}
	if !validStatuses[string(inv.Status)] {
		return ErrInvalidInvoiceStatus
	}

	// 1. Generate Gapless Invoice Number if DRAFT or PENDING and not provided
	if (inv.Status == "DRAFT" || inv.Status == "PENDING") && inv.InvoiceNumber == "" {
		if s.seqServ != nil {
			num, err := s.seqServ.GetNextInvoiceNumber(ctx, inv.OrgID)
			if err != nil {
				s.logger.Error("Failed to generate invoice number", zap.Error(err), zap.String("org_id", inv.OrgID.String()))
			} else {
				inv.InvoiceNumber = num
			}
		}
	}

	// 2. Fetch Billing Settings for Tax Region
	settings, err := s.repo.GetSettings(ctx, inv.OrgID)
	if s.taxServ != nil && settings != nil && settings.TaxCountry != "" {
		// Calculate Regional Taxes
		taxLines, err := s.taxServ.CalculateInvoiceTaxes(ctx, inv, settings.TaxCountry, settings.TaxState)
		if err != nil {
			s.logger.Error("Failed to calculate taxes", zap.Error(err), zap.String("org_id", inv.OrgID.String()))
		} else if len(taxLines) > 0 {
			// Update Invoice amounts based on calculated taxes
			var totalTax float64
			for _, tl := range taxLines {
				totalTax += tl.Amount
			}
			inv.TotalAmount = inv.Subtotal - inv.DiscountAmount + totalTax
			inv.BalanceDue = inv.TotalAmount - inv.AmountPaid

			// Save tax lines later after invoice creation
			defer func(lines []models.InvoiceTaxLine) {
				_ = s.repo.CreateTaxLines(ctx, lines)
			}(taxLines)
		}
	}

	err = s.repo.CreateInvoice(ctx, inv)
	if err != nil {
		return err
	}

	// Fire NATS notification for invoice created
	if s.notifServ != nil {
		_ = s.notifServ.PublishEvent(ctx, "invoice_created", inv.OrgID, map[string]interface{}{
			"invoice_id":     inv.ID,
			"invoice_number": inv.InvoiceNumber,
			"total_amount":   inv.TotalAmount,
			"currency":       inv.Currency,
		})
	}

	// Enqueue PDF rendering task
	if s.queueClient != nil {
		if err := s.queueClient.EnqueueRenderPDF(ctx, inv.ID); err != nil {
			s.logger.Error("Failed to enqueue PDF rendering task", zap.Error(err), zap.String("invoice_id", inv.ID.String()))
		}
	}

	// Audit log - invoice created
	if s.auditRepo != nil {
		_ = s.auditRepo.CreateAuditLogEntry(ctx, inv.ID, models.AuditActionCreated, nil, inv, inv.CreatedBy, "", "", "", "")
	}

	// Record Ledger Entry
	if s.ledgerServ != nil {
		_ = s.ledgerServ.RecordInvoiceIssue(ctx, inv.OrgID, inv.ID, inv.TotalAmount, inv.Currency)
	}

	return nil
}

// validateDiscount validates invoice discount configuration
func (s *BillingService) validateDiscount(inv *models.Invoice) error {
	if inv.DiscountType == "" {
		inv.DiscountType = string(models.DiscountNone)
	}

	if !models.IsValidDiscountType(inv.DiscountType) {
		return errors.New("invalid discount type")
	}

	if inv.DiscountValue < 0 {
		return errors.New("discount value cannot be negative")
	}

	if models.DiscountType(inv.DiscountType) == models.DiscountPercentage && inv.DiscountValue > 100 {
		return errors.New("percentage discount cannot exceed 100%")
	}

	return nil
}

func (s *BillingService) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	return s.repo.GetInvoice(ctx, id)
}

func (s *BillingService) UpdateInvoice(ctx context.Context, inv *models.Invoice) error {
	// CRITICAL: Check if invoice is editable
	existing, err := s.repo.GetInvoice(ctx, inv.ID)
	if err != nil {
		return err
	}

	if existing == nil {
		return ErrInvoiceNotFound
	}

	// Immutability check - cannot edit issued/paid invoices
	if !IsInvoiceEditable(string(existing.Status)) {
		return ErrInvoiceImmutable
	}

	// Validate status transition if status is changing
	if inv.Status != existing.Status {
		if !IsValidStatusTransition(string(existing.Status), string(inv.Status)) {
			return ErrInvalidStatusTransition
		}
	}

	// Recalculate totals
	total := 0.0
	for i := range inv.Items {
		item := &inv.Items[i]
		itemTotal := float64(item.Quantity) * item.UnitPrice
		item.Total = itemTotal + (itemTotal * (item.TaxPercent / 100))
		total += item.Total
	}
	inv.TotalAmount = total
	inv.BalanceDue = total - inv.AmountPaid

	err = s.repo.UpdateInvoice(ctx, inv)
	if err != nil {
		return err
	}

	// Audit log - invoice updated
	if s.auditRepo != nil {
		_ = s.auditRepo.CreateAuditLogEntry(ctx, inv.ID, models.AuditActionUpdated, existing, inv, inv.UpdatedBy, "", "", "", "")
	}

	return nil
}

func (s *BillingService) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteInvoice(ctx, id)
}

func (s *BillingService) RecordPayment(ctx context.Context, id uuid.UUID, amount float64) error {
	// Validate payment amount
	if amount <= 0 {
		return ErrNegativeAmount
	}

	// Get invoice
	invoice, err := s.repo.GetInvoice(ctx, id)
	if err != nil {
		return err
	}

	if invoice == nil {
		return ErrInvoiceNotFound
	}

	// Check if invoice is already fully paid
	if invoice.Status == models.StatusPaid {
		return ErrInvoiceAlreadyPaid
	}

	// CRITICAL: Calculate outstanding amount
	outstanding := invoice.TotalAmount - invoice.AmountPaid

	// Payment cannot exceed outstanding amount
	if amount > outstanding {
		return ErrPaymentExceedsOutstanding
	}

	// Update payment amount
	invoice.AmountPaid += amount
	invoice.BalanceDue = invoice.TotalAmount - invoice.AmountPaid

	// Update status based on payment
	if invoice.BalanceDue <= 0 {
		invoice.BalanceDue = 0
		invoice.Status = models.StatusPaid
	} else if invoice.AmountPaid > 0 {
		invoice.Status = models.StatusPartial
	}

	err = s.repo.RecordPayment(ctx, id, amount)
	if err != nil {
		return err
	}

	// Record Ledger Entry
	if s.ledgerServ != nil {
		// Pass empty uuid for paymentID if none is provided via this basic API
		_ = s.ledgerServ.RecordPaymentReceived(ctx, invoice.OrgID, invoice.ID, uuid.Nil, amount, invoice.Currency)
	}

	// Fire NATS notification for payment success
	if s.notifServ != nil {
		_ = s.notifServ.PublishEvent(ctx, "payment_success", invoice.OrgID, map[string]interface{}{
			"invoice_id":     id,
			"amount_paid":    amount,
			"balance_due":    invoice.BalanceDue,
			"invoice_status": invoice.Status,
		})
	}

	return nil
}

func (s *BillingService) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.BillingSettings, error) {
	settings, err := s.repo.GetSettings(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		// Return default settings
		return &models.BillingSettings{
			OrgID:                orgID,
			DefaultTaxRate:       18.0,
			BaseCurrency:         "USD",
			GracePeriodDays:      14,
			AutoInvoicingEnabled: false,
		}, nil
	}
	return settings, nil
}

func (s *BillingService) UpdateSettings(ctx context.Context, settings *models.BillingSettings) error {
	return s.repo.UpdateSettings(ctx, settings)
}

func (s *BillingService) GetMetrics(ctx context.Context, orgID uuid.UUID) (*models.BillingMetrics, error) {
	return s.repo.GetMetrics(ctx, orgID)
}

// GenerateDraftInvoice creates a draft invoice including subscription plan and usage
func (s *BillingService) GenerateDraftInvoice(ctx context.Context, orgID uuid.UUID) (*models.Invoice, error) {
	// 1. Get Subscription
	sub, err := s.subServ.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return nil, fmt.Errorf("no active subscription found")
	}

	// 1.1 Get Plan Details (needed for components)
	plans, err := s.subServ.ListPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	var plan *models.Plan
	for _, p := range plans {
		if p.ID == sub.PlanID {
			plan = &p
			break
		}
	}
	if plan == nil {
		return nil, fmt.Errorf("subscription plan not found: %s", sub.PlanID)
	}

	// 2. Get Usage Summaries
	usageSummaries, err := s.meterServ.GetUnbilledUsageByOrg(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	// 3. Build Invoice
	inv := &models.Invoice{
		ID:            uuid.New(),
		OrgID:         orgID,
		InvoiceNumber: fmt.Sprintf("DRAFT-%s", time.Now().Format("20060102")),
		InvoiceDate:   time.Now().UTC(),
		DueDate:       time.Now().AddDate(0, 0, 14).UTC(), // Default 14 days
		Status:        models.InvoiceStatus("DRAFT"),
		Currency:      "USD",
		Items:         []models.InvoiceItem{},
	}

	// Add Plan Base Fees (Flat & Per Unit)
	for _, pc := range plan.Components {
		// For simplicity, we only add fixed components here. Usage based are handled by metering.
		if pc.Type == models.ComponentFlat || pc.Type == models.ComponentPerUnit {
			qty := 1
			if pc.Type == models.ComponentPerUnit {
				// TODO: Get actual count (e.g. active vehicles)
				// For now, default to 1 or pull from metadata if available
				qty = 5 // Placeholder for Starter
			}
			inv.Items = append(inv.Items, models.InvoiceItem{
				ID:          uuid.New(),
				InvoiceID:   inv.ID,
				Description: fmt.Sprintf("%s (%s)", pc.Name, plan.Name),
				Quantity:    qty,
				UnitPrice:   pc.UnitPrice,
				TaxPercent:  18.0, // Default tax
			})
		}
	}

	// Add Usage Items
	for _, us := range usageSummaries {
		inv.Items = append(inv.Items, models.InvoiceItem{
			ID:          uuid.New(),
			InvoiceID:   inv.ID,
			Description: us.Name,
			Quantity:    int(us.Total),
			UnitPrice:   us.Amount / us.Total, // Derived unit price
			TaxPercent:  18.0,
		})
	}

	// Save as Draft
	err = s.CreateInvoice(ctx, inv)
	if err != nil {
		return nil, fmt.Errorf("failed to save draft invoice: %w", err)
	}

	// Mark usage as billed to this draft
	if len(usageSummaries) > 0 {
		err = s.meterServ.MarkUsageAsInvoiced(ctx, orgID, inv.ID)
		if err != nil {
			s.logger.Warn("Failed to mark usage as billed", zap.Error(err))
		}
	}

	return inv, nil
}
