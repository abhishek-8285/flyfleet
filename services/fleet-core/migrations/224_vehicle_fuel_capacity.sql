-- Migration: 224_vehicle_fuel_capacity
-- Description: Create vehicle_specs table for fuel capacity validation

CREATE TABLE IF NOT EXISTS vehicle_specs (
    vehicle_id UUID PRIMARY KEY, -- References vehicles table
    fuel_type VARCHAR(50),
    tank_capacity_liters DECIMAL(10, 2) NOT NULL,
    efficiency_expected_km_l DECIMAL(10, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Comments
COMMENT ON TABLE vehicle_specs IS 'Technical specifications for vehicles used for validation';
COMMENT ON COLUMN vehicle_specs.tank_capacity_liters IS 'Maximum fuel tank capacity in liters';
