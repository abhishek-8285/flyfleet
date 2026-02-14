-- Create fuel log status enum
DO $$ BEGIN
    CREATE TYPE fuel_log_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create fuel logs table
CREATE TABLE IF NOT EXISTS fuel_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    station_name TEXT NOT NULL,
    fuel_type TEXT NOT NULL,
    quantity DECIMAL(12,2) NOT NULL,
    unit_price DECIMAL(10,3) NOT NULL,
    total_cost DECIMAL(12,2) NOT NULL,
    odometer DECIMAL(12,2) NOT NULL,
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status fuel_log_status NOT NULL DEFAULT 'PENDING',
    receipt_url TEXT,
    notes TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for organization and vehicle
CREATE INDEX IF NOT EXISTS idx_fuel_logs_org_vehicle ON fuel_logs(organization_id, vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_status ON fuel_logs(status);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_date ON fuel_logs(transaction_date);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_fuel_logs_updated_at
    BEFORE UPDATE ON fuel_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
