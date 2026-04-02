-- Missing Tables Stubs for Governance and India Integration
-- This migration provides the missing schema for workers and repositories

-- 1. Governance Settings
CREATE TABLE IF NOT EXISTS governance_settings (
    org_id UUID PRIMARY KEY REFERENCES organizations(id),
    retention_policies JSONB DEFAULT '{}',
    masking_rules JSONB DEFAULT '{}',
    auto_expiry_enabled BOOLEAN DEFAULT FALSE,
    download_ttl_hours INTEGER DEFAULT 24,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- 2. Vehicle Verifications
CREATE TABLE IF NOT EXISTS vehicle_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id),
    org_id UUID NOT NULL REFERENCES organizations(id),
    registration_number VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    verification_date TIMESTAMP WITH TIME ZONE,
    chassis_number VARCHAR(50),
    engine_number VARCHAR(50),
    model VARCHAR(100),
    manufacturer VARCHAR(100),
    fuel_type VARCHAR(20),
    body_type VARCHAR(50),
    cubic_capacity INTEGER,
    seating_capacity INTEGER,
    color VARCHAR(30),
    manufacturing_year INTEGER,
    owner_name VARCHAR(255),
    owner_address TEXT,
    owner_city VARCHAR(100),
    owner_district VARCHAR(100),
    owner_state VARCHAR(100),
    owner_pin VARCHAR(10),
    rc_status VARCHAR(20),
    insurance_company VARCHAR(100),
    insurance_policy VARCHAR(100),
    insurance_expiry TIMESTAMP WITH TIME ZONE,
    puc_number VARCHAR(100),
    puc_expiry TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- 3. Driver Verifications
CREATE TABLE IF NOT EXISTS driver_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id),
    org_id UUID NOT NULL REFERENCES organizations(id),
    license_number VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    verification_date TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    father_name VARCHAR(255),
    date_of_birth TIMESTAMP WITH TIME ZONE NOT NULL,
    gender VARCHAR(10),
    address TEXT,
    city VARCHAR(100),
    district VARCHAR(100),
    state VARCHAR(100),
    pin VARCHAR(10),
    license_class TEXT[],
    issue_date TIMESTAMP WITH TIME ZONE NOT NULL,
    validity_from TIMESTAMP WITH TIME ZONE NOT NULL,
    validity_to TIMESTAMP WITH TIME ZONE NOT NULL,
    last_transaction VARCHAR(255),
    enrollment_number VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- 4. Local Permits
CREATE TABLE IF NOT EXISTS local_permits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id),
    driver_id UUID REFERENCES drivers(id),
    org_id UUID NOT NULL REFERENCES organizations(id),
    permit_number VARCHAR(50) NOT NULL,
    permit_type VARCHAR(50) NOT NULL,
    permit_category VARCHAR(50),
    permit_sub_category VARCHAR(50),
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to TIMESTAMP WITH TIME ZONE NOT NULL,
    issuing_authority VARCHAR(255) NOT NULL,
    jurisdiction VARCHAR(255),
    route_details TEXT,
    number_of_trips INTEGER,
    max_load_capacity FLOAT,
    number_of_seats INTEGER,
    fare FLOAT,
    status VARCHAR(20) NOT NULL,
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- 5. GST Compliances
CREATE TABLE IF NOT EXISTS gst_compliances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    gstin VARCHAR(15) NOT NULL,
    status VARCHAR(20) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    trade_name VARCHAR(255),
    address TEXT NOT NULL,
    centre_jurisdiction VARCHAR(100),
    state_jurisdiction VARCHAR(100),
    date_of_registration TIMESTAMP WITH TIME ZONE NOT NULL,
    date_of_cancellation TIMESTAMP WITH TIME ZONE,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL,
    return_filing_frequency VARCHAR(20),
    return_filing_status VARCHAR(20),
    payment_status VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- 6. Integration Configs
CREATE TABLE IF NOT EXISTS integration_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    service_type VARCHAR(50) NOT NULL,
    api_endpoint TEXT NOT NULL,
    api_key TEXT,
    username TEXT,
    password TEXT,
    client_id TEXT,
    client_secret TEXT,
    environment VARCHAR(20) NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    next_sync_at TIMESTAMP WITH TIME ZONE,
    sync_interval INTEGER DEFAULT 3600,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);
