package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	metricQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_batcher_queue_depth",
		Help: "Current number of events in the telemetry batcher queue",
	})
	metricBatchFlushes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "telemetry_batcher_flushes_total",
		Help: "Total number of telemetry batch flushes",
	}, []string{"status"})
	metricBatchSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "telemetry_batcher_flush_size",
		Help:    "Size of the telemetry batches flushed to the database",
		Buckets: []float64{10, 50, 100, 500, 1000, 2000},
	})
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
	flushSig  chan chan struct{} // Channel to request a manual flush
	rdb       *redis.Client
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
		flushSig:  make(chan chan struct{}),
	}
}

func (b *TelemetryBatcher) SetRedisClient(rdb *redis.Client) {
	b.rdb = rdb
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

// Add appends an event to the batching queue. If full, blocks with a timeout before spilling over to Redis.
func (b *TelemetryBatcher) Add(event models.TelemetryEvent) {
	select {
	case b.ch <- event:
		metricQueueDepth.Inc()
	case <-time.After(200 * time.Millisecond):
		b.logger.Warn("telemetry batcher channel full, redirecting to spillover cache", zap.String("event_id", event.EventID.String()))
		go b.spillover(event)
	}
}

func (b *TelemetryBatcher) spillover(event models.TelemetryEvent) {
	if b.rdb == nil {
		b.logger.Error("failed to spillover telemetry: Redis client not initialized")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	data, err := json.Marshal(event)
	if err != nil {
		b.logger.Error("failed to marshal spillover event", zap.Error(err))
		return
	}

	if err := b.rdb.LPush(ctx, "telemetry:spillover", data).Err(); err != nil {
		b.logger.Error("failed to push telemetry event to Redis spillover", zap.Error(err))
	} else {
		b.logger.Info("telemetry event successfully spilled over to Redis", zap.String("event_id", event.EventID.String()))
	}
}

// ForceFlush triggers an immediate flush of the current buffer.
func (b *TelemetryBatcher) ForceFlush() {
	done := make(chan struct{})
	b.flushSig <- done
	<-done
}

// QueueStats returns the current number of events in the queue and its total capacity.
func (b *TelemetryBatcher) QueueStats() (int, int) {
	return len(b.ch), cap(b.ch)
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
			metricQueueDepth.Dec()
			if len(buffer) >= b.batchSize {
				b.flush(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				b.flush(buffer)
				buffer = buffer[:0]
			}
		case done := <-b.flushSig:
			if len(buffer) > 0 {
				b.flush(buffer)
				buffer = buffer[:0]
			}
			close(done)
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
		metricBatchFlushes.WithLabelValues("error").Inc()
	} else {
		b.logger.Debug("flushed telemetry batch", zap.Int("count", len(events)))
		metricBatchFlushes.WithLabelValues("success").Inc()
		metricBatchSize.Observe(float64(len(events)))
	}
}
