-- 105_fix_admin_roles.sql
-- Fixes missing admin role assignments for the default test organization

DO $$
DECLARE
    u_id UUID := '550e8400-e29b-41d4-a716-446655440001'; -- admin@delta.io
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000'; -- Delta Logistics
    r_id UUID := '550e8400-e29b-41d4-a716-446655440999'; -- admin role
BEGIN
    -- 1. Ensure user exists
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = u_id) THEN
        RETURN;
    END IF;

    -- 2. Ensure tenant exists
    IF NOT EXISTS (SELECT 1 FROM tenants WHERE id = t_id) THEN
        RETURN;
    END IF;

    -- 3. Map to organization_users (Admin Support Domain)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organization_users') THEN
        INSERT INTO organization_users (tenant_id, email, role, status)
        VALUES (t_id, 'admin@delta.io', 'ADMIN', 'ACTIVE')
        ON CONFLICT DO NOTHING;
    END IF;

    -- 4. Map to user_tenants (IAM Domain)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_tenants') THEN
        INSERT INTO user_tenants (user_id, tenant_id, role, is_active)
        VALUES (u_id, t_id, 'admin', true)
        ON CONFLICT DO NOTHING;
    END IF;

    -- 5. Ensure admin role exists and is assigned (IAM Fine-grained)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        INSERT INTO roles (id, tenant_id, name, description, is_system, permission_mask)
        VALUES (r_id, t_id, 'admin', 'Administrator with full access', true, 9223372036854775807)
        ON CONFLICT DO NOTHING;

        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_roles') THEN
            INSERT INTO user_roles (user_id, role_id, tenant_id)
            VALUES (u_id, r_id, t_id)
            ON CONFLICT DO NOTHING;
        END IF;
    END IF;
END $$;
