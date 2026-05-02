-- Migration: 067_monitoring_health
-- Description: Creates tables for health history, alerts, and alert thresholds

-- 1. Create health_history table for tracking health check history
CREATE TABLE IF NOT EXISTS health_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
);

DO $$ 
BEGIN 
    ALTER TABLE health_history ADD COLUMN IF NOT EXISTS timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE health_history ADD COLUMN IF NOT EXISTS component VARCHAR(100) NOT NULL DEFAULT 'unknown';
    ALTER TABLE health_history ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'up';
    ALTER TABLE health_history ADD COLUMN IF NOT EXISTS response_time_ms INTEGER;
    ALTER TABLE health_history ADD COLUMN IF NOT EXISTS error TEXT;
    ALTER TABLE health_history ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_health_history_component ON health_history(component);
CREATE INDEX IF NOT EXISTS idx_health_history_timestamp ON health_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_health_history_component_timestamp ON health_history(component, timestamp DESC);

-- 2. Create health_alerts table for active alerts
CREATE TABLE IF NOT EXISTS health_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
);

DO $$ 
BEGIN 
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS alert_type VARCHAR(50) NOT NULL DEFAULT 'SYSTEM';
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS component VARCHAR(100) NOT NULL DEFAULT 'unknown';
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS metric VARCHAR(100) NOT NULL DEFAULT 'health';
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS current_value DOUBLE PRECISION NOT NULL DEFAULT 0;
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS threshold DOUBLE PRECISION NOT NULL DEFAULT 0;
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS severity VARCHAR(20) NOT NULL DEFAULT 'info';
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS message TEXT NOT NULL DEFAULT '';
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS triggered_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS acknowledged_at TIMESTAMPTZ;
    ALTER TABLE health_alerts ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Indexes for alerts
CREATE INDEX IF NOT EXISTS idx_health_alerts_status ON health_alerts(status);
CREATE INDEX IF NOT EXISTS idx_health_alerts_component ON health_alerts(component);
CREATE INDEX IF NOT EXISTS idx_health_alerts_severity ON health_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_health_alerts_triggered_at ON health_alerts(triggered_at DESC);

-- 3. Create alert_thresholds table for alert configuration
CREATE TABLE IF NOT EXISTS alert_thresholds (
    component VARCHAR(100) NOT NULL,
    metric VARCHAR(100) NOT NULL,
    PRIMARY KEY (component, metric)
);

DO $$ 
BEGIN 
    ALTER TABLE alert_thresholds ADD COLUMN IF NOT EXISTS threshold DOUBLE PRECISION NOT NULL DEFAULT 0;
    ALTER TABLE alert_thresholds ADD COLUMN IF NOT EXISTS operator VARCHAR(10) NOT NULL DEFAULT 'gt';
    ALTER TABLE alert_thresholds ADD COLUMN IF NOT EXISTS severity VARCHAR(20) NOT NULL DEFAULT 'warning';
    ALTER TABLE alert_thresholds ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
    ALTER TABLE alert_thresholds ADD COLUMN IF NOT EXISTS notification_channels TEXT[] DEFAULT '{}';
    ALTER TABLE alert_thresholds ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE alert_thresholds ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Comments
COMMENT ON TABLE health_history IS 'Historical health check results for trend analysis';
COMMENT ON TABLE health_alerts IS 'Active system health alerts';
COMMENT ON TABLE alert_thresholds IS 'Alert threshold configuration';

COMMENT ON COLUMN health_history.status IS 'up=healthy, down=unhealthy, degraded=partial, not_configured=not monitored';
COMMENT ON COLUMN health_alerts.status IS 'active=new alert, acknowledged=being handled, resolved=fixed';
COMMENT ON COLUMN alert_thresholds.operator IS 'gt=greater than, lt=less than, gte=greater or equal, lte=less or equal, eq=equal';
COMMENT ON COLUMN alert_thresholds.notification_channels IS 'Array of notification channels: email, slack, sms';

-- View for recent health status by component
CREATE OR REPLACE VIEW latest_health_status AS
SELECT DISTINCT ON (component)
    component,
    status,
    response_time_ms,
    error,
    timestamp as last_checked
FROM health_history
ORDER BY component, timestamp DESC;

-- View for alert summary
CREATE OR REPLACE VIEW alert_summary AS
SELECT
    severity,
    status,
    COUNT(*) as count,
    MIN(triggered_at) as oldest_alert,
    MAX(triggered_at) as newest_alert
FROM health_alerts
GROUP BY severity, status;

-- Function to cleanup old health history (keep last 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_health_history()
RETURNS void AS $$
BEGIN
    DELETE FROM health_history
    WHERE timestamp < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup resolved alerts (keep last 30 days)
CREATE OR REPLACE FUNCTION cleanup_resolved_alerts()
RETURNS void AS $$
BEGIN
    DELETE FROM health_alerts
    WHERE status = 'resolved'
      AND resolved_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
