-- +goose Up
-- +goose StatementBegin

-- 1. Create event edit history table for audit trails
CREATE TABLE IF NOT EXISTS event_edit_history (
    id UUID PRIMARY KEY,
    original_event_id UUID NOT NULL,
    edited_event_id UUID NOT NULL,
    edit_reason TEXT NOT NULL,
    edited_by UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 2. Update unidentified driver events with missing regulatory fields
ALTER TABLE unidentified_driver_events ADD COLUMN IF NOT EXISTS engine_hours DOUBLE PRECISION;
ALTER TABLE unidentified_driver_events ADD COLUMN IF NOT EXISTS resolved BOOLEAN DEFAULT false;

-- 3. Add integrity fields to duty status logs
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS checksum TEXT;
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS sequence_id BIGINT;
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS origin VARCHAR(100);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE duty_status_logs DROP COLUMN IF EXISTS origin;
ALTER TABLE duty_status_logs DROP COLUMN IF EXISTS sequence_id;
ALTER TABLE duty_status_logs DROP COLUMN IF EXISTS checksum;

ALTER TABLE unidentified_driver_events DROP COLUMN IF EXISTS resolved;
ALTER TABLE unidentified_driver_events DROP COLUMN IF EXISTS engine_hours;

DROP TABLE IF EXISTS event_edit_history;

-- +goose StatementEnd
