-- Index to support the scheduled-delivery worker polling for due notifications.
CREATE INDEX IF NOT EXISTS idx_notifications_scheduled
ON notifications (scheduled_at)
WHERE status = 'scheduled';

-- Index to support quiet-hours lookups. The quiet_hours table is keyed by
-- user_id (primary key), but this index documents and guarantees the
-- lookup pattern used by the preference repository.
CREATE INDEX IF NOT EXISTS idx_quiet_hours_user_id
ON quiet_hours (user_id);
