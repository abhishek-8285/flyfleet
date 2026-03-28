package worker

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/repository"
	"go.uber.org/zap"
)

// OutboxRelay constants
const (
	RelayLockKey     = 1001 // Postgres advisory lock key
	DefaultBatchSize = 100
	CleanupInterval  = 1 * time.Hour
	RetentionPeriodH = 48
)

// OutboxRelay polls the telemetry_outbox and publishes to NATS JetStream.
// Multi-pod safe: Uses SKIP LOCKED for batching and optional Advisory Lock for leader-only cleanup.
type OutboxRelay struct {
	repo          repository.OutboxRepository
	mb            messaging.MessageBroker
	logger        *zap.Logger
	pollTicker    *time.Ticker
	cleanupTicker *time.Ticker
}

func NewOutboxRelay(repo repository.OutboxRepository, mb messaging.MessageBroker, logger *zap.Logger) *OutboxRelay {
	return &OutboxRelay{
		repo:          repo,
		mb:            mb,
		logger:        logger,
		pollTicker:    time.NewTicker(500 * time.Millisecond),
		cleanupTicker: time.NewTicker(CleanupInterval),
	}
}

func (w *OutboxRelay) Start(ctx context.Context) {
	w.logger.Info("Starting Telemetry Outbox Relay")
	defer w.pollTicker.Stop()
	defer w.cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping Telemetry Outbox Relay (Graceful)")
			return
		case <-w.pollTicker.C:
			w.processBatch(ctx)
		case <-w.cleanupTicker.C:
			w.runCleanup(ctx)
		}
	}
}

func (w *OutboxRelay) processBatch(ctx context.Context) {
	msgs, err := w.repo.GetPending(ctx, DefaultBatchSize)
	if err != nil {
		w.logger.Error("failed to fetch pending outbox messages", zap.Error(err))
		return
	}

	if len(msgs) == 0 {
		return
	}

	for _, m := range msgs {
		// Publish to JetStream (telemetry.raw)
		if err := w.mb.PublishJS("telemetry.raw", m.Payload); err != nil {
			w.logger.Warn("failed to relay outbox message", zap.Int64("id", m.ID), zap.Error(err))
			if markErr := w.repo.MarkFailed(ctx, m.ID, err); markErr != nil {
				w.logger.Error("failed to mark outbox as failed", zap.Int64("id", m.ID), zap.Error(markErr))
			}
			continue
		}

		// Mark as processed
		if err := w.repo.MarkProcessed(ctx, m.ID); err != nil {
			w.logger.Error("failed to mark outbox as processed", zap.Int64("id", m.ID), zap.Error(err))
		}
	}
}

func (w *OutboxRelay) runCleanup(ctx context.Context) {
	// Periodic cleanup of processed/failed messages older than 48h
	// We use an advisory lock to ensure only one instance runs cleanup at a time
	locked, err := w.repo.TryLock(ctx, RelayLockKey)
	if err != nil {
		w.logger.Error("failed to acquire cleanup lock", zap.Error(err))
		return
	}
	if !locked {
		return // Another instance is cleaning up
	}
	defer w.repo.Unlock(ctx, RelayLockKey)

	w.logger.Info("Running Outbox Cleanup")
	count, err := w.repo.Cleanup(ctx, RetentionPeriodH)
	if err != nil {
		w.logger.Error("failed to cleanup outbox", zap.Error(err))
		return
	}
	if count > 0 {
		w.logger.Info("Outbox cleanup complete", zap.Int64("removed_count", count))
	}
}
