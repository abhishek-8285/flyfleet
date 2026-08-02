package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/queue"
)

// ScheduledDeliveryWorker re-dispatches notifications whose scheduled_at time
// has arrived. It flips due "scheduled" notifications to "queued" and invokes
// the dispatcher's HandleEvent directly, reusing the existing idempotent
// delivery flow to perform the actual delivery. Re-publishing through NATS is
// avoided because JetStream deduplicates messages by msg.ID (event_id), which
// would silently drop the re-dispatched message.
type ScheduledDeliveryWorker struct {
	db           *pgxpool.Pool
	dispatcher   *Dispatcher
	logger       *slog.Logger
	pollInterval time.Duration
	batchSize    int
}

// NewScheduledDeliveryWorker creates a ScheduledDeliveryWorker.
func NewScheduledDeliveryWorker(
	db *pgxpool.Pool,
	dispatcher *Dispatcher,
	logger *slog.Logger,
) *ScheduledDeliveryWorker {
	return &ScheduledDeliveryWorker{
		db:           db,
		dispatcher:   dispatcher,
		logger:       logger,
		pollInterval: 5 * time.Second,
		batchSize:    50,
	}
}

// Run starts the scheduled-delivery poller. It blocks until ctx is cancelled.
func (w *ScheduledDeliveryWorker) Run(ctx context.Context) error {
	w.logger.InfoContext(
		ctx,
		"scheduled delivery worker started",
		"poll_interval", w.pollInterval,
	)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "scheduled delivery worker stopped")

			return ctx.Err()
		case <-ticker.C:
			if err := w.processDueNotifications(ctx); err != nil {
				w.logger.ErrorContext(
					ctx,
					"scheduled delivery tick failed",
					"error", err,
				)
			}
		}
	}
}

// scheduledNotification represents a notification row that is ready for re-dispatch.
type scheduledNotification struct {
	ID              string
	TenantID        string
	EventID         string
	Channel         string
	Category        string
	ToAddress       *string
	RecipientUserID *string
	Payload         []byte
	ScheduledAt     time.Time
}

// processDueNotifications atomically claims due "scheduled" notifications,
// flips them to "queued", and invokes HandleEvent on the dispatcher directly
// for each so it can deliver them.
func (w *ScheduledDeliveryWorker) processDueNotifications(ctx context.Context) error {
	rows, err := w.db.Query(ctx, `
		WITH due AS (
			SELECT id
			FROM notifications
			WHERE scheduled_at <= now()
			  AND status = 'scheduled'
			ORDER BY scheduled_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE notifications
		SET status = 'queued', updated_at = now()
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
			notifications.scheduled_at
	`, w.batchSize)
	if err != nil {
		return fmt.Errorf("claim due notifications: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var n scheduledNotification

		if err := rows.Scan(
			&n.ID,
			&n.TenantID,
			&n.EventID,
			&n.Channel,
			&n.Category,
			&n.ToAddress,
			&n.RecipientUserID,
			&n.Payload,
			&n.ScheduledAt,
		); err != nil {
			return fmt.Errorf("scan notification: %w", err)
		}

		msg, ok := w.reconstructMessage(n)
		if !ok {
			w.logger.ErrorContext(
				ctx,
				"failed to reconstruct scheduled notification message",
				"notification_id", n.ID,
			)

			continue
		}

		if err := w.dispatcher.HandleEvent(ctx, msg); err != nil {
			w.logger.ErrorContext(
				ctx,
				"failed to re-dispatch scheduled notification",
				"notification_id", n.ID,
				"event_id", n.EventID,
				"error", err,
			)

			continue
		}

		w.logger.DebugContext(
			ctx,
			"re-dispatched scheduled notification",
			"notification_id", n.ID,
			"event_id", n.EventID,
			"channel", n.Channel,
		)
	}

	return rows.Err()
}

// reconstructMessage builds a queue.Message from a scheduled notification row,
// suitable for passing to HandleEvent. The scheduled_at field is intentionally
// omitted (cleared) so the dispatcher delivers the notification immediately on
// re-dispatch rather than deferring it again.
func (w *ScheduledDeliveryWorker) reconstructMessage(
	n scheduledNotification,
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
