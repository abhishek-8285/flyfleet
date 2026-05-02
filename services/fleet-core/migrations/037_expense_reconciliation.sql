-- Create fuel_transactions table
CREATE TABLE IF NOT EXISTS fuel_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    vehicle_id UUID NOT NULL, -- Linked to vehicles table (loosely constrained for microservices)
    driver_id UUID,          -- Linked to drivers/users
    card_id VARCHAR(100),    -- External fuel card identifier
    station_location VARCHAR(255),
    liters DECIMAL(10, 2) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    odometer_reading DECIMAL(10, 2),
    efficiency_calculation DECIMAL(10, 2), -- Computed km/l or mpg
    anomaly_flag BOOLEAN DEFAULT FALSE,    -- True if suspicion of theft/inefficiency
    transaction_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create expense_policies table
CREATE TABLE IF NOT EXISTS expense_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    category VARCHAR(50) NOT NULL, -- e.g., 'MEALS', 'LODGING', 'FUEL', 'MAINTENANCE'
    limit_amount DECIMAL(10, 2),   -- Auto-approval limit
    currency VARCHAR(3) DEFAULT 'USD',
    requires_approval BOOLEAN DEFAULT TRUE,
    receipt_required BOOLEAN DEFAULT TRUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, category)
);

-- Create expense_reports table
CREATE TABLE IF NOT EXISTS expense_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    submitter_id UUID NOT NULL, -- Driver or Employee ID
    period_start DATE,
    period_end DATE,
    total_amount DECIMAL(10, 2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, APPROVED, REJECTED, PAID
    approved_by UUID,
    approval_date TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_fuel_transactions_tenant ON fuel_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_fuel_transactions_vehicle ON fuel_transactions(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fuel_transactions_date ON fuel_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_expense_reports_tenant_status ON expense_reports(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_expense_reports_submitter ON expense_reports(submitter_id);
