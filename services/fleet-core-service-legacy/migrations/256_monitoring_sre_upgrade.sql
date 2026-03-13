-- Migration: 256_monitoring_sre_upgrade.sql
-- Description: SRE-grade observability schema including synthetic checks, SLOs, and auto-remediation.

-- Phase 3: Synthetic Monitoring
CREATE TABLE IF NOT EXISTS synthetic_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL, -- UP, DOWN, DEGRADED
    response_time_ms INTEGER,
    error TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_synthetic_checks_name_time ON synthetic_checks(check_name, timestamp DESC);

-- Phase 4: SLO and Error Budgets
CREATE TABLE IF NOT EXISTS service_slos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    availability_target DECIMAL(5, 2) NOT NULL, -- e.g. 99.90
    latency_target_ms INTEGER,
    error_rate_target DECIMAL(5, 2),
    period_days INTEGER DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_name)
);

CREATE TABLE IF NOT EXISTS error_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL REFERENCES service_slos(service_name) ON DELETE CASCADE,
    allowed_downtime_minutes INTEGER NOT NULL,
    consumed_minutes INTEGER DEFAULT 0,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Phase 5: Alert Grouping
CREATE TABLE IF NOT EXISTS alert_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_key VARCHAR(255) UNIQUE NOT NULL, -- component:metric:severity
    component VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, SUPPRESSED, RESOLVED
    alert_count INTEGER DEFAULT 1,
    first_triggered TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_triggered TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Phase 6: Auto-Remediation
CREATE TABLE IF NOT EXISTS remediation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_key VARCHAR(255) NOT NULL REFERENCES alert_groups(group_key),
    action_type VARCHAR(50) NOT NULL, -- RESTART_REDIS, SCALE_WORKERS, FLUSH_CACHE
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, EXECUTED, FAILED, SUCCESS
    result_metadata JSONB DEFAULT '{}',
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Phase 7: Incident Management
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'DETECTED', -- DETECTED, ACKNOWLEDGED, ESCALATED, RESOLVED
    group_key VARCHAR(255) REFERENCES alert_groups(group_key),
    service_name VARCHAR(255),
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    external_ref_id VARCHAR(255) -- PagerDuty/OpsGenie ID
);

CREATE INDEX idx_incidents_status ON incidents(status);
CREATE INDEX idx_incidents_detected ON incidents(detected_at DESC);
