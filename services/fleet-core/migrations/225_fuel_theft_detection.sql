-- Migration: 225_fuel_theft_detection
-- Description: Create fuel_efficiency_history table for advanced theft detection

CREATE TABLE IF NOT EXISTS fuel_efficiency_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL,
    distance DECIMAL(10, 2) NOT NULL,
    fuel_used DECIMAL(10, 2) NOT NULL,
    km_per_l DECIMAL(10, 2) NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    anomaly_flag BOOLEAN DEFAULT FALSE,
    anomaly_reason TEXT
);

-- Indexes
CREATE INDEX idx_fuel_eff_vehicle ON fuel_efficiency_history(vehicle_id);
CREATE INDEX idx_fuel_eff_date ON fuel_efficiency_history(recorded_at);

-- Comments
COMMENT ON TABLE fuel_efficiency_history IS 'Historical fuel efficiency data for anomaly detection';
