-- Migration: 048_custom_billing_config
-- Description: Allow each organization to customize their invoices

-- Create billing configurations table
CREATE TABLE IF NOT EXISTS billing_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL UNIQUE,
    
    -- Company Information
    company_name VARCHAR(255) NOT NULL,
    company_address TEXT,
    company_gstin VARCHAR(20),
    company_pan VARCHAR(10),
    company_email VARCHAR(255),
    company_phone VARCHAR(20),
    company_website VARCHAR(255),
    company_logo_url TEXT, -- Base64 or S3 URL
    
    -- Branding
    primary_color VARCHAR(7) DEFAULT '#3b82f6',
    secondary_color VARCHAR(7) DEFAULT '#1e40af',
    theme VARCHAR(50) DEFAULT 'modern', -- modern, classic, minimal
    
    -- Payment Details
    bank_name VARCHAR(255),
    bank_account_name VARCHAR(255),
    bank_account_number VARCHAR(50),
    bank_ifsc VARCHAR(20),
    bank_branch VARCHAR(255),
    bank_upi VARCHAR(255),
    
    -- Invoice Settings
    invoice_prefix VARCHAR(20) DEFAULT 'INV',
    invoice_terms TEXT[], -- Array of terms
    invoice_notes TEXT,
    show_hsn BOOLEAN DEFAULT true,
    show_qr_code BOOLEAN DEFAULT true,
    show_payment_info BOOLEAN DEFAULT true,
    
    -- Defaults
    default_currency VARCHAR(10) DEFAULT 'INR',
    default_tax_calculation VARCHAR(20) DEFAULT 'exclusive', -- inclusive/exclusive
    default_payment_terms_days INTEGER DEFAULT 30,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- Indexes
CREATE INDEX idx_billing_config_org_id ON billing_configurations(org_id);
CREATE INDEX idx_billing_config_active ON billing_configurations(is_active);

-- Comments
COMMENT ON TABLE billing_configurations IS 'Per-organization billing/invoice customization';
COMMENT ON COLUMN billing_configurations.org_id IS 'Organization ID (one config per org)';
COMMENT ON COLUMN billing_configurations.company_logo_url IS 'Base64 encoded or S3 URL';
COMMENT ON COLUMN billing_configurations.invoice_terms IS 'Array of terms and conditions';

-- Insert default config for existing organizations (if any)
INSERT INTO billing_configurations (
    org_id,
    company_name,
    primary_color,
    invoice_prefix,
    default_currency,
    default_payment_terms_days
)
SELECT 
    id as org_id,
    name as company_name,
    '#3b82f6' as primary_color,
    'INV' as invoice_prefix,
    'INR' as default_currency,
    30 as default_payment_terms_days
FROM organizations
WHERE id NOT IN (SELECT org_id FROM billing_configurations)
ON CONFLICT (org_id) DO NOTHING;

-- Function to get or create billing config
CREATE OR REPLACE FUNCTION get_or_create_billing_config(p_org_id UUID)
RETURNS billing_configurations AS $$
DECLARE
    config billing_configurations;
BEGIN
    -- Try to get existing config
    SELECT * INTO config
    FROM billing_configurations
    WHERE org_id = p_org_id AND is_active = true;
    
    -- If not found, return defaults
    IF config.id IS NULL THEN
        config.org_id := p_org_id;
        config.company_name := 'Organization';
        config.primary_color := '#3b82f6';
        config.invoice_prefix := 'INV';
        config.default_currency := 'INR';
        config.default_payment_terms_days := 30;
        config.show_hsn := true;
        config.show_qr_code := true;
        config.show_payment_info := true;
    END IF;
    
    RETURN config;
END;
$$ LANGUAGE plpgsql;
