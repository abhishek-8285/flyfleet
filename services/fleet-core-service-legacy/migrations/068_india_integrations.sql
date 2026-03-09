-- Migration: 068_india_integrations.sql
-- Description: Creates tables for India-specific integrations (FASTag, E-Way Bill, VAHAN, SARATHI, GST, Permits)
-- Created: 2026-03-03

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 1. FASTag Transactions Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS fastag_transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_number          VARCHAR(50) NOT NULL,
    vehicle_id          UUID NOT NULL,
    driver_id           UUID NOT NULL,
    organization_id     UUID NOT NULL,
    transaction_id      VARCHAR(100) NOT NULL UNIQUE,
    transaction_type    VARCHAR(20) NOT NULL CHECK (transaction_type IN ('debit', 'credit')),
    amount              DECIMAL(10, 2) NOT NULL CHECK (amount >= 0),
    toll_plaza          VARCHAR(200) NOT NULL,
    toll_plaza_id       VARCHAR(50) NOT NULL,
    lane_id             VARCHAR(50) NOT NULL,
    transaction_time    TIMESTAMPTZ NOT NULL,
    transaction_date    DATE NOT NULL,
    settlement_date     DATE,
    status              VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'pending')),
    error_code          VARCHAR(50),
    error_description   TEXT,
    reconciled          BOOLEAN NOT NULL DEFAULT FALSE,
    reconciliation_date TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by          UUID NOT NULL
);

-- Indexes for FASTag transactions
CREATE INDEX IF NOT EXISTS idx_fastag_tag_number ON fastag_transactions(tag_number);
CREATE INDEX IF NOT EXISTS idx_fastag_vehicle_id ON fastag_transactions(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fastag_organization_id ON fastag_transactions(organization_id);
CREATE INDEX IF NOT EXISTS idx_fastag_transaction_date ON fastag_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_fastag_transaction_time ON fastag_transactions(transaction_time DESC);
CREATE INDEX IF NOT EXISTS idx_fastag_status ON fastag_transactions(status);
CREATE INDEX IF NOT EXISTS idx_fastag_reconciled ON fastag_transactions(reconciled);

-- ============================================================================
-- 2. E-Way Bills Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS e_way_bills (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    eway_bill_number        VARCHAR(20) NOT NULL UNIQUE,
    eway_bill_date          TIMESTAMPTZ NOT NULL,
    eway_bill_valid_till    TIMESTAMPTZ NOT NULL,
    document_number         VARCHAR(100) NOT NULL,
    document_date           TIMESTAMPTZ NOT NULL,
    document_type           VARCHAR(20) NOT NULL, -- INV, CHL, WBILL, etc.
    goods_description       TEXT NOT NULL,
    goods_value             DECIMAL(15, 2) NOT NULL CHECK (goods_value > 0),
    goods_tax_rate          DECIMAL(5, 2) NOT NULL CHECK (goods_tax_rate >= 0),
    goods_unit              VARCHAR(50) NOT NULL,
    goods_quantity          DECIMAL(15, 3) NOT NULL CHECK (goods_quantity > 0),
    from_place              VARCHAR(200) NOT NULL,
    from_pincode            VARCHAR(10) NOT NULL,
    from_state_code         INTEGER NOT NULL,
    to_place                VARCHAR(200) NOT NULL,
    to_pincode              VARCHAR(10) NOT NULL,
    to_state_code           INTEGER NOT NULL,
    transporter_id          VARCHAR(50),
    transporter_name        VARCHAR(200),
    transport_mode          VARCHAR(20) CHECK (transport_mode IN ('road', 'rail', 'air', 'ship')),
    transport_document_number VARCHAR(100),
    transport_document_date DATE,
    transit_type            VARCHAR(20) CHECK (transit_type IN ('own', 'hired', 'railways')),
    vehicle_number          VARCHAR(20),
    vehicle_type            VARCHAR(10) CHECK (vehicle_type IN ('reg', 'lr', 'cn')),
    remaining_distance      DECIMAL(10, 2),
    user_gstin              VARCHAR(20) NOT NULL,
    recipient_gstin         VARCHAR(20),
    generated_by            VARCHAR(100) NOT NULL,
    generated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled               BOOLEAN NOT NULL DEFAULT FALSE,
    cancelled_by            VARCHAR(100),
    cancelled_on            TIMESTAMPTZ,
    cancelled_reason        TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID NOT NULL,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by              UUID NOT NULL
);

-- Indexes for E-Way Bills
CREATE INDEX IF NOT EXISTS idx_eway_bill_number ON e_way_bills(eway_bill_number);
CREATE INDEX IF NOT EXISTS idx_eway_vehicle_number ON e_way_bills(vehicle_number);
CREATE INDEX IF NOT EXISTS idx_eway_transporter ON e_way_bills(transporter_id);
CREATE INDEX IF NOT EXISTS idx_eway_generated_at ON e_way_bills(generated_at DESC);
CREATE INDEX IF NOT EXISTS idx_eway_valid_till ON e_way_bills(eway_bill_valid_till);
CREATE INDEX IF NOT EXISTS idx_eway_cancelled ON e_way_bills(cancelled);
CREATE INDEX IF NOT EXISTS idx_eway_gstin ON e_way_bills(user_gstin);

-- ============================================================================
-- 3. Vehicle Verifications (VAHAN) Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS vehicle_verifications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id          UUID NOT NULL,
    organization_id     UUID NOT NULL,
    registration_number VARCHAR(20) NOT NULL,
    status              VARCHAR(20) NOT NULL CHECK (status IN ('verified', 'pending', 'failed')),
    verification_date   TIMESTAMPTZ NOT NULL,
    chassis_number      VARCHAR(50),
    engine_number       VARCHAR(50),
    model               VARCHAR(100),
    manufacturer        VARCHAR(100),
    fuel_type           VARCHAR(30),
    body_type           VARCHAR(50),
    cubic_capacity      INTEGER,
    seating_capacity    INTEGER,
    color               VARCHAR(30),
    manufacturing_year  INTEGER,
    owner_name          VARCHAR(200),
    owner_address       TEXT,
    owner_city          VARCHAR(100),
    owner_district      VARCHAR(100),
    owner_state         VARCHAR(100),
    owner_pin           VARCHAR(10),
    rc_status           VARCHAR(20) CHECK (rc_status IN ('active', 'inactive', 'suspended')),
    insurance_company   VARCHAR(100),
    insurance_policy    VARCHAR(100),
    insurance_expiry    TIMESTAMPTZ,
    puc_number          VARCHAR(50),
    puc_expiry          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by          UUID NOT NULL,

    CONSTRAINT unique_registration UNIQUE (registration_number)
);

