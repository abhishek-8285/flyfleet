-- +goose Up
CREATE TABLE IF NOT EXISTS integration_system_flags (
    flag_key   TEXT PRIMARY KEY,
    flag_value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed initial backpressure flag
INSERT INTO integration_system_flags (flag_key, flag_value) 
VALUES ('backpressure_paused', 'false')
ON CONFLICT (flag_key) DO NOTHING;

-- All tables and indexes handled defensively above.
