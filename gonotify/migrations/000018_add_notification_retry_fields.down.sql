DROP INDEX IF EXISTS idx_notifications_dead_letter;
DROP INDEX IF EXISTS idx_notifications_retry_due;

ALTER TABLE notifications
DROP COLUMN IF EXISTS last_provider_config_id,
DROP COLUMN IF EXISTS retry_policy,
DROP COLUMN IF EXISTS next_retry_at,
DROP COLUMN IF EXISTS max_attempts,
DROP COLUMN IF EXISTS attempt_count;
