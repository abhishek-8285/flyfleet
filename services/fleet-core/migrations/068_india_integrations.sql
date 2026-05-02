-- Migration: 068_india_integrations
-- Description: India-specific integrations with tenant isolation.

-- 1. FASTag Transactions
CREATE TABLE IF NOT EXISTS fastag_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS driver_id UUID NOT NULL;
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS tag_number VARCHAR(50) NOT NULL DEFAULT '';
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS transaction_id VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS transaction_type VARCHAR(20) NOT NULL DEFAULT 'debit';
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS amount DECIMAL(10, 2) NOT NULL DEFAULT 0;
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS toll_plaza VARCHAR(200) NOT NULL DEFAULT '';
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS transaction_time TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'success';
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS reconciled BOOLEAN DEFAULT FALSE;
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE fastag_transactions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_fastag_tenant ON fastag_transactions(tenant_id);

-- 2. E-Way Bills
CREATE TABLE IF NOT EXISTS e_way_bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS eway_bill_number VARCHAR(20) NOT NULL DEFAULT '';
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS eway_bill_date TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS eway_bill_valid_till TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS document_number VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS goods_value DECIMAL(15, 2) NOT NULL DEFAULT 0;
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS from_pincode VARCHAR(10) NOT NULL DEFAULT '';
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS to_pincode VARCHAR(10) NOT NULL DEFAULT '';
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS vehicle_number VARCHAR(20);
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS user_gstin VARCHAR(20) NOT NULL DEFAULT '';
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE e_way_bills ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_eway_tenant ON e_way_bills(tenant_id);

-- 3. Vehicle Verifications (VAHAN)
CREATE TABLE IF NOT EXISTS vehicle_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS registration_number VARCHAR(20) NOT NULL DEFAULT '';
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending';
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS verification_date TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS owner_name VARCHAR(200);
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS rc_status VARCHAR(20);
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS insurance_expiry TIMESTAMPTZ;
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS puc_expiry TIMESTAMPTZ;
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_vahan_tenant ON vehicle_verifications(tenant_id);

-- 4. Driver Verifications (SARATHI)
CREATE TABLE IF NOT EXISTS driver_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS driver_id UUID NOT NULL;
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS license_number VARCHAR(50) NOT NULL DEFAULT '';
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending';
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS verification_date TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS name VARCHAR(200) NOT NULL DEFAULT '';
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS validity_to DATE NOT NULL DEFAULT CURRENT_DATE;
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_sarathi_tenant ON driver_verifications(tenant_id);

-- 5. GST Compliance
CREATE TABLE IF NOT EXISTS gst_compliances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE gst_compliances ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE gst_compliances ADD COLUMN IF NOT EXISTS gstin VARCHAR(20) NOT NULL DEFAULT '';
    ALTER TABLE gst_compliances ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending';
    ALTER TABLE gst_compliances ADD COLUMN IF NOT EXISTS legal_name VARCHAR(200) NOT NULL DEFAULT '';
    ALTER TABLE gst_compliances ADD COLUMN IF NOT EXISTS date_of_registration TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE gst_compliances ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE gst_compliances ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_gst_tenant ON gst_compliances(tenant_id);

-- 6. Integration Configs
CREATE TABLE IF NOT EXISTS integration_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS service_type VARCHAR(30) NOT NULL DEFAULT 'vahan';
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS api_endpoint VARCHAR(500) NOT NULL DEFAULT '';
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS api_key VARCHAR(200);
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS environment VARCHAR(20) NOT NULL DEFAULT 'sandbox';
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT TRUE;
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE integration_configs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
    
    -- Ensure unique constraint
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'integration_configs' AND constraint_name = 'uq_integration_configs_tenant_service') THEN
        ALTER TABLE integration_configs ADD CONSTRAINT uq_integration_configs_tenant_service UNIQUE (tenant_id, service_type);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_int_config_tenant ON integration_configs(tenant_id);
