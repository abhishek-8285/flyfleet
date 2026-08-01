-- name: GetEventByIdempotencyKey :one
SELECT
    id::text AS id,
    status
FROM events
WHERE tenant_id = $1
  AND idempotency_key = $2;

-- name: InsertEvent :one
INSERT INTO events (
    tenant_id,
    event_type,
    user_id,
    idempotency_key,
    payload,
    metadata,
    status
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    'received'
)
RETURNING id::text AS id;

-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (
    tenant_id,
    aggregate_type,
    aggregate_id,
    topic,
    payload
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);

-- name: SelectUnpublishedOutboxMessages :many
SELECT
    id::text AS id,
    topic,
    payload
FROM outbox_messages
WHERE published_at IS NULL
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxMessagePublished :exec
UPDATE outbox_messages
SET published_at = now()
WHERE id = $1;