-- Indexes for Vehicle Verifications
CREATE INDEX IF NOT EXISTS idx_vehicle_reg_number ON vehicle_verifications(registration_number);
CREATE INDEX IF NOT EXISTS idx_vehicle_vehicle_id ON vehicle_verifications(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_organization_id ON vehicle_verifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_status ON vehicle_verifications(status);
CREATE INDEX IF NOT EXISTS idx_vehicle_rc_status ON vehicle_verifications(rc_status);

-- ============================================================================
-- 4. Driver Verifications (SARATHI) Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS driver_verifications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id           UUID NOT NULL,
    organization_id     UUID NOT NULL,
    license_number      VARCHAR(50) NOT NULL,
    status              VARCHAR(20) NOT NULL CHECK (status IN ('verified', 'pending', 'failed')),
    verification_date   TIMESTAMPTZ NOT NULL,
    name                VARCHAR(200) NOT NULL,
    father_name         VARCHAR(200),
    date_of_birth       DATE NOT NULL,
    gender              VARCHAR(10),
    address             TEXT,
    city                VARCHAR(100),
    district            VARCHAR(100),
    state               VARCHAR(100),
    pin                 VARCHAR(10),
    license_class       TEXT[], -- MCWG, LMV, etc.
    issue_date          DATE NOT NULL,
    validity_from       DATE NOT NULL,
    validity_to         DATE NOT NULL,
    last_transaction    VARCHAR(100),
    enrollment_number   VARCHAR(100),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by          UUID NOT NULL,

    CONSTRAINT unique_license_number UNIQUE (license_number)
);

