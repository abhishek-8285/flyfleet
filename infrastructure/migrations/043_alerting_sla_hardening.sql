-- migration: 043_alerting_sla_hardening.sql
SET search_path TO public;

-- 1. Add deduplication and escalation tracking to alerts
ALTER TABLE alerts 
ADD COLUMN IF NOT EXISTS dedup_key TEXT,
ADD COLUMN IF NOT EXISTS escalation_level INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_escalated_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS escalated_to UUID[],
ADD COLUMN IF NOT EXISTS sla_deadline TIMESTAMP WITH TIME ZONE;

-- 2. Index for deduplication
-- Only deduplicate 'new' alerts to allow re-triggering if resolved
CREATE INDEX IF NOT EXISTS idx_alerts_dedup_active 
ON alerts (dedup_key) 
WHERE status = 'new';

-- 3. Index for escalation worker
CREATE INDEX IF NOT EXISTS idx_alerts_escalation_worker
ON alerts (status, sla_deadline)
WHERE status = 'new';

-- 4. SLA Metrics Table (for programmatic monitoring)
CREATE TABLE IF NOT EXISTS sla_metrics (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    org_id UUID NOT NULL REFERENCES organizations(id),
    metric_name VARCHAR(50) NOT NULL, -- e.g., 'alert_response_time', 'system_uptime'
    value DECIMAL(15, 4) NOT NULL,
    metadata JSONB
);

SELECT create_hypertable('sla_metrics', 'time', if_not_exists => TRUE);

-- 5. Enable RLS
ALTER TABLE sla_metrics ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS sla_tenant_policy ON sla_metrics;
CREATE POLICY sla_tenant_policy ON sla_metrics
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- 6. Permissions
GRANT SELECT, INSERT ON sla_metrics TO flyfleet_app;
