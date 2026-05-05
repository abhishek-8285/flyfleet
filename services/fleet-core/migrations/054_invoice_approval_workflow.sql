-- Migration: 054_invoice_approval_workflow
-- Description: Complete invoice approval workflow with multi-level approvals

-- Add approval status column to invoices
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approval_status VARCHAR(20) DEFAULT 'not_required';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approval_required_at TIMESTAMPTZ;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approval_completed_at TIMESTAMPTZ;

-- Add constraint for valid approval status
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name = 'invoices' AND constraint_name = 'chk_invoice_approval_status') THEN
        ALTER TABLE invoices ADD CONSTRAINT chk_invoice_approval_status 
            CHECK (approval_status IN ('not_required', 'pending', 'approved', 'rejected'));
    END IF;
END $$;

-- Update invoice_approvals table with more fields
ALTER TABLE invoice_approvals ADD COLUMN IF NOT EXISTS approval_level INTEGER DEFAULT 1;
ALTER TABLE invoice_approvals ADD COLUMN IF NOT EXISTS is_required BOOLEAN DEFAULT true;

-- Create approval policy table
CREATE TABLE IF NOT EXISTS invoice_approval_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

-- Defensive setup for all columns
DO $$ 
BEGIN 
    -- Column Harmonization (Legacy org_id -> tenant_id)
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'invoice_approval_policies' AND column_name = 'org_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'invoice_approval_policies' AND column_name = 'tenant_id') THEN
        ALTER TABLE invoice_approval_policies RENAME COLUMN org_id TO tenant_id;
    END IF;

    -- Ensure all columns exist
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS name VARCHAR(255);
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS min_amount DECIMAL(12, 2) NOT NULL DEFAULT 0;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS max_amount DECIMAL(12, 2);
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS required_approvals INTEGER DEFAULT 1;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS approver_roles TEXT[] NOT NULL DEFAULT '{}';
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS specific_approvers UUID[];
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS approval_type VARCHAR(20) DEFAULT 'any';
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS auto_approve_below DECIMAL(12, 2);
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS auto_approve_roles TEXT[];
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS created_by UUID;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE invoice_approval_policies ADD COLUMN IF NOT EXISTS updated_by UUID;

    -- Constraints (Defensive)
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'invoice_approval_policies' AND constraint_type = 'UNIQUE' AND constraint_name = 'uq_invoice_approval_policies_tenant_name') THEN
        ALTER TABLE invoice_approval_policies ADD CONSTRAINT uq_invoice_approval_policies_tenant_name UNIQUE(tenant_id, name);
    END IF;
END $$;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_invoices_approval_status ON invoices(approval_status);
CREATE INDEX IF NOT EXISTS idx_invoice_approval_policies_tenant ON invoice_approval_policies(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invoice_approval_policies_active ON invoice_approval_policies(is_active);

-- Comments
COMMENT ON COLUMN invoices.approval_status IS 'Approval workflow status';
COMMENT ON TABLE invoice_approval_policies IS 'Approval policies based on amount and other criteria';
COMMENT ON COLUMN invoice_approval_policies.approval_type IS 'any=any approver, all=all approvers, sequential=ordered approvals';

-- Function to check if invoice requires approval
CREATE OR REPLACE FUNCTION invoice_requires_approval(
    p_tenant_id UUID,
    p_amount DECIMAL,
    p_created_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_policy_count INTEGER;
BEGIN
    -- Check if any active policy applies
    SELECT COUNT(*) INTO v_policy_count
    FROM invoice_approval_policies
    WHERE tenant_id = p_tenant_id
      AND is_active = true
      AND min_amount <= p_amount
      AND (max_amount IS NULL OR max_amount >= p_amount);
    
    RETURN v_policy_count > 0;
END;
$$ LANGUAGE plpgsql;

-- Function to get required approvers for an invoice
CREATE OR REPLACE FUNCTION get_required_approvers(
    p_tenant_id UUID,
    p_amount DECIMAL
) RETURNS TABLE(role VARCHAR, specific_user UUID) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        unnest(app.approver_roles) as role,
        unnest(app.specific_approvers) as specific_user
    FROM invoice_approval_policies app
    WHERE app.tenant_id = p_tenant_id
      AND app.is_active = true
      AND app.min_amount <= p_amount
      AND (app.max_amount IS NULL OR app.max_amount >= p_amount)
    ORDER BY app.priority DESC;
END;
$$ LANGUAGE plpgsql;
