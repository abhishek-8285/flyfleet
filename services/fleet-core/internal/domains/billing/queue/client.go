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

// BillingQueueClient provides methods to enqueue billing tasks
type BillingQueueClient struct {
	client *asynq.Client
	logger *zap.Logger
}

func NewBillingQueueClient(redisAddr string, logger *zap.Logger) *BillingQueueClient {
	return &BillingQueueClient{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
		logger: logger,
	}
}

// Close closes the connection to redis
func (c *BillingQueueClient) Close() error {
	return c.client.Close()
}

// EnqueueGenerateInvoice adds a monthly invoice generation task to the queue
func (c *BillingQueueClient) EnqueueGenerateInvoice(ctx context.Context, payload models.GenerateInvoicePayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(models.TypeGenerateInvoice, data)
	info, err := c.client.EnqueueContext(ctx, task, asynq.Queue("billing"))
	if err != nil {
		return fmt.Errorf("enqueue task: %w", err)
	}

	c.logger.Info("Enqueued invoice generation task",
		zap.String("org_id", payload.OrgID.String()),
		zap.String("task_id", info.ID))
	return nil
}

// EnqueueDunningTask adds a payment retry or dunning notification task
func (c *BillingQueueClient) EnqueueDunningTask(ctx context.Context, payload models.ProcessDunningPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(models.TypeProcessDunning, data)
	// Add delay or schedule for the retry
	info, err := c.client.EnqueueContext(ctx, task, asynq.Queue("billing"))
	if err != nil {
		return fmt.Errorf("enqueue dunning task: %w", err)
	}

	c.logger.Info("Enqueued dunning task",
		zap.String("invoice_id", payload.InvoiceID.String()),
		zap.String("task_id", info.ID))
	return nil
}

// EnqueueRenderPDF adds a PDF rendering task to the queue
func (c *BillingQueueClient) EnqueueRenderPDF(ctx context.Context, invoiceID uuid.UUID) error {
	payload := models.RenderPDFPayload{
		InvoiceID: invoiceID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(models.TypeRenderPDF, data)
	info, err := c.client.EnqueueContext(ctx, task, asynq.Queue("billing"))
	if err != nil {
		return fmt.Errorf("enqueue render pdf task: %w", err)
	}

	c.logger.Info("Enqueued PDF rendering task",
		zap.String("invoice_id", invoiceID.String()),
		zap.String("task_id", info.ID))
	return nil
}
