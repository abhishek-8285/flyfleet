-- Migration: 232_fuel_logs_anomaly_fields
-- Description: Add anomaly tracking fields to fuel_logs

ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS anomaly_flag BOOLEAN DEFAULT FALSE;
ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS anomaly_reason TEXT;

-- Index
CREATE INDEX idx_fuel_logs_anomaly ON fuel_logs(anomaly_flag) WHERE anomaly_flag = TRUE;

-- Comments
COMMENT ON COLUMN fuel_logs.anomaly_flag IS 'Flagged if suspicion of theft or inefficiency detected by anomaly worker';
COMMENT ON COLUMN fuel_logs.anomaly_reason IS 'Reason for flagging as anomalous';
