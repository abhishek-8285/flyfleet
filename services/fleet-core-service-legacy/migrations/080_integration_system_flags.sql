-- +goose Up
CREATE TABLE integration_system_flags (
    flag_key   TEXT PRIMARY KEY,
    flag_value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed initial backpressure flag
INSERT INTO integration_system_flags (flag_key, flag_value) VALUES ('backpressure_paused', 'false');

-- +goose Down
DROP TABLE IF EXISTS integration_system_flags;
