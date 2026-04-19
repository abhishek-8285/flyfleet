-- +goose Up
-- +goose StatementBegin
-- Fixes NULL values in trips table that cause Go sql.Scan failures

-- 1. Ensure text fields have default values and are NOT NULL
UPDATE trips SET title = '' WHERE title IS NULL;
ALTER TABLE trips ALTER COLUMN title SET DEFAULT '';
-- Note: keep it nullable if there's no NOT NULL constraint yet, but set a default

UPDATE trips SET description = '' WHERE description IS NULL;
ALTER TABLE trips ALTER COLUMN description SET DEFAULT '';

UPDATE trips SET customer_name = '' WHERE customer_name IS NULL;
ALTER TABLE trips ALTER COLUMN customer_name SET DEFAULT '';

UPDATE trips SET customer_phone = '' WHERE customer_phone IS NULL;
ALTER TABLE trips ALTER COLUMN customer_phone SET DEFAULT '';

UPDATE trips SET customer_email = '' WHERE customer_email IS NULL;
ALTER TABLE trips ALTER COLUMN customer_email SET DEFAULT '';

UPDATE trips SET cancellation_reason = '' WHERE cancellation_reason IS NULL;
ALTER TABLE trips ALTER COLUMN cancellation_reason SET DEFAULT '';

-- +goose StatementEnd
