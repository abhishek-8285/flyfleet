package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/observability"
	"github.com/gonotify/gonotify/internal/queue"
)

// RetryWorker re-dispatches notifications in "retry_pending" status whose
// next_retry_at time has arrived. It follows the same pattern as
// ScheduledDeliveryWorker: atomically claims due rows with FOR UPDATE SKIP
// LOCKED, flips status to "queued", and dispatches directly through the
// dispatcher to avoid NATS redelivery.
type RetryWorker struct {
	db           *pgxpool.Pool
	dispatcher   *Dispatcher
	logger       *slog.Logger
	pollInterval time.Duration
	batchSize    int
}

// NewRetryWorker creates a RetryWorker.
func NewRetryWorker(
	db *pgxpool.Pool,
	dispatcher *Dispatcher,
	logger *slog.Logger,
) *RetryWorker {
	return &RetryWorker{
		db:           db,
		dispatcher:   dispatcher,
		logger:       logger,
		pollInterval: 5 * time.Second,
		batchSize:    50,
	}
}

// Run starts the retry poller. It blocks until ctx is cancelled.
func (w *RetryWorker) Run(ctx context.Context) error {
	w.logger.InfoContext(
		ctx,
		"retry worker started",
		"poll_interval", w.pollInterval,
	)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "retry worker stopped")

			return ctx.Err()
		case <-ticker.C:
			if err := w.processDueRetries(ctx); err != nil {
				w.logger.ErrorContext(
					ctx,
					"retry tick failed",
					"error", err,
				)
			}
		}
	}
}

// retryNotification represents a retry_pending notification ready for re-dispatch.
type retryNotification struct {
	ID              string
	TenantID        string
	EventID         string
	Channel         string
	Category        string
	ToAddress       *string
	RecipientUserID *string
	Payload         []byte
	NextRetryAt     time.Time
}

// processDueRetries atomically claims due retry_pending notifications,
// flips them to "queued", and invokes HandleEvent on the dispatcher directly
// for each so it can re-attempt delivery.
func (w *RetryWorker) processDueRetries(ctx context.Context) error {
	rows, err := w.db.Query(ctx, `
		WITH due AS (
			SELECT id
			FROM notifications
			WHERE next_retry_at <= now()
			  AND status = 'retry_pending'
			ORDER BY next_retry_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE notifications
		SET status = 'queued', next_retry_at = NULL, updated_at = now()
		FROM due
		WHERE notifications.id = due.id
		RETURNING
			notifications.id,
			notifications.tenant_id,
			notifications.event_id,
			notifications.channel,
			notifications.category,
			notifications.to_address,
			notifications.recipient_user_id,
			notifications.payload,
			notifications.next_retry_at
	`, w.batchSize)
	if err != nil {
		return fmt.Errorf("claim due retries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var n retryNotification

		if err := rows.Scan(
			&n.ID,
			&n.TenantID,
			&n.EventID,
			&n.Channel,
			&n.Category,
			&n.ToAddress,
			&n.RecipientUserID,
			&n.Payload,
			&n.NextRetryAt,
		); err != nil {
			return fmt.Errorf("scan retry notification: %w", err)
		}

		observability.RetryProcessed.WithLabelValues(n.TenantID, n.Channel).Inc()

		msg, ok := reconstructRetryMessage(n)
		if !ok {
			w.logger.ErrorContext(
				ctx,
				"failed to reconstruct retry notification message",
				"notification_id", n.ID,
			)

			continue
		}

		if err := w.dispatcher.HandleEvent(ctx, msg); err != nil {
			observability.RetryFailed.WithLabelValues(n.TenantID, n.Channel).Inc()

			w.logger.ErrorContext(
				ctx,
				"failed to re-dispatch retry notification",
				"notification_id", n.ID,
				"event_id", n.EventID,
				"error", err,
			)

			continue
		}

		w.logger.DebugContext(
			ctx,
			"re-dispatched retry notification",
			"notification_id", n.ID,
			"event_id", n.EventID,
			"channel", n.Channel,
		)
	}

	return rows.Err()
}

// reconstructRetryMessage builds a queue.Message from a retry notification row,
// suitable for passing to HandleEvent. The next_retry_at field is intentionally
// omitted (cleared) so the dispatcher delivers the notification immediately on
// re-dispatch rather than deferring it again.
func reconstructRetryMessage(
	n retryNotification,
) (queue.Message, bool) {
	var data map[string]any
	if err := json.Unmarshal(n.Payload, &data); err != nil {
		return queue.Message{}, false
	}

	eventType, _ := data["event"].(string)
	externalUserID, _ := data["user_id"].(string)

	metadata := map[string]string{
		"event_type": eventType,
		"user_id":    externalUserID,
		"channels":   n.Channel,
	}

	if n.Category != "" {
		metadata["category"] = n.Category
	}

	return queue.Message{
		ID:         n.EventID,
		TenantID:   n.TenantID,
		Type:       "event.received",
		OccurredAt: time.Now().UTC(),
		Payload:    n.Payload,
		Metadata:   metadata,
	}, true
}
