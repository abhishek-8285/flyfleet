-- Migration: 051_invoice_rbac
-- Description: Add role-based access control for invoices with tenant isolation.

-- 1. Create invoice permissions table
CREATE TABLE IF NOT EXISTS invoice_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    role_name VARCHAR(50) NOT NULL,
    can_create_invoice BOOLEAN DEFAULT false,
    can_edit_invoice BOOLEAN DEFAULT false,
    can_delete_invoice BOOLEAN DEFAULT false,
    can_issue_invoice BOOLEAN DEFAULT false,
    can_cancel_invoice BOOLEAN DEFAULT false,
    can_record_payment BOOLEAN DEFAULT false,
    can_refund_payment BOOLEAN DEFAULT false,
    can_create_credit_note BOOLEAN DEFAULT false,
    can_approve_credit_note BOOLEAN DEFAULT false,
    can_view_invoices BOOLEAN DEFAULT true,
    can_export_invoices BOOLEAN DEFAULT false,
    can_approve_invoices BOOLEAN DEFAULT false,
    approval_limit DECIMAL(12, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, role_name)
);

-- 2. Create user invoice access overrides table
CREATE TABLE IF NOT EXISTS user_invoice_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL,
    override_role BOOLEAN DEFAULT false,
    can_create_invoice BOOLEAN DEFAULT false,
    can_edit_invoice BOOLEAN DEFAULT false,
    can_delete_invoice BOOLEAN DEFAULT false,
    can_issue_invoice BOOLEAN DEFAULT false,
    can_cancel_invoice BOOLEAN DEFAULT false,
    can_record_payment BOOLEAN DEFAULT false,
    can_approve_invoices BOOLEAN DEFAULT false,
    approval_limit DECIMAL(12, 2),
    max_invoice_amount DECIMAL(12, 2),
    require_approval_above DECIMAL(12, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, user_id)
);

-- 3. Create invoice approval matrix table
CREATE TABLE IF NOT EXISTS invoice_approval_matrix (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    min_amount DECIMAL(12, 2) NOT NULL,
    max_amount DECIMAL(12, 2),
    required_role VARCHAR(50) NOT NULL,
    required_approvals INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, min_amount, max_amount)
);

-- 4. Create invoice approvals table
CREATE TABLE IF NOT EXISTS invoice_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    approver_id UUID NOT NULL,
    approver_role VARCHAR(50),
    approval_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    approval_comments TEXT,
    requested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    decided_at TIMESTAMPTZ,
    ip_address INET,
    user_agent TEXT,
    UNIQUE(invoice_id, approver_id)
);

-- 5. Indexes
CREATE INDEX IF NOT EXISTS idx_invoice_permissions_tenant ON invoice_permissions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_invoice_access_user ON user_invoice_access(user_id);
CREATE INDEX IF NOT EXISTS idx_invoice_approval_matrix_tenant ON invoice_approval_matrix(tenant_id);

-- 6. Insert default role permissions
INSERT INTO invoice_permissions (tenant_id, role_name, can_create_invoice, can_edit_invoice, can_delete_invoice, can_issue_invoice, can_cancel_invoice, can_record_payment, can_view_invoices, can_export_invoices, can_approve_invoices) VALUES
('00000000-0000-0000-0000-000000000000', 'admin', true, true, true, true, true, true, true, true, true),
('00000000-0000-0000-0000-000000000000', 'manager', true, true, false, true, false, true, true, true, true),
('00000000-0000-0000-0000-000000000000', 'accountant', true, true, false, false, false, true, true, true, false),
('00000000-0000-0000-0000-000000000000', 'viewer', false, false, false, false, false, false, true, false, false)
ON CONFLICT (tenant_id, role_name) DO NOTHING;

-- 7. Permission check function
CREATE OR REPLACE FUNCTION check_invoice_permission(
    p_user_id UUID,
    p_tenant_id UUID,
    p_permission VARCHAR(50)
) RETURNS BOOLEAN AS $$
DECLARE
    user_role VARCHAR(50);
    has_permission BOOLEAN;
    override BOOLEAN;
BEGIN
    SELECT role INTO user_role FROM users WHERE id = p_user_id AND tenant_id = p_tenant_id;
    IF user_role IS NULL THEN RETURN false; END IF;
    
    SELECT override_role INTO override FROM user_invoice_access 
    WHERE user_id = p_user_id AND tenant_id = p_tenant_id AND deleted_at IS NULL;
    
    IF override THEN
        EXECUTE format('SELECT %I FROM user_invoice_access WHERE user_id = $1 AND tenant_id = $2 AND deleted_at IS NULL', p_permission)
        INTO has_permission USING p_user_id, p_tenant_id;
        RETURN COALESCE(has_permission, false);
    ELSE
        EXECUTE format('SELECT %I FROM invoice_permissions WHERE role_name = $1 AND tenant_id = $2', p_permission)
        INTO has_permission USING user_role, p_tenant_id;
        RETURN COALESCE(has_permission, false);
    END IF;
END;
$$ LANGUAGE plpgsql;
