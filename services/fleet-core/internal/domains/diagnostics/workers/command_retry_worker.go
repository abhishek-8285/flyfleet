package workers

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/service"
)

type CommandRetryWorker struct {
	cmdService *service.DeviceCommandService
	maxRetries int
}

func NewCommandRetryWorker(cs *service.DeviceCommandService) *CommandRetryWorker {
	return &CommandRetryWorker{
		cmdService: cs,
		maxRetries: 5,
	}
}

func (w *CommandRetryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("CommandRetryWorker stopped")
			return
		case <-ticker.C:
			w.Process(ctx)
		}
	}
}

func (w *CommandRetryWorker) Process(ctx context.Context) {
	log.Println("Running command retry matching logic")
	// Logic to pull failed/stuck commands and retry them
}

func (w *CommandRetryWorker) calculateBackoff(retryCount int) time.Duration {
	if retryCount >= w.maxRetries {
		return 0 // Stop retrying
	}
	// Exponential backoff: 2^retryCount * 1 minute
	backoffMinutes := math.Pow(2, float64(retryCount))
	return time.Duration(backoffMinutes) * time.Minute
}
