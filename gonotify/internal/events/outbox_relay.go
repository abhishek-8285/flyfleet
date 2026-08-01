package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/queue"
)

type OutboxRelay struct {
	db       *pgxpool.Pool
	bus      queue.Bus
	logger   *slog.Logger
	interval time.Duration
}

func NewOutboxRelay(
	db *pgxpool.Pool,
	bus queue.Bus,
	logger *slog.Logger,
	interval time.Duration,
) *OutboxRelay {
	return &OutboxRelay{
		db:       db,
		bus:      bus,
		logger:   logger,
		interval: interval,
	}
}

func (r *OutboxRelay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.logger.InfoContext(ctx, "outbox relay started")

	for {
		select {
		case <-ctx.Done():
			r.logger.InfoContext(ctx, "outbox relay stopped")
			return nil
		case <-ticker.C:
			if err := r.poll(ctx); err != nil {
				r.logger.ErrorContext(ctx, "outbox relay poll failed", "error", err)
			}
		}
	}
}

func (r *OutboxRelay) poll(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	rows, err := tx.Query(
		ctx,
		`
		SELECT
			id::text,
			topic,
			payload
		FROM outbox_messages
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT 50
		FOR UPDATE SKIP LOCKED
		`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type outboxRow struct {
		ID      string
		Topic   string
		Payload []byte
	}

	var items []outboxRow

	for rows.Next() {
		var item outboxRow

		if err := rows.Scan(
			&item.ID,
			&item.Topic,
			&item.Payload,
		); err != nil {
			return err
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		var msg queue.Message

		if err := json.Unmarshal(item.Payload, &msg); err != nil {
			r.logger.ErrorContext(
				ctx,
				"failed to unmarshal outbox payload",
				"outbox_id", item.ID,
				"error", err,
			)

			// Mark as published to avoid poison outbox entries.
			// In production, move to DLQ instead.
			if _, err := tx.Exec(
				ctx,
				`
				UPDATE outbox_messages
				SET published_at = now()
				WHERE id = $1
				`,
				item.ID,
			); err != nil {
				return err
			}

			continue
		}

		if msg.ID == "" {
			msg.ID = item.ID
		}

		if msg.OccurredAt.IsZero() {
			msg.OccurredAt = time.Now().UTC()
		}

		if err := r.bus.Publish(ctx, item.Topic, msg); err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			`
			UPDATE outbox_messages
			SET published_at = now()
			WHERE id = $1
			`,
			item.ID,
		); err != nil {
			return err
		}

		r.logger.InfoContext(
			ctx,
			"outbox message published",
			"outbox_id", item.ID,
			"topic", item.Topic,
			"message_id", msg.ID,
		)
	}

	return tx.Commit(ctx)
}
