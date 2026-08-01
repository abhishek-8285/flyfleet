ALTER TABLE notifications
ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

ALTER TABLE notifications
DROP CONSTRAINT IF EXISTS uq_notifications_tenant_idempotency_key;

ALTER TABLE notifications
ADD CONSTRAINT uq_notifications_tenant_idempotency_key
UNIQUE (tenant_id, idempotency_key);
