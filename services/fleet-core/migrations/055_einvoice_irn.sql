-- Migration: 055_einvoice_irn
-- Description: Add e-invoicing (IRN) support for India GST compliance with tenant isolation.

-- 1. Create e-invoice IRN table
CREATE TABLE IF NOT EXISTS einvoice_irn (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    irn VARCHAR(64) UNIQUE,
    irn_status VARCHAR(20) DEFAULT 'pending',
    ack_no VARCHAR(100),
    ack_dt TIMESTAMPTZ,
    signed_response TEXT,
    qr_code_str TEXT,
    qr_code_image BYTEA,
    cancelled BOOLEAN DEFAULT false,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,
    retry_count INTEGER DEFAULT 0,
    last_retry_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    CONSTRAINT chk_einvoice_retry_count CHECK (retry_count >= 0 AND retry_count <= 5)
);

-- 2. Create e-invoice configuration table
CREATE TABLE IF NOT EXISTS einvoice_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) UNIQUE,
    app_key VARCHAR(255),
    secret_key VARCHAR(255),
    username VARCHAR(100),
    password VARCHAR(255),
    is_enabled BOOLEAN DEFAULT false,
    environment VARCHAR(20) DEFAULT 'sandbox',
    gstin VARCHAR(15),
    mandatory_above DECIMAL(12, 2) DEFAULT 50000,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_einvoice_gstin CHECK (gstin IS NULL OR LENGTH(gstin) = 15)
);

-- 3. Create e-invoice item mapping table
CREATE TABLE IF NOT EXISTS einvoice_item_mapping (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    product_id UUID,
    hsn_code VARCHAR(20) NOT NULL,
    sac_code VARCHAR(20),
    nic_code VARCHAR(20),
    gst_rate DECIMAL(5, 2),
    exemption_reason TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, hsn_code)
);

-- 4. Indexes
CREATE INDEX IF NOT EXISTS idx_einvoice_irn_invoice ON einvoice_irn(invoice_id);
CREATE INDEX IF NOT EXISTS idx_einvoice_irn_tenant ON einvoice_irn(tenant_id);

-- 5. Helper functions
CREATE OR REPLACE FUNCTION generate_irn_hash(
    p_invoice_number VARCHAR,
    p_fin_year VARCHAR,
    p_gstin VARCHAR,
    p_inv_type VARCHAR,
    p_inv_value DECIMAL
) RETURNS VARCHAR AS $$
DECLARE
    v_input TEXT;
BEGIN
    v_input := p_invoice_number || '|' || p_fin_year || '|' || p_gstin || '|' || p_inv_type || '|' || p_inv_value::TEXT;
    RETURN encode(sha256(v_input::bytea), 'hex');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION is_einvoice_mandatory(
    p_tenant_id UUID,
    p_amount DECIMAL
) RETURNS BOOLEAN AS $$
DECLARE
    v_mandatory_above DECIMAL;
    v_enabled BOOLEAN;
BEGIN
    SELECT mandatory_above, is_enabled INTO v_mandatory_above, v_enabled
    FROM einvoice_config WHERE tenant_id = p_tenant_id AND is_active = true;
    RETURN COALESCE(v_enabled, false) AND p_amount > COALESCE(v_mandatory_above, 50000);
END;
$$ LANGUAGE plpgsql;

-- 6. Compliance View
CREATE OR REPLACE VIEW einvoice_compliance_status AS
SELECT 
    i.id as invoice_id,
    i.invoice_number,
    i.tenant_id,
    i.total_amount,
    ei.irn,
    ei.irn_status,
    ei.ack_no,
    ei.created_at as irn_created_at,
    CASE 
        WHEN ei.irn_status = 'generated' THEN 'Compliant'
        WHEN ei.irn_status = 'pending' THEN 'Pending'
        WHEN ei.irn_status = 'failed' THEN 'Failed'
        WHEN is_einvoice_mandatory(i.tenant_id, i.total_amount) THEN 'Non-Compliant'
        ELSE 'Not Required'
    END as compliance_status
FROM invoices i
LEFT JOIN einvoice_irn ei ON i.id = ei.invoice_id
WHERE i.deleted_at IS NULL;
