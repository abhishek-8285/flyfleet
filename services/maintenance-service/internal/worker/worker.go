package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const (
	TypeMaintenanceReminder = "maintenance:reminder"
	TypeOverdueChecker      = "maintenance:overdue_checker"
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
	w.logger.Info("starting asynq worker")
	return w.srv.Run(mux)
}

func (w *Worker) Stop() {
	w.srv.Shutdown()
}

// Job Handlers

func HandleMaintenanceReminder(ctx context.Context, t *asynq.Task) error {
	var p map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Logic to send notification...
	fmt.Printf("Sending maintenance reminder for org %s, schedule %s\n", p["org_id"], p["id"])
	return nil
}

func HandleOverdueChecker(ctx context.Context, t *asynq.Task) error {
	// Logic to scan DB for overdue schedules...
	fmt.Println("Scanning for overdue maintenance schedules...")
	return nil
}

