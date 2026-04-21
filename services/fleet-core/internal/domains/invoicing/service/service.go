package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/flyfleet/fleet-core/internal/domains/invoicing/repository"
	"github.com/google/uuid"
)

type InvoicingService struct {
	repo         repository.InvoiceRepository
	pdfGenerator *pdf.PDFGenerator
}

func NewInvoicingService(repo repository.InvoiceRepository) *InvoicingService {
	return &InvoicingService{
		repo:         repo,
		pdfGenerator: pdf.NewPDFGenerator(30*time.Second, ""), // Uses discovery logic
	}
}

func (s *InvoicingService) CreateInvoice(ctx context.Context, tenantID uuid.UUID, inv *invoicing.Invoice) (*invoicing.Invoice, error) {
	// 1. Validation
	if inv.CustomerName == "" {
		return nil, fmt.Errorf("customer name is required")
	}
	if inv.CustomerPhone == "" {
		return nil, fmt.Errorf("customer phone number is mandatory")
	}

	// 2. Enrich invoice
	inv.TenantID = tenantID
	inv.InvoiceNumber = s.generateInvoiceNumber()
	inv.Status = invoicing.StatusPending
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = time.Now()

	// 2. Calculate totals
	var subtotal float64
	var totalTax float64
	for i := range inv.Items {
		inv.Items[i].TenantID = tenantID
		inv.Items[i].Amount = inv.Items[i].Quantity * inv.Items[i].UnitPrice
		
		// Use provided tax percent or default to 0
		taxRate := inv.Items[i].TaxPercent / 100.0
		itemTax := inv.Items[i].Amount * taxRate
		
		subtotal += inv.Items[i].Amount
		totalTax += itemTax
	}
	inv.Subtotal = subtotal
	inv.TaxAmount = totalTax
	inv.TotalAmount = inv.Subtotal + inv.TaxAmount
	inv.AmountPaid = 0
	inv.AmountRemaining = inv.TotalAmount

	// 3. Persist
	return s.repo.Create(ctx, inv)
}

func (s *InvoicingService) GetInvoice(ctx context.Context, tenantID, id uuid.UUID) (*invoicing.Invoice, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *InvoicingService) ListInvoices(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error) {
	return s.repo.List(ctx, tenantID, limit, offset)
}

func (s *InvoicingService) VoidInvoice(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, tenantID, id, invoicing.StatusVoid, uuid.Nil)
}

func (s *InvoicingService) GeneratePDF(ctx context.Context, tenantID, id uuid.UUID) ([]byte, string, error) {
	inv, err := s.GetInvoice(ctx, tenantID, id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get invoice: %w", err)
	}

	html := GenerateInvoiceHTML(inv)
	pdfData, err := s.pdfGenerator.GenerateFromHTML(ctx, html, pdf.DefaultA4Portrait())
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	filename := fmt.Sprintf("invoice-%s.pdf", inv.InvoiceNumber)
	return pdfData, filename, nil
}

func (s *InvoicingService) generateInvoiceNumber() string {
	return fmt.Sprintf("INV-%d", time.Now().UnixNano()/1e6)
}
