package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// BillingProcessor defines necessary billing operations
type BillingProcessor interface {
	GenerateDraftInvoice(ctx context.Context, orgID uuid.UUID) (*models.Invoice, error)
	GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*models.Invoice, error)
}

// PDFRenderer defines necessary PDF operations
type PDFRenderer interface {
	GetOrGeneratePDF(ctx context.Context, invoice *models.Invoice) ([]byte, error)
}

// BillingWorkerServer handles asynchronous billing tasks
type BillingWorkerServer struct {
	server      *asynq.Server
	mux         *asynq.ServeMux
	service     BillingProcessor
	invoiceServ PDFRenderer
	logger      *zap.Logger
}

func NewBillingWorkerServer(redisAddr string, s BillingProcessor, invS PDFRenderer, logger *zap.Logger) *BillingWorkerServer {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"billing": 1,
			},
			Logger: logger.Sugar(),
		},
	)

	mux := asynq.NewServeMux()
	ws := &BillingWorkerServer{
		server:      server,
		mux:         mux,
		service:     s,
		invoiceServ: invS,
		logger:      logger,
	}

	mux.HandleFunc(models.TypeGenerateInvoice, ws.HandleGenerateInvoice)
	mux.HandleFunc(models.TypeProcessDunning, ws.HandleProcessDunning)
	mux.HandleFunc(models.TypeRenderPDF, ws.HandleRenderPDF)

	return ws
}

// Run starts the worker server
func (ws *BillingWorkerServer) Run() error {
	return ws.server.Run(ws.mux)
}

// HandleGenerateInvoice processes the monthly invoice generation
func (ws *BillingWorkerServer) HandleGenerateInvoice(ctx context.Context, t *asynq.Task) error {
	var payload models.GenerateInvoicePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	ws.logger.Info("Generating automated invoice", zap.String("tenant_id", payload.OrgID.String()))

	invoice, err := ws.service.GenerateDraftInvoice(ctx, payload.OrgID)
	if err != nil {
		ws.logger.Error("Failed to generate automated invoice", zap.Error(err), zap.String("tenant_id", payload.OrgID.String()))
		return fmt.Errorf("generate draft invoice failed: %w", err)
	}

	ws.logger.Info("Successfully generated automated draft invoice",
		zap.String("tenant_id", payload.OrgID.String()),
		zap.String("invoice_id", invoice.ID.String()))

	return nil
}

// HandleProcessDunning handles payment retry and user notification
func (ws *BillingWorkerServer) HandleProcessDunning(ctx context.Context, t *asynq.Task) error {
	var payload models.ProcessDunningPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	ws.logger.Info("Processing dunning task", zap.String("invoice_id", payload.InvoiceID.String()))
	// Placeholder for payment gateway retry and email/SMS alerts logic
	return nil
}

// HandleRenderPDF processes asynchronous PDF generation and storage
func (ws *BillingWorkerServer) HandleRenderPDF(ctx context.Context, t *asynq.Task) error {
	var payload models.RenderPDFPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	ws.logger.Info("Rendering invoice PDF background task", zap.String("invoice_id", payload.InvoiceID.String()))

	// Use InvoiceService to get or generate PDF (it will save to storage)
	// We first need the invoice object
	invoice, err := ws.service.GetInvoice(ctx, payload.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice not found: %s: %w", payload.InvoiceID, asynq.SkipRetry)
	}

	_, err = ws.invoiceServ.GetOrGeneratePDF(ctx, invoice)
	if err != nil {
		ws.logger.Error("Failed to render invoice PDF", zap.Error(err), zap.String("invoice_id", payload.InvoiceID.String()))
		return err
	}

	ws.logger.Info("Successfully rendered and stored invoice PDF", zap.String("invoice_id", payload.InvoiceID.String()))
	return nil
}
