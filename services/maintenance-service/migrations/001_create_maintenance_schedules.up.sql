-- 001_create_maintenance_schedules.up.sql

-- Standard update_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Maintenance Schedules Table
CREATE TABLE IF NOT EXISTS maintenance_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('preventive', 'corrective', 'inspection')),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ NOT NULL,
    due_odometer INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'in_progress', 'completed', 'overdue')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    assigned_to UUID,
    estimated_cost DECIMAL(10,2) DEFAULT 0.00,
    actual_cost DECIMAL(10,2),
    parts_required JSONB DEFAULT '[]'::jsonb,
    notes TEXT,
    completed_at TIMESTAMPTZ,
    completed_by UUID,
    
    -- Mandatory Standard Fields
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMPTZ
);

-- Mandatory Indexes
CREATE INDEX idx_maintenance_schedules_org_id ON maintenance_schedules(org_id) WHERE is_deleted = false;
CREATE INDEX idx_maintenance_schedules_status ON maintenance_schedules(status) WHERE is_deleted = false;
CREATE INDEX idx_maintenance_schedules_vehicle_id ON maintenance_schedules(vehicle_id) WHERE is_deleted = false;

-- Mandatory Trigger
CREATE TRIGGER update_maintenance_schedules_updated_at
    BEFORE UPDATE ON maintenance_schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
soul
