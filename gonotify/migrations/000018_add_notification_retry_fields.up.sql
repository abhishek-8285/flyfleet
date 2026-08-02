-- Adds retry-related columns to notifications.
-- attempt_count: incremented by handleSendFailure on each failed delivery attempt.
-- max_attempts: configurable per-notification cap on total send attempts (default 5).
-- next_retry_at: when status = 'retry_pending', the RetryWorker dispatches once this time arrives.
-- retry_policy: 'exponential' (default) or 'fixed' — selects the backoff strategy.
-- last_provider_config_id: traces which provider config was last tried (FK to provider_configs).
ALTER TABLE notifications
ADD COLUMN IF NOT EXISTS attempt_count INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS max_attempts INT NOT NULL DEFAULT 5,
ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS retry_policy TEXT NOT NULL DEFAULT 'exponential',
ADD COLUMN IF NOT EXISTS last_provider_config_id UUID REFERENCES provider_configs(id);

-- Index for the RetryWorker to find due retry notifications efficiently.
CREATE INDEX IF NOT EXISTS idx_notifications_retry_due
ON notifications (next_retry_at ASC)
WHERE status = 'retry_pending';

-- Index for admin dead-letter listing.
CREATE INDEX IF NOT EXISTS idx_notifications_dead_letter
ON notifications (created_at DESC)
WHERE status = 'dead_letter';
