ALTER TABLE inbox_items
DROP CONSTRAINT IF EXISTS uq_inbox_items_tenant_idempotency_key;

ALTER TABLE inbox_items
DROP COLUMN IF EXISTS idempotency_key;
