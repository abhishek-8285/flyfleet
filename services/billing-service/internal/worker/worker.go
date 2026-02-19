package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const (
	TypeMonthlyInvoiceGeneration = "billing:monthly_invoice_generation"
)

type Worker struct {
	srv    *asynq.Server
	logger *zap.Logger
}

func NewWorker(redisURL string, logger *zap.Logger) *Worker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisURL},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	return &Worker{
		srv:    srv,
		logger: logger,
	}
}

func (w *Worker) Start(mux *asynq.ServeMux) error {
	w.logger.Info("starting billing worker")
	return w.srv.Run(mux)
}

func (w *Worker) Stop() {
	w.srv.Shutdown()
}

// Job Handlers

func HandleMonthlyInvoiceGeneration(ctx context.Context, t *asynq.Task) error {
	var p map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	fmt.Printf("Generating monthly invoices for organization %s\n", p["org_id"])
	return nil
}

