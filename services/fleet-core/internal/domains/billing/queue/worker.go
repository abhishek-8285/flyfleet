package queue

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/flyfleet/fleet-core/internal/domains/billing/storage"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// BillingConfigProvider defines the interface needed for organization branding
type BillingConfigProvider interface {
	GetConfig(ctx context.Context, tenantID uuid.UUID) (*models.BillingConfiguration, error)
}

// DocumentProvider defines the interface for invoice document storage
type DocumentProvider interface {
	Create(ctx context.Context, d *models.InvoiceDocument) error
	GetByInvoiceID(ctx context.Context, invoiceID string) ([]models.InvoiceDocument, error)
}

// InvoiceWorker handles asynchronous billing tasks
type InvoiceWorker struct {
	repo                 *repository.BillingRepository
	templateService      *pdf.TemplateService
	pdfGenerator         *pdf.PDFGenerator
	storage              storage.ObjectStorage
	docRepo              DocumentProvider
	billingConfigService BillingConfigProvider
	signer               *pdf.Signer
	metrics              *pdf.Metrics
	logger               *zap.Logger
}

func NewInvoiceWorker(
	repo *repository.BillingRepository,
	ts *pdf.TemplateService,
	pg *pdf.PDFGenerator,
	st storage.ObjectStorage,
	dr DocumentProvider,
	bc BillingConfigProvider,
	s *pdf.Signer,
	mh *pdf.Metrics,
	logger *zap.Logger,
) *InvoiceWorker {
	return &InvoiceWorker{
		repo:                 repo,
		templateService:      ts,
		pdfGenerator:         pg,
		storage:              st,
		docRepo:              dr,
		billingConfigService: bc,
		signer:               s,
		metrics:              mh,
		logger:               logger,
	}
}

// ProcessTask handles the PDF rendering task
func (w *InvoiceWorker) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if t.Type() != TaskInvoiceRender {
		return fmt.Errorf("unexpected task type: %s", t.Type())
	}

	var p struct {
		InvoiceID string `json:"invoice_id"`
	}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	invoiceID, err := uuid.Parse(p.InvoiceID)
	if err != nil {
		return fmt.Errorf("invalid invoice id: %w", err)
	}

	// 1. Fetch Invoice and Items
	invoice, err := (*w.repo).GetInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// 2. Fetch Branding Config
	config, err := w.billingConfigService.GetConfig(ctx, invoice.TenantID)
	if err != nil {
		w.logger.Warn("Failed to fetch billing config, using defaults", zap.Error(err), zap.String("tenant_id", invoice.TenantID.String()))
		config = models.DefaultBillingConfig(invoice.TenantID)
	}

	// 3. Fetch Compiled Template
	tmpl, err := w.templateService.GetTemplate(ctx, config.Theme, 0)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// 4. Create Rendering Snapshot
	snapshot := map[string]interface{}{
		"invoice": invoice,
		"config":  config,
	}

	// 5. Render Template to HTML
	var htmlBuf bytes.Buffer
	renderData := struct {
		Invoice *models.Invoice
		Config  *models.BillingConfiguration
	}{
		Invoice: invoice,
		Config:  config,
	}
	if err := tmpl.Execute(&htmlBuf, renderData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 6. Generate PDF from HTML
	pdfData, err := w.pdfGenerator.GenerateFromHTML(ctx, htmlBuf.String(), pdf.DefaultA4Portrait())
	if err != nil {
		w.metrics.IncFailure(tmpl.Name(), "pdf_generation_failed")
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	start := time.Now()
	defer func() {
		w.metrics.ObserveRender(tmpl.Name(), time.Since(start).Seconds())
	}()

	// 7. Calculate Checksum (SHA256)
	hash := sha256.Sum256(pdfData)
	checksum := hex.EncodeToString(hash[:])

	// 8. Sign PDF (Compliance)
	signedData, err := w.signer.SignPDF(pdfData)
	if err != nil {
		w.metrics.IncFailure(tmpl.Name(), "pdf_signing_failed")
		return fmt.Errorf("failed to sign PDF: %w", err)
	}

	// 9. Upload to ObjectStorage
	if err := w.storage.Save(ctx, invoice, signedData); err != nil {
		w.metrics.IncFailure(tmpl.Name(), "storage_save_failed")
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	// 10. Store PDF URL and Snapshot in registry
	doc := &models.InvoiceDocument{
		InvoiceID:       invoice.ID,
		TemplateVersion: 1, // Mock: In real system, version from metadata
		Snapshot:        snapshot,
		FileURL:         fmt.Sprintf("https://storage.flyfleet.io/invoices/%s.pdf", invoice.ID.String()),
		StorageKey:      fmt.Sprintf("invoices/%s.pdf", invoice.ID.String()),
		Checksum:        checksum,
	}
	if err := w.docRepo.Create(ctx, doc); err != nil {
		w.metrics.IncFailure(tmpl.Name(), "doc_repo_create_failed")
		return fmt.Errorf("failed to register document: %w", err)
	}

	w.logger.Info("Invoice PDF rendered and stored successfully",
		zap.String("invoice_id", invoice.ID.String()),
		zap.String("document_id", doc.ID.String()),
	)

	return nil
}

// RegenerateFromSnapshot handles re-rendering a PDF using a previous snapshot to ensure version locking
func (w *InvoiceWorker) RegenerateFromSnapshot(ctx context.Context, invoiceID uuid.UUID) ([]byte, error) {
	// Fetch Document to get snapshot and version
	docs, err := w.docRepo.GetByInvoiceID(ctx, invoiceID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice document: %w", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("invoice document not found for regeneration")
	}

	doc := &docs[0] // Get latest

	// Fetch specific template version locked in the document
	// We use the version saved in the document.
	tmpl, err := w.templateService.GetTemplate(ctx, "professional", doc.TemplateVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get locked template version: %w", err)
	}

	// Re-render using Snapshot data
	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, doc.Snapshot); err != nil {
		return nil, fmt.Errorf("failed to execute template with snapshot: %w", err)
	}

	// Generate PDF
	pdfData, err := w.pdfGenerator.GenerateFromHTML(ctx, htmlBuf.String(), pdf.DefaultA4Portrait())
	if err != nil {
		return nil, fmt.Errorf("failed to re-generate PDF: %w", err)
	}

	return pdfData, nil
}
