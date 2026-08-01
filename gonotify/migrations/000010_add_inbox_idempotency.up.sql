ALTER TABLE inbox_items
ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

ALTER TABLE inbox_items
DROP CONSTRAINT IF EXISTS uq_inbox_items_tenant_idempotency_key;

ALTER TABLE inbox_items
ADD CONSTRAINT uq_inbox_items_tenant_idempotency_key
UNIQUE (tenant_id, idempotency_key);
