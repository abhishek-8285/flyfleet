-- Create maintenance schedules table
CREATE TABLE IF NOT EXISTS maintenance_schedules (
    id VARCHAR(36) PRIMARY KEY,
    org_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('routine', 'preventive', 'corrective', 'emergency')),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    due_odometer INTEGER,
    due_engine_hours DECIMAL(10,2),
    assigned_to UUID,
    priority VARCHAR(10) NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'critical')) DEFAULT 'medium',
    status VARCHAR(15) NOT NULL CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled')) DEFAULT 'scheduled',
    estimated_cost DECIMAL(10,2),
    actual_cost DECIMAL(10,2),
    parts_required JSONB DEFAULT '[]'::jsonb,
    service_items JSONB DEFAULT '[]'::jsonb,
    notes TEXT,
    reminders JSONB DEFAULT '[]'::jsonb,
    recurring JSONB,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID
);

-- Create maintenance logs table
CREATE TABLE IF NOT EXISTS maintenance_logs (
    id VARCHAR(36) PRIMARY KEY,
    org_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    schedule_id VARCHAR(36),
    type VARCHAR(20) NOT NULL CHECK (type IN ('routine', 'preventive', 'corrective', 'emergency')),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_minutes INTEGER NOT NULL,
    performed_by UUID NOT NULL,
    cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    odometer INTEGER NOT NULL,
    engine_hours DECIMAL(10,2),
    parts_used JSONB DEFAULT '[]'::jsonb,
    service_items JSONB DEFAULT '[]'::jsonb,
    notes TEXT,
    documents TEXT[],
    status VARCHAR(20) NOT NULL CHECK (status IN ('completed', 'partially_completed')) DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_maintenance_schedules_org_id ON maintenance_schedules(org_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_schedules_vehicle_id ON maintenance_schedules(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_schedules_status ON maintenance_schedules(status);
CREATE INDEX IF NOT EXISTS idx_maintenance_schedules_due_date ON maintenance_schedules(due_date);
CREATE INDEX IF NOT EXISTS idx_maintenance_schedules_priority ON maintenance_schedules(priority);

CREATE INDEX IF NOT EXISTS idx_maintenance_logs_org_id ON maintenance_logs(org_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_logs_vehicle_id ON maintenance_logs(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_logs_start_date ON maintenance_logs(start_date);
CREATE INDEX IF NOT EXISTS idx_maintenance_logs_performed_by ON maintenance_logs(performed_by);

-- Create updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_maintenance_schedules_updated_at
    BEFORE UPDATE ON maintenance_schedules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_maintenance_logs_updated_at
    BEFORE UPDATE ON maintenance_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
