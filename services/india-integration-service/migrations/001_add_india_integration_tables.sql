-- Migration to create tables for India integration service
-- This includes tables for FASTag transactions, E-Way Bills, Vehicle/Driver verification,
-- Local permits, GST compliance, and integration configurations

-- Table for FASTag transactions
CREATE TABLE IF NOT EXISTS fastag_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_number VARCHAR(50) NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    transaction_id VARCHAR(100) UNIQUE NOT NULL,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('debit', 'credit')),
    amount DECIMAL(10, 2) NOT NULL,
    toll_plaza VARCHAR(100) NOT NULL,
    toll_plaza_id VARCHAR(50) NOT NULL,
    lane_id VARCHAR(20) NOT NULL,
    transaction_time TIMESTAMP WITH TIME ZONE NOT NULL,
    transaction_date DATE NOT NULL,
    settlement_date DATE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'pending')),
    error_code VARCHAR(20),
    error_description TEXT,
    reconciled BOOLEAN DEFAULT FALSE,
    reconciliation_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fastag_transactions
CREATE INDEX idx_fastag_transactions_vehicle_id ON fastag_transactions(vehicle_id);
CREATE INDEX idx_fastag_transactions_tag_number ON fastag_transactions(tag_number);
CREATE INDEX idx_fastag_transactions_transaction_date ON fastag_transactions(transaction_date);
CREATE INDEX idx_fastag_transactions_status ON fastag_transactions(status);
CREATE INDEX idx_fastag_transactions_organization_id ON fastag_transactions(organization_id);

-- Table for E-Way Bills
CREATE TABLE IF NOT EXISTS e_way_bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    eway_bill_number VARCHAR(50) UNIQUE NOT NULL,
    eway_bill_date TIMESTAMP WITH TIME ZONE NOT NULL,
    eway_bill_valid_till TIMESTAMP WITH TIME ZONE NOT NULL,
    document_number VARCHAR(50) NOT NULL,
    document_date DATE NOT NULL,
    document_type VARCHAR(20) NOT NULL CHECK (document_type IN ('INV', 'CHL', 'WBILL', 'OTH')),
    goods_description TEXT,
    goods_value DECIMAL(12, 2) NOT NULL,
    goods_tax_rate DECIMAL(5, 2),
    goods_unit VARCHAR(20),
    goods_quantity DECIMAL(10, 2),
    from_place VARCHAR(100) NOT NULL,
    from_pincode VARCHAR(10) NOT NULL,
    from_state_code INTEGER NOT NULL,
    to_place VARCHAR(100) NOT NULL,
    to_pincode VARCHAR(10) NOT NULL,
    to_state_code INTEGER NOT NULL,
    transporter_id VARCHAR(50),
    transporter_name VARCHAR(100),
    transport_mode VARCHAR(20) CHECK (transport_mode IN ('road', 'rail', 'air', 'ship')),
    transport_document_number VARCHAR(50),
    transport_document_date VARCHAR(20),
    transit_type VARCHAR(20) CHECK (transit_type IN ('own', 'hired', 'railways')),
    vehicle_number VARCHAR(20),
    vehicle_type VARCHAR(20) CHECK (vehicle_type IN ('reg', 'lr', 'cn')),
    remaining_distance DECIMAL(8, 2),
    user_gstin VARCHAR(15),
    recipient_gstin VARCHAR(15),
    generated_by VARCHAR(100),
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cancelled BOOLEAN DEFAULT FALSE,
    cancelled_by VARCHAR(100),
    cancelled_on TIMESTAMP WITH TIME ZONE,
    cancelled_reason VARCHAR(200),
    updated_by VARCHAR(100),
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for e_way_bills
CREATE INDEX idx_e_way_bills_eway_bill_number ON e_way_bills(eway_bill_number);
CREATE INDEX idx_e_way_bills_document_number ON e_way_bills(document_number);
CREATE INDEX idx_e_way_bills_vehicle_number ON e_way_bills(vehicle_number);
CREATE INDEX idx_e_way_bills_transporter_id ON e_way_bills(transporter_id);
CREATE INDEX idx_e_way_bills_generated_at ON e_way_bills(generated_at);
CREATE INDEX idx_e_way_bills_cancelled ON e_way_bills(cancelled);

-- Table for Vehicle Verification (Vahan integration)
CREATE TABLE IF NOT EXISTS vehicle_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    registration_number VARCHAR(20) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('verified', 'pending', 'failed')),
    verification_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    chassis_number VARCHAR(30),
    engine_number VARCHAR(30),
    model VARCHAR(50),
    manufacturer VARCHAR(100),
    fuel_type VARCHAR(20),
    body_type VARCHAR(50),
    cubic_capacity INTEGER,
    seating_capacity INTEGER,
    color VARCHAR(30),
    manufacturing_year INTEGER,
    owner_name VARCHAR(100),
    owner_address TEXT,
    owner_city VARCHAR(50),
    owner_district VARCHAR(50),
    owner_state VARCHAR(50),
    owner_pin VARCHAR(10),
    rc_status VARCHAR(20) CHECK (rc_status IN ('active', 'inactive', 'suspended')),
    insurance_company VARCHAR(100),
    insurance_policy VARCHAR(50),
    insurance_expiry DATE,
    puc_number VARCHAR(50),
    puc_expiry DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for vehicle_verifications
