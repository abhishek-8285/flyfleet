-- Migration: 241_fraud_detection_fields
-- Description: Add fraud scoring fields to fuel and finance tables

ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS fraud_score DECIMAL(5, 2) DEFAULT 0;
ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS risk_level VARCHAR(10) DEFAULT 'LOW';

ALTER TABLE fuel_transactions ADD COLUMN IF NOT EXISTS fraud_score DECIMAL(5, 2) DEFAULT 0;
ALTER TABLE fuel_transactions ADD COLUMN IF NOT EXISTS risk_level VARCHAR(10) DEFAULT 'LOW';

-- Fraud incidents table
CREATE TABLE IF NOT EXISTS fraud_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(30) NOT NULL,               -- fuel_log, fuel_transaction
    entity_id UUID NOT NULL,
    org_id UUID NOT NULL,
    vehicle_id UUID,
    driver_id UUID,
    fraud_score DECIMAL(5, 2) NOT NULL,
    risk_level VARCHAR(10) NOT NULL,                -- LOW, MEDIUM, HIGH, CRITICAL
    signals JSONB DEFAULT '[]',                     -- array of triggered detection signals
    status VARCHAR(20) DEFAULT 'OPEN',              -- OPEN, INVESTIGATING, RESOLVED, FALSE_POSITIVE
    investigated_by UUID,
    resolved_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_fraud_incidents_org ON fraud_incidents(org_id);
CREATE INDEX idx_fraud_incidents_entity ON fraud_incidents(entity_type, entity_id);
CREATE INDEX idx_fraud_incidents_risk ON fraud_incidents(risk_level, status);
CREATE INDEX idx_fraud_incidents_vehicle ON fraud_incidents(vehicle_id) WHERE vehicle_id IS NOT NULL;

COMMENT ON TABLE fraud_incidents IS 'Fraud detected incidents from the risk scoring engine';
