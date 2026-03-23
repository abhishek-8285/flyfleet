-- +goose Up
-- Add result_key to analytics_jobs for Gzip compressed payload references
ALTER TABLE analytics_jobs ADD COLUMN IF NOT EXISTS result_key VARCHAR(255);

-- +goose Down
-- Remove result_key from analytics_jobs
ALTER TABLE analytics_jobs DROP COLUMN IF EXISTS result_key;