-- Indexes for Driver Verifications
CREATE INDEX IF NOT EXISTS idx_driver_license_number ON driver_verifications(license_number);
CREATE INDEX IF NOT EXISTS idx_driver_driver_id ON driver_verifications(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_organization_id ON driver_verifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_driver_status ON driver_verifications(status);
CREATE INDEX IF NOT EXISTS idx_driver_validity_to ON driver_verifications(validity_to);

-- ============================================================================
-- 5. Local Permits Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS local_permits (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id          UUID NOT NULL,
    driver_id           UUID NOT NULL,
    organization_id     UUID NOT NULL,
    permit_number       VARCHAR(50) NOT NULL UNIQUE,
    permit_type         VARCHAR(30) NOT NULL CHECK (permit_type IN ('stage_carriage', 'contract_carriage', 'goods_vehicle')),
    permit_category     VARCHAR(30), -- ordinary, tourist, educational
    permit_sub_category VARCHAR(30), -- public, private
    valid_from          TIMESTAMPTZ NOT NULL,
    valid_to            TIMESTAMPTZ NOT NULL,
    issuing_authority   VARCHAR(200) NOT NULL,
    jurisdiction        VARCHAR(200),
    route_details       TEXT,
    number_of_trips     INTEGER CHECK (number_of_trips > 0),
    max_load_capacity   DECIMAL(10, 2) CHECK (max_load_capacity > 0),
    number_of_seats     INTEGER CHECK (number_of_seats > 0),
    fare                DECIMAL(10, 2) CHECK (fare >= 0),
    status              VARCHAR(20) NOT NULL CHECK (status IN ('active', 'expired', 'cancelled', 'suspended')),
    remarks             TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by          UUID NOT NULL
);

-- Indexes for Local Permits
CREATE INDEX IF NOT EXISTS idx_permit_number ON local_permits(permit_number);
CREATE INDEX IF NOT EXISTS idx_permit_vehicle_id ON local_permits(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_permit_driver_id ON local_permits(driver_id);
CREATE INDEX IF NOT EXISTS idx_permit_organization_id ON local_permits(organization_id);
CREATE INDEX IF NOT EXISTS idx_permit_status ON local_permits(status);
CREATE INDEX IF NOT EXISTS idx_permit_valid_to ON local_permits(valid_to);
CREATE INDEX IF NOT EXISTS idx_permit_type ON local_permits(permit_type);

-- ============================================================================
-- 6. GST Compliance Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS gst_compliances (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id         UUID NOT NULL,
    gstin                   VARCHAR(20) NOT NULL UNIQUE,
    status                  VARCHAR(20) NOT NULL CHECK (status IN ('active', 'cancelled', 'suspended')),
    legal_name              VARCHAR(200) NOT NULL,
    trade_name              VARCHAR(200),
    address                 TEXT NOT NULL,
    centre_jurisdiction     VARCHAR(200),
    state_jurisdiction      VARCHAR(200),
    date_of_registration    TIMESTAMPTZ NOT NULL,
    date_of_cancellation    TIMESTAMPTZ,
    last_updated            TIMESTAMPTZ NOT NULL,
    return_filing_frequency VARCHAR(20), -- monthly, quarterly, annually
    return_filing_status    VARCHAR(30), -- compliant, non_compliant
    payment_status          VARCHAR(30), -- compliant, non_compliant
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID NOT NULL,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by              UUID NOT NULL
);

-- Indexes for GST Compliance
CREATE INDEX IF NOT EXISTS idx_gst_gstin ON gst_compliances(gstin);
CREATE INDEX IF NOT EXISTS idx_gst_organization_id ON gst_compliances(organization_id);
CREATE INDEX IF NOT EXISTS idx_gst_status ON gst_compliances(status);
CREATE INDEX IF NOT EXISTS idx_gst_return_status ON gst_compliances(return_filing_status);

-- ============================================================================
-- 7. Integration Configurations Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS integration_configs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL,
    service_type        VARCHAR(30) NOT NULL CHECK (service_type IN ('fastag', 'ewaybill', 'vahan', 'sarathi', 'gst', 'permit')),
    api_endpoint        VARCHAR(500) NOT NULL,
    api_key             VARCHAR(200),
    username            VARCHAR(100),
    password            VARCHAR(200),
    client_id           VARCHAR(100),
    client_secret       VARCHAR(200),
    environment         VARCHAR(20) NOT NULL CHECK (environment IN ('production', 'staging', 'sandbox')),
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    last_sync_at        TIMESTAMPTZ,
    next_sync_at        TIMESTAMPTZ,
    sync_interval       INTEGER NOT NULL DEFAULT 60, -- in minutes
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by          UUID NOT NULL,

    CONSTRAINT unique_org_service UNIQUE (organization_id, service_type)
);

-- Indexes for Integration Configurations
CREATE INDEX IF NOT EXISTS idx_config_organization_id ON integration_configs(organization_id);
CREATE INDEX IF NOT EXISTS idx_config_service_type ON integration_configs(service_type);
CREATE INDEX IF NOT EXISTS idx_config_active ON integration_configs(active);
CREATE INDEX IF NOT EXISTS idx_config_environment ON integration_configs(environment);

-- ============================================================================
-- Triggers for auto-updating updated_at timestamps
-- ============================================================================

-- FASTag transactions
CREATE OR REPLACE FUNCTION update_fastag_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_fastag ON fastag_transactions;
CREATE TRIGGER trigger_update_fastag
    BEFORE UPDATE ON fastag_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_fastag_timestamp();

-- E-Way Bills
CREATE OR REPLACE FUNCTION update_ewaybill_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_ewaybill ON e_way_bills;
CREATE TRIGGER trigger_update_ewaybill
    BEFORE UPDATE ON e_way_bills
    FOR EACH ROW
    EXECUTE FUNCTION update_ewaybill_timestamp();

-- Vehicle Verifications
CREATE OR REPLACE FUNCTION update_vehicle_verification_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_vehicle_verification ON vehicle_verifications;
CREATE TRIGGER trigger_update_vehicle_verification
    BEFORE UPDATE ON vehicle_verifications
    FOR EACH ROW
    EXECUTE FUNCTION update_vehicle_verification_timestamp();

-- Driver Verifications
CREATE OR REPLACE FUNCTION update_driver_verification_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_driver_verification ON driver_verifications;
CREATE TRIGGER trigger_update_driver_verification
    BEFORE UPDATE ON driver_verifications
    FOR EACH ROW
    EXECUTE FUNCTION update_driver_verification_timestamp();

-- Local Permits
CREATE OR REPLACE FUNCTION update_permit_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_permit ON local_permits;
CREATE TRIGGER trigger_update_permit
    BEFORE UPDATE ON local_permits
    FOR EACH ROW
    EXECUTE FUNCTION update_permit_timestamp();

-- GST Compliance
CREATE OR REPLACE FUNCTION update_gst_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_gst ON gst_compliances;
CREATE TRIGGER trigger_update_gst
    BEFORE UPDATE ON gst_compliances
    FOR EACH ROW
    EXECUTE FUNCTION update_gst_timestamp();

-- Integration Configs
CREATE OR REPLACE FUNCTION update_config_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_config ON integration_configs;
CREATE TRIGGER trigger_update_config
    BEFORE UPDATE ON integration_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_config_timestamp();

-- ============================================================================
-- Comments for documentation
-- ============================================================================

COMMENT ON TABLE fastag_transactions IS 'FASTag electronic toll collection transactions';
COMMENT ON TABLE e_way_bills IS 'E-Way Bill electronic goods movement permits';
COMMENT ON TABLE vehicle_verifications IS 'VAHAN vehicle registration verification records';
COMMENT ON TABLE driver_verifications IS 'SARATHI driver license verification records';
COMMENT ON TABLE local_permits IS 'Local transport permits and authorizations';
COMMENT ON TABLE gst_compliances IS 'GST compliance status and filing records';
COMMENT ON TABLE integration_configs IS 'Third-party integration configurations for India services';

COMMENT ON COLUMN fastag_transactions.transaction_type IS 'debit or credit transaction';
COMMENT ON COLUMN fastag_transactions.status IS 'success, failed, or pending';
COMMENT ON COLUMN e_way_bills.document_type IS 'INV (Invoice), CHL (Challan), WBILL (Waybill), etc.';
COMMENT ON COLUMN vehicle_verifications.rc_status IS 'Registration Certificate status: active, inactive, suspended';
COMMENT ON COLUMN driver_verifications.license_class IS 'Array of license classes: MCWG, LMV, HMV, etc.';
COMMENT ON COLUMN local_permits.permit_type IS 'stage_carriage, contract_carriage, or goods_vehicle';
COMMENT ON COLUMN gst_compliances.return_filing_status IS 'compliant or non_compliant with filing requirements';
COMMENT ON COLUMN integration_configs.environment IS 'production, staging, or sandbox environment';

-- ============================================================================
-- End of Migration
-- ============================================================================
