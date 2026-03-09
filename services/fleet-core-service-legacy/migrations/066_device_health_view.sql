-- Migration: 066_device_health_view.sql
-- Description: Aggregate views for device health and auth tokens for devices

-- Device Authentication 
CREATE TABLE IF NOT EXISTS device_auth_tokens (
    device_id UUID PRIMARY KEY REFERENCES devices(id),
    device_token TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Device Fleet Health View
CREATE OR REPLACE VIEW device_health_summary AS
SELECT
    device_id,
    MAX(timestamp) AS last_seen,
    AVG(battery_voltage) AS avg_battery,
    COUNT(*) AS telemetry_events
FROM device_telemetry
GROUP BY device_id;
