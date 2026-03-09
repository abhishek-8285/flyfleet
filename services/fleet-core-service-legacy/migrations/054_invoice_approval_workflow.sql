-- Migration: 054_invoice_approval_workflow
-- Description: Complete invoice approval workflow with multi-level approvals

-- Add approval status column to invoices
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approval_status VARCHAR(20) DEFAULT 'not_required';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approval_required_at TIMESTAMPTZ;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS approval_completed_at TIMESTAMPTZ;

-- Add constraint for valid approval status
ALTER TABLE invoices ADD CONSTRAINT chk_invoice_approval_status 
    CHECK (approval_status IN ('not_required', 'pending', 'approved', 'rejected'));

-- Update invoice_approvals table with more fields
ALTER TABLE invoice_approvals ADD COLUMN IF NOT EXISTS approval_level INTEGER DEFAULT 1;
ALTER TABLE invoice_approvals ADD COLUMN IF NOT EXISTS is_required BOOLEAN DEFAULT true;

-- Create approval policy table
CREATE TABLE IF NOT EXISTS invoice_approval_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    
    -- Policy name and description
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Approval conditions
    min_amount DECIMAL(12, 2) NOT NULL,
    max_amount DECIMAL(12, 2), -- NULL means no upper limit
    
    -- Required approvers
    required_approvals INTEGER DEFAULT 1,
    approver_roles TEXT[] NOT NULL, -- Roles that can approve
    specific_approvers UUID[], -- Specific user IDs (optional)
    
    -- Workflow
    approval_type VARCHAR(20) DEFAULT 'any', -- any (any approver), all (all approvers), sequential
    priority INTEGER DEFAULT 0,
    
    -- Auto-approval rules
    auto_approve_below DECIMAL(12, 2), -- Auto-approve invoices below this amount
    auto_approve_roles TEXT[], -- Roles that can auto-approve
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    
    UNIQUE(org_id, name)
);

-- Indexes
CREATE INDEX idx_invoices_approval_status ON invoices(approval_status);
CREATE INDEX idx_invoice_approval_policies_org ON invoice_approval_policies(org_id);
CREATE INDEX idx_invoice_approval_policies_active ON invoice_approval_policies(is_active);

-- Comments
COMMENT ON COLUMN invoices.approval_status IS 'Approval workflow status';
COMMENT ON TABLE invoice_approval_policies IS 'Approval policies based on amount and other criteria';
COMMENT ON COLUMN invoice_approval_policies.approval_type IS 'any=any approver, all=all approvers, sequential=ordered approvals';

-- Function to check if invoice requires approval
CREATE OR REPLACE FUNCTION invoice_requires_approval(
    p_org_id UUID,
    p_amount DECIMAL,
    p_created_by UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_policy_count INTEGER;
BEGIN
    -- Check if any active policy applies
    SELECT COUNT(*) INTO v_policy_count
    FROM invoice_approval_policies
    WHERE org_id = p_org_id
      AND is_active = true
      AND min_amount <= p_amount
      AND (max_amount IS NULL OR max_amount >= p_amount);
    
    RETURN v_policy_count > 0;
END;
$$ LANGUAGE plpgsql;

-- Function to get required approvers for an invoice
CREATE OR REPLACE FUNCTION get_required_approvers(
    p_org_id UUID,
    p_amount DECIMAL
) RETURNS TABLE(role VARCHAR, specific_user UUID) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        unnest(app.approver_roles) as role,
        unnest(app.specific_approvers) as specific_user
    FROM invoice_approval_policies app
    WHERE app.org_id = p_org_id
      AND app.is_active = true
      AND app.min_amount <= p_amount
      AND (app.max_amount IS NULL OR app.max_amount >= p_amount)
    ORDER BY app.priority DESC;
END;
$$ LANGUAGE plpgsql;
