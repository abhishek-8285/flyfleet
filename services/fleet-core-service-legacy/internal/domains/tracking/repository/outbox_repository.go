package repository

import (
	"context"
	"encoding/json"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// OutboxRepository defines the contract for management of the transactional outbox.
type OutboxRepository interface {
	AddToOutbox(ctx context.Context, tx pgx.Tx, event *models.TelemetryEvent) error
	GetPending(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkProcessed(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, err error) error
	Cleanup(ctx context.Context, retentionHours int) (int64, error)
	TryLock(ctx context.Context, lockKey int64) (bool, error)
	Unlock(ctx context.Context, lockKey int64) error
}

type OutboxMessage struct {
	ID      int64
	Payload models.TelemetryEvent
}

type PostgresOutboxRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewOutboxRepository(db *pgxpool.Pool, logger *zap.Logger) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db, logger: logger}
}

// AddToOutbox inserts a telemetry event into the outbox within the existing transaction.
func (r *PostgresOutboxRepository) AddToOutbox(ctx context.Context, tx pgx.Tx, event *models.TelemetryEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO telemetry_outbox (partition_id, payload, status)
		VALUES ($1, $2, 'pending')`

	// Partition ID could be derived from recorded_at month
	partitionID := int(event.RecordedAt.Month())

	_, err = tx.Exec(ctx, q, partitionID, payload)
	return err
}

// GetPending fetches batch of pending messages for the relay.
func (r *PostgresOutboxRepository) GetPending(ctx context.Context, limit int) ([]OutboxMessage, error) {
	const q = `
		SELECT id, payload FROM telemetry_outbox
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []OutboxMessage
	for rows.Next() {
		var m OutboxMessage
		var payload []byte
		if err := rows.Scan(&m.ID, &payload); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payload, &m.Payload); err != nil {
			r.logger.Error("failed to unmarshal outbox payload", zap.Int64("id", m.ID), zap.Error(err))
			continue
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (r *PostgresOutboxRepository) MarkProcessed(ctx context.Context, id int64) error {
	const q = `UPDATE telemetry_outbox SET status = 'processed', last_attempt = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *PostgresOutboxRepository) MarkFailed(ctx context.Context, id int64, relayErr error) error {
	const q = `
		UPDATE telemetry_outbox
		SET attempts = attempts + 1, last_attempt = NOW(), error_msg = $2,
		    status = CASE WHEN attempts >= 5 THEN 'failed' ELSE 'pending' END
		WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id, relayErr.Error())
	return err
}

func (r *PostgresOutboxRepository) Cleanup(ctx context.Context, retentionHours int) (int64, error) {
	const q = `DELETE FROM telemetry_outbox WHERE (status = 'processed' OR status = 'failed') AND created_at < NOW() - $1 * INTERVAL '1 hour'`
	tag, err := r.db.Exec(ctx, q, retentionHours)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *PostgresOutboxRepository) TryLock(ctx context.Context, lockKey int64) (bool, error) {
	var locked bool
	err := r.db.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&locked)
	return locked, err
}

func (r *PostgresOutboxRepository) Unlock(ctx context.Context, lockKey int64) error {
	_, err := r.db.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockKey)
	return err
}
