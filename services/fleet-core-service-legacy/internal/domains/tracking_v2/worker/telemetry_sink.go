package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/repository"
	"go.uber.org/zap"
)

const (
	batchSize     = 1000
	flushInterval = 500 * time.Millisecond
)

// TelemetryBulkSinkWorker consumes raw telemetry from NATS and persists it in batches.
type TelemetryBulkSinkWorker struct {
	mb        messaging.MessageBroker
	repo      repository.TelemetryRepository
	logger    *zap.Logger
	
	buffer []models.TelemetryEvent
	mu     sync.Mutex
	stop   chan struct{}
}

func NewTelemetryBulkSinkWorker(mb messaging.MessageBroker, repo repository.TelemetryRepository, logger *zap.Logger) *TelemetryBulkSinkWorker {
	return &TelemetryBulkSinkWorker{
		mb:     mb,
		repo:   repo,
		logger: logger,
		buffer: make([]models.TelemetryEvent, 0, batchSize),
		stop:   make(chan struct{}),
	}
}

func (w *TelemetryBulkSinkWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting Telemetry Bulk Sink Worker", zap.Int("batch_size", batchSize), zap.Duration("flush_interval", flushInterval))

	// Subscribe to the raw telemetry stream
	err := w.mb.SubscribeJS("telemetry.raw", "telemetry-sink", "telemetry_persistence", func(data []byte) {
		var event models.TelemetryEvent
		if err := json.Unmarshal(data, &event); err != nil {
			w.logger.Error("failed to unmarshal telemetry event", zap.Error(err))
			return
		}

		w.mu.Lock()
		w.buffer = append(w.buffer, event)
		shouldFlush := len(w.buffer) >= batchSize
		w.mu.Unlock()

		if shouldFlush {
			w.flush()
		}
	})

	if err != nil {
		return err
	}

	// Periodic flush timer
	ticker := time.NewTicker(flushInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				w.flush()
			case <-w.stop:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (w *TelemetryBulkSinkWorker) flush() {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return
	}
	
	toFlush := w.buffer
	w.buffer = make([]models.TelemetryEvent, 0, batchSize)
	w.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	if err := w.repo.BulkIngest(ctx, toFlush); err != nil {
		w.logger.Error("bulk flush failed", zap.Error(err), zap.Int("count", len(toFlush)))
		// In a production scenario, we might want to dead-letter these or retry.
		// For stabilization MVP, we log and drop to prevent worker hang.
	} else {
		w.logger.Debug("bulk flush successful", 
			zap.Int("count", len(toFlush)), 
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func (w *TelemetryBulkSinkWorker) Stop() {
	close(w.stop)
}
