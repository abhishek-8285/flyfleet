-- Migration: 048_custom_billing_config
-- Description: Allow each organization to customize their invoices with tenant isolation.

-- 1. Create billing configurations table
CREATE TABLE IF NOT EXISTS billing_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) UNIQUE,
    company_name VARCHAR(255) NOT NULL,
    company_address TEXT,
    company_gstin VARCHAR(20),
    company_pan VARCHAR(10),
    company_email VARCHAR(255),
    company_phone VARCHAR(20),
    company_website VARCHAR(255),
    company_logo_url TEXT,
    primary_color VARCHAR(7) DEFAULT '#3b82f6',
    secondary_color VARCHAR(7) DEFAULT '#1e40af',
    theme VARCHAR(50) DEFAULT 'modern',
    bank_name VARCHAR(255),
    bank_account_name VARCHAR(255),
    bank_account_number VARCHAR(50),
    bank_ifsc VARCHAR(20),
    bank_branch VARCHAR(255),
    bank_upi VARCHAR(255),
    invoice_prefix VARCHAR(20) DEFAULT 'INV',
    invoice_terms TEXT[],
    invoice_notes TEXT,
    show_hsn BOOLEAN DEFAULT true,
    show_qr_code BOOLEAN DEFAULT true,
    show_payment_info BOOLEAN DEFAULT true,
    default_currency VARCHAR(10) DEFAULT 'INR',
    default_tax_calculation VARCHAR(20) DEFAULT 'exclusive',
    default_payment_terms_days INTEGER DEFAULT 30,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

CREATE INDEX IF NOT EXISTS idx_billing_config_tenant_id ON billing_configurations(tenant_id);

-- 2. Seed default config
INSERT INTO billing_configurations (tenant_id, company_name)
SELECT id, name FROM tenants
WHERE id NOT IN (SELECT tenant_id FROM billing_configurations)
ON CONFLICT (tenant_id) DO NOTHING;

-- 3. Configuration helper
CREATE OR REPLACE FUNCTION get_or_create_billing_config(p_tenant_id UUID)
RETURNS billing_configurations AS $$
DECLARE
    config billing_configurations;
BEGIN
    SELECT * INTO config FROM billing_configurations
    WHERE tenant_id = p_tenant_id AND is_active = true;
    
    IF config.id IS NULL THEN
        config.tenant_id := p_tenant_id;
        config.company_name := 'Organization';
        config.primary_color := '#3b82f6';
        config.invoice_prefix := 'INV';
        config.default_currency := 'INR';
        config.default_payment_terms_days := 30;
    END IF;
    
    RETURN config;
END;
$$ LANGUAGE plpgsql;
