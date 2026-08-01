ALTER TABLE notifications
DROP CONSTRAINT IF EXISTS uq_notifications_tenant_idempotency_key;

ALTER TABLE notifications
DROP COLUMN IF EXISTS idempotency_key;
