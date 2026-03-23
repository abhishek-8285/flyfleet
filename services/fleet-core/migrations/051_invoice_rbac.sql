-- Migration: 051_invoice_rbac
-- Description: Add role-based access control for invoices

-- Create invoice permissions table
CREATE TABLE IF NOT EXISTS invoice_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    role_name VARCHAR(50) NOT NULL, -- admin, manager, accountant, viewer, etc.
    
    -- Invoice permissions
    can_create_invoice BOOLEAN DEFAULT false,
    can_edit_invoice BOOLEAN DEFAULT false,
    can_delete_invoice BOOLEAN DEFAULT false,
    can_issue_invoice BOOLEAN DEFAULT false,
    can_cancel_invoice BOOLEAN DEFAULT false,
    
    -- Payment permissions
    can_record_payment BOOLEAN DEFAULT false,
    can_refund_payment BOOLEAN DEFAULT false,
    
    -- Credit note permissions
    can_create_credit_note BOOLEAN DEFAULT false,
    can_approve_credit_note BOOLEAN DEFAULT false,
    
    -- View permissions
    can_view_invoices BOOLEAN DEFAULT true,
    can_export_invoices BOOLEAN DEFAULT false,
    
    -- Approval permissions
    can_approve_invoices BOOLEAN DEFAULT false,
    approval_limit DECIMAL(12, 2), -- Maximum amount user can approve
    
    -- Constraints
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(org_id, role_name)
);

-- Create user invoice access overrides table (for per-user permissions)
CREATE TABLE IF NOT EXISTS user_invoice_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    user_id UUID NOT NULL,
    
    -- Override role permissions
    override_role BOOLEAN DEFAULT false,
    
    -- Specific permissions (if override is true)
    can_create_invoice BOOLEAN DEFAULT false,
    can_edit_invoice BOOLEAN DEFAULT false,
    can_delete_invoice BOOLEAN DEFAULT false,
    can_issue_invoice BOOLEAN DEFAULT false,
    can_cancel_invoice BOOLEAN DEFAULT false,
    can_record_payment BOOLEAN DEFAULT false,
    can_approve_invoices BOOLEAN DEFAULT false,
    approval_limit DECIMAL(12, 2),
    
    -- Restrictions
    max_invoice_amount DECIMAL(12, 2), -- Maximum invoice amount user can create
    require_approval_above DECIMAL(12, 2), -- Require approval for invoices above this amount
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    UNIQUE(org_id, user_id)
);

-- Create invoice approval matrix table
CREATE TABLE IF NOT EXISTS invoice_approval_matrix (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    
    -- Approval thresholds
    min_amount DECIMAL(12, 2) NOT NULL,
    max_amount DECIMAL(12, 2), -- NULL means no upper limit
    
    -- Required approvers
    required_role VARCHAR(50) NOT NULL,
    required_approvals INTEGER DEFAULT 1,
    
    -- Conditions
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(org_id, min_amount, max_amount)
);

-- Create invoice approvals table
CREATE TABLE IF NOT EXISTS invoice_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    
    -- Approval details
    approver_id UUID NOT NULL,
    approver_role VARCHAR(50),
    approval_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    approval_comments TEXT,
    
    -- Timestamps
    requested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    decided_at TIMESTAMPTZ,
    
    -- Audit
    ip_address INET,
    user_agent TEXT,
    
    UNIQUE(invoice_id, approver_id)
);

-- Indexes
CREATE INDEX idx_invoice_permissions_org ON invoice_permissions(org_id);
CREATE INDEX idx_user_invoice_access_user ON user_invoice_access(user_id);
CREATE INDEX idx_invoice_approval_matrix_org ON invoice_approval_matrix(org_id);
CREATE INDEX idx_invoice_approvals_invoice ON invoice_approvals(invoice_id);
CREATE INDEX idx_invoice_approvals_approver ON invoice_approvals(approver_id);
CREATE INDEX idx_invoice_approvals_status ON invoice_approvals(approval_status);

-- Comments
COMMENT ON TABLE invoice_permissions IS 'Role-based permissions for invoice operations';
COMMENT ON TABLE user_invoice_access IS 'Per-user invoice access overrides';
COMMENT ON TABLE invoice_approval_matrix IS 'Approval rules based on invoice amount';
COMMENT ON TABLE invoice_approvals IS 'Invoice approval tracking';

-- Insert default role permissions
INSERT INTO invoice_permissions (org_id, role_name, can_create_invoice, can_edit_invoice, can_delete_invoice, can_issue_invoice, can_cancel_invoice, can_record_payment, can_view_invoices, can_export_invoices, can_approve_invoices) VALUES
('00000000-0000-0000-0000-000000000000', 'admin', true, true, true, true, true, true, true, true, true),
('00000000-0000-0000-0000-000000000000', 'manager', true, true, false, true, false, true, true, true, true),
('00000000-0000-0000-0000-000000000000', 'accountant', true, true, false, false, false, true, true, true, false),
('00000000-0000-0000-0000-000000000000', 'viewer', false, false, false, false, false, false, true, false, false)
ON CONFLICT (org_id, role_name) DO NOTHING;

-- Function to check if user has permission
CREATE OR REPLACE FUNCTION check_invoice_permission(
    p_user_id UUID,
    p_org_id UUID,
    p_permission VARCHAR(50)
) RETURNS BOOLEAN AS $$
DECLARE
    user_role VARCHAR(50);
    has_permission BOOLEAN;
    override BOOLEAN;
BEGIN
    -- Get user's role
    SELECT role INTO user_role FROM users WHERE id = p_user_id AND org_id = p_org_id;
    
    IF user_role IS NULL THEN
        RETURN false;
    END IF;
    
    -- Check for user-specific override
    SELECT override_role INTO override 
    FROM user_invoice_access 
    WHERE user_id = p_user_id AND org_id = p_org_id AND deleted_at IS NULL;
    
    IF override THEN
        -- Use user-specific permissions
        EXECUTE format('SELECT %I FROM user_invoice_access WHERE user_id = $1 AND org_id = $2 AND deleted_at IS NULL', p_permission)
        INTO has_permission USING p_user_id, p_org_id;
        RETURN COALESCE(has_permission, false);
    ELSE
        -- Use role-based permissions
        EXECUTE format('SELECT %I FROM invoice_permissions WHERE role_name = $1 AND org_id = $2', p_permission)
        INTO has_permission USING user_role, p_org_id;
        RETURN COALESCE(has_permission, false);
    END IF;
END;
$$ LANGUAGE plpgsql;
