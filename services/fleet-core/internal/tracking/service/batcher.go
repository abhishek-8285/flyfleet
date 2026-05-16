package service

import (
	"context"
	"sync"
	"time"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
	"go.uber.org/zap"
)

// TelemetryBatcher handles high-throughput asynchronous writes to the database.
// It buffers events and flushes them using BulkIngest (COPY FROM) for performance.
type TelemetryBatcher struct {
	repo      repository.TelemetryRepositoryer
	logger    *zap.Logger
	ch        chan models.TelemetryEvent
	batchSize int
	interval  time.Duration
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewTelemetryBatcher(repo repository.TelemetryRepositoryer, logger *zap.Logger, batchSize int, interval time.Duration) *TelemetryBatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &TelemetryBatcher{
		repo:      repo,
		logger:    logger,
		ch:        make(chan models.TelemetryEvent, batchSize*2),
		batchSize: batchSize,
		interval:  interval,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start launches the background worker.
func (b *TelemetryBatcher) Start() {
	b.wg.Add(1)
	go b.worker()
}

// Stop gracefully shuts down the batcher, ensuring all buffered events are flushed.
func (b *TelemetryBatcher) Stop() {
	b.cancel()
	close(b.ch)
	b.wg.Wait()
}

// Add appends an event to the batching queue.
func (b *TelemetryBatcher) Add(event models.TelemetryEvent) {
	select {
	case b.ch <- event:
	default:
		b.logger.Warn("telemetry batcher channel full, dropping event", zap.String("event_id", event.EventID.String()))
	}
}

func (b *TelemetryBatcher) worker() {
	defer b.wg.Done()
	buffer := make([]models.TelemetryEvent, 0, b.batchSize)
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			// Flush remaining
			b.flush(buffer)
			return
		case event, ok := <-b.ch:
			if !ok {
				b.flush(buffer)
				return
			}
			buffer = append(buffer, event)
			if len(buffer) >= b.batchSize {
				b.flush(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				b.flush(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func (b *TelemetryBatcher) flush(events []models.TelemetryEvent) {
	if len(events) == 0 {
		return
	}
	// Use a background context for flushing to ensure it completes even if the worker is stopping.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := b.repo.BulkIngest(ctx, events); err != nil {
		b.logger.Error("failed to flush telemetry batch", zap.Int("count", len(events)), zap.Error(err))
	} else {
		b.logger.Debug("flushed telemetry batch", zap.Int("count", len(events)))
	}
}
