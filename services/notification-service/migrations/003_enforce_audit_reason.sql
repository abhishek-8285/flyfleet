-- Up
ALTER TABLE notification_preference_audit ALTER COLUMN change_reason SET NOT NULL;

-- Down
ALTER TABLE notification_preference_audit ALTER COLUMN change_reason DROP NOT NULL;
