-- Adds a suppress_reason column to notifications so that the reason a
-- notification was suppressed (preference_disabled, suppressed_address,
-- missing_consent) is directly queryable from the notifications table.
ALTER TABLE notifications
ADD COLUMN suppress_reason TEXT;
