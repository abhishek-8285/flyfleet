-- Migration: 064_diagnostics_device_alerts.sql
-- Description: Device alerts engine schema

CREATE TABLE IF NOT EXISTS device_alerts (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID REFERENCES devices(id),
    alert_type TEXT,
    severity TEXT CHECK(severity IN ('INFO', 'WARNING', 'CRITICAL')),
    alert_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP
);

CREATE INDEX idx_device_alerts_device_id ON device_alerts(device_id);
CREATE INDEX idx_device_alerts_unresolved ON device_alerts(alert_type) WHERE resolved_at IS NULL;
