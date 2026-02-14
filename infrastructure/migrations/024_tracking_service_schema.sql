-- Migration to add telemetry analytics tables
-- This adds advanced analytics capabilities to the tracking service

-- Table for storing analytics results
CREATE TABLE IF NOT EXISTS telemetry_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL,
    analysis_type VARCHAR(100) NOT NULL,
    data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confidence DECIMAL(3,2) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    CONSTRAINT fk_telemetry_analytics_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
);

-- Individual indexes
CREATE INDEX IF NOT EXISTS idx_telemetry_analytics_vehicle_id ON telemetry_analytics (vehicle_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_analytics_type ON telemetry_analytics (analysis_type);
CREATE INDEX IF NOT EXISTS idx_telemetry_analytics_timestamp ON telemetry_analytics (timestamp);

-- Table for storing anomaly detection results
CREATE TABLE IF NOT EXISTS telemetry_anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL,
    anomaly_type VARCHAR(100) NOT NULL,
    data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confidence DECIMAL(3,2) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    anomaly_score DECIMAL(5,2) NOT NULL,
    recommendation TEXT,
    CONSTRAINT fk_telemetry_anomalies_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
);

-- Individual indexes
CREATE INDEX IF NOT EXISTS idx_telemetry_anomalies_vehicle_id ON telemetry_anomalies (vehicle_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_anomalies_type ON telemetry_anomalies (anomaly_type);
CREATE INDEX IF NOT EXISTS idx_telemetry_anomalies_timestamp ON telemetry_anomalies (timestamp);
CREATE INDEX IF NOT EXISTS idx_telemetry_anomalies_score ON telemetry_anomalies (anomaly_score DESC);

-- Create hypertable for time-series analytics data
SELECT create_hypertable('telemetry_analytics', 'timestamp', if_not_exists => TRUE);
SELECT create_hypertable('telemetry_anomalies', 'timestamp', if_not_exists => TRUE);

-- Indexes for better query performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_analytics_composite ON telemetry_analytics (vehicle_id, analysis_type, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_anomalies_composite ON telemetry_anomalies (vehicle_id, anomaly_type, timestamp DESC);

-- Add comments to describe the tables
COMMENT ON TABLE telemetry_analytics IS 'Stores advanced analytics results for vehicle telemetry data';
COMMENT ON COLUMN telemetry_analytics.analysis_type IS 'Type of analysis performed (e.g., speed_analysis, fuel_efficiency)';
COMMENT ON COLUMN telemetry_analytics.confidence IS 'Confidence score for the analysis results (0-1)';
COMMENT ON TABLE telemetry_anomalies IS 'Stores detected anomalies in vehicle telemetry data';
COMMENT ON COLUMN telemetry_anomalies.anomaly_type IS 'Type of anomaly detected (e.g., unusual_speed, fuel_theft)';
COMMENT ON COLUMN telemetry_anomalies.anomaly_score IS 'Numerical score representing severity of the anomaly';
COMMENT ON COLUMN telemetry_anomalies.recommendation IS 'Suggested action for addressing the anomaly';