CREATE INDEX idx_vehicle_verifications_reg_number ON vehicle_verifications(registration_number);
CREATE INDEX idx_vehicle_verifications_vehicle_id ON vehicle_verifications(vehicle_id);
CREATE INDEX idx_vehicle_verifications_status ON vehicle_verifications(status);
CREATE INDEX idx_vehicle_verifications_organization_id ON vehicle_verifications(organization_id);

-- Table for Driver Verification (Sarathi integration)
CREATE TABLE IF NOT EXISTS driver_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    license_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('verified', 'pending', 'failed')),
    verification_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(100) NOT NULL,
    father_name VARCHAR(100),
    date_of_birth DATE,
    gender VARCHAR(10),
    address TEXT,
    city VARCHAR(50),
    district VARCHAR(50),
    state VARCHAR(50),
    pin VARCHAR(10),
    license_class JSONB, -- Store as JSON array of license classes
    issue_date DATE,
    validity_from DATE,
    validity_to DATE,
    last_transaction VARCHAR(100),
    enrollment_number VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for driver_verifications
CREATE INDEX idx_driver_verifications_license_number ON driver_verifications(license_number);
CREATE INDEX idx_driver_verifications_driver_id ON driver_verifications(driver_id);
CREATE INDEX idx_driver_verifications_status ON driver_verifications(status);
CREATE INDEX idx_driver_verifications_organization_id ON driver_verifications(organization_id);

-- Table for Local Permits
CREATE TABLE IF NOT EXISTS local_permits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    permit_number VARCHAR(50) UNIQUE NOT NULL,
    permit_type VARCHAR(50) NOT NULL CHECK (permit_type IN ('stage_carriage', 'contract_carriage', 'goods_vehicle')),
    permit_category VARCHAR(50) NOT NULL CHECK (permit_category IN ('ordinary', 'tourist', 'educational')),
    permit_sub_category VARCHAR(50) CHECK (permit_sub_category IN ('public', 'private')),
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    issuing_authority VARCHAR(100),
    jurisdiction VARCHAR(100),
    route_details TEXT,
    number_of_trips INTEGER,
    max_load_capacity DECIMAL(10, 2),
    number_of_seats INTEGER,
    fare DECIMAL(10, 2),
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'expired', 'cancelled', 'suspended')),
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for local_permits
CREATE INDEX idx_local_permits_permit_number ON local_permits(permit_number);
CREATE INDEX idx_local_permits_vehicle_id ON local_permits(vehicle_id);
CREATE INDEX idx_local_permits_driver_id ON local_permits(driver_id);
CREATE INDEX idx_local_permits_status ON local_permits(status);
CREATE INDEX idx_local_permits_valid_from_to ON local_permits(valid_from, valid_to);
CREATE INDEX idx_local_permits_organization_id ON local_permits(organization_id);

-- Table for GST Compliance
CREATE TABLE IF NOT EXISTS gst_compliances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    gstin VARCHAR(15) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'cancelled', 'suspended')),
    legal_name VARCHAR(200) NOT NULL,
    trade_name VARCHAR(200),
    address TEXT,
    centre_jurisdiction VARCHAR(100),
    state_jurisdiction VARCHAR(100),
    date_of_registration DATE,
    date_of_cancellation DATE,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    return_filing_frequency VARCHAR(20) CHECK (return_filing_frequency IN ('monthly', 'quarterly', 'annually')),
    return_filing_status VARCHAR(20) CHECK (return_filing_status IN ('compliant', 'non_compliant')),
    payment_status VARCHAR(20) CHECK (payment_status IN ('compliant', 'non_compliant')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for gst_compliances
CREATE INDEX idx_gst_compliances_gstin ON gst_compliances(gstin);
CREATE INDEX idx_gst_compliances_organization_id ON gst_compliances(organization_id);
CREATE INDEX idx_gst_compliances_status ON gst_compliances(status);

-- Table for Integration Configurations
CREATE TABLE IF NOT EXISTS integration_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    service_type VARCHAR(50) NOT NULL CHECK (service_type IN ('fastag', 'ewaybill', 'vahan', 'sarathi', 'gst')),
    api_endpoint TEXT NOT NULL,
    api_key VARCHAR(200),
    username VARCHAR(100),
    password VARCHAR(200),
    client_id VARCHAR(100),
    client_secret VARCHAR(200),
    environment VARCHAR(20) DEFAULT 'production' CHECK (environment IN ('production', 'staging', 'sandbox')),
    active BOOLEAN DEFAULT TRUE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    next_sync_at TIMESTAMP WITH TIME ZONE,
    sync_interval INTEGER DEFAULT 60, -- in minutes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for integration_configs
CREATE INDEX idx_integration_configs_org_service ON integration_configs(organization_id, service_type);
CREATE INDEX idx_integration_configs_service_type ON integration_configs(service_type);
CREATE INDEX idx_integration_configs_active ON integration_configs(active);

-- Function to update the 'updated_at' column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update the 'updated_at' column
CREATE TRIGGER update_fastag_transactions_updated_at 
    BEFORE UPDATE ON fastag_transactions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_e_way_bills_updated_at 
    BEFORE UPDATE ON e_way_bills 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_verifications_updated_at 
    BEFORE UPDATE ON vehicle_verifications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_driver_verifications_updated_at 
    BEFORE UPDATE ON driver_verifications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_local_permits_updated_at 
    BEFORE UPDATE ON local_permits 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_gst_compliances_updated_at 
    BEFORE UPDATE ON gst_compliances 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_integration_configs_updated_at 
    BEFORE UPDATE ON integration_configs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();