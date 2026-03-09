package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TaskInvoiceRender = "invoice:render_pdf"
)

// TaskProducer handles enqueuing tasks to Asynq
type TaskProducer struct {
	client *asynq.Client
}

func NewTaskProducer(client *asynq.Client) *TaskProducer {
	return &TaskProducer{client: client}
}

func (p *TaskProducer) EnqueueInvoiceRender(ctx context.Context, invoiceID string) error {
	payload := map[string]interface{}{"invoice_id": invoiceID}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskInvoiceRender, data)
	_, err = p.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(10),
		asynq.Timeout(2*time.Minute),
		asynq.Queue("billing"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	return nil
}
func (p *TaskProducer) EnqueueBulkInvoiceRender(ctx context.Context, invoiceIDs []string) error {
	for _, id := range invoiceIDs {
		if err := p.EnqueueInvoiceRender(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
