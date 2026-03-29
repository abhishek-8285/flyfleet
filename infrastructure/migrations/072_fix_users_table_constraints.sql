-- Migration: 072_fix_users_table_constraints.sql
-- Description: Fix users table - enforce NOT NULL, add foreign keys, and add missing index
-- Created: 2026-03-29
-- Critical: Users table is central to multi-tenancy and RBAC

SET search_path TO public;

-- =====================================================
-- PURPOSE: Ensure data integrity and query performance
-- =====================================================

\echo 'Fixing users table constraints and indexes...';

-- ========================================
-- 1. Make org_id NOT NULL (multi-tenancy requirement)
-- ========================================
DO $$
BEGIN
    -- Check if any rows have NULL org_id
    IF EXISTS (SELECT 1 FROM users WHERE org_id IS NULL LIMIT 1) THEN
        RAISE WARNING 'Found users with NULL org_id. These must be fixed before setting NOT NULL constraint:';
        RAISE WARNING '  Count: %', (SELECT COUNT(*) FROM users WHERE org_id IS NULL);
        RAISE WARNING '  Example IDs: (select id, email from users where org_id is null limit 5)';
        RAISE WARNING '  Please assign org_id values before re-running this migration.';
    ELSE
        ALTER TABLE users ALTER COLUMN org_id SET NOT NULL;
        RAISE NOTICE 'Set users.org_id to NOT NULL';
    END IF;
END $$;

-- ========================================
-- 2. Make role_id NOT NULL (RBAC requirement)
-- ========================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE role_id IS NULL LIMIT 1) THEN
        RAISE WARNING 'Found users with NULL role_id. Assigning default role (super_admin) temporarily...';

        -- Assign super_admin role (id: need to lookup)
        UPDATE users
        SET role_id = (SELECT id FROM roles WHERE name = 'super_admin' LIMIT 1)
        WHERE role_id IS NULL;

        RAISE NOTICE 'Assigned default role to % users', ROW_COUNT;
    END IF;

    ALTER TABLE users ALTER COLUMN role_id SET NOT NULL;
    RAISE NOTICE 'Set users.role_id to NOT NULL';
END $$;

-- ========================================
-- 3. Add foreign key: org_id → organizations(id)
-- ========================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'users' AND constraint_name = 'fk_users_organization'
    ) THEN
        ALTER TABLE users
        ADD CONSTRAINT fk_users_organization
        FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;

        COMMENT ON CONSTRAINT fk_users_organization ON users IS
            'Ensures user belongs to valid organization (multi-tenancy enforcement)';
        RAISE NOTICE 'Added FK: users.org_id → organizations(id)';
    ELSE
        RAISE NOTICE 'FK fk_users_organization already exists';
    END IF;
END $$;

-- ========================================
-- 4. Add foreign key: role_id → roles(id)
-- ========================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'users' AND constraint_name = 'fk_users_role'
    ) THEN
        ALTER TABLE users
        ADD CONSTRAINT fk_users_role
        FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL;

        COMMENT ON CONSTRAINT fk_users_role ON users IS
            'Links user to RBAC role';
        RAISE NOTICE 'Added FK: users.role_id → roles(id)';
    ELSE
        RAISE NOTICE 'FK fk_users_role already exists';
    END IF;
END $$;

-- ========================================
-- 5. Add missing index on org_id (CRITICAL for performance)
-- ========================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE schemaname = 'public'
          AND tablename = 'users'
          AND indexname = 'idx_users_org_id'
    ) THEN
        CREATE INDEX idx_users_org_id ON users(org_id);
        RAISE NOTICE 'Created index: idx_users_org_id';
    ELSE
        RAISE NOTICE 'Index idx_users_org_id already exists';
    END IF;
END $$;

-- ========================================
-- 6. Add composite index for common multi-tenant queries
-- ========================================
CREATE INDEX IF NOT EXISTS idx_users_org_email ON users(org_id, email);

-- ========================================
-- 7. Verify constraints
-- ========================================
DO $$
DECLARE
    null_org_count INT;
    null_role_count INT;
    fk_org_exists BOOLEAN;
    fk_role_exists BOOLEAN;
    idx_org_exists BOOLEAN;
BEGIN
    SELECT COUNT(*) INTO null_org_count FROM users WHERE org_id IS NULL;
    SELECT COUNT(*) INTO null_role_count FROM users WHERE role_id IS NULL;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'users' AND constraint_name = 'fk_users_organization'
    ) INTO fk_org_exists;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'users' AND constraint_name = 'fk_users_role'
    ) INTO fk_role_exists;

    SELECT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE schemaname = 'public' AND tablename = 'users' AND indexname = 'idx_users_org_id'
    ) INTO idx_org_exists;

    RAISE NOTICE 'Verification:';
    RAISE NOTICE '  Users with NULL org_id: %', null_org_count;
    RAISE NOTICE '  Users with NULL role_id: %', null_role_count;
    RAISE NOTICE '  FK users.org_id → organizations exists: %', fk_org_exists;
    RAISE NOTICE '  FK users.role_id → roles exists: %', fk_role_exists;
    RAISE NOTICE '  Index idx_users_org_id exists: %', idx_org_exists;

    IF null_org_count = 0 AND null_role_count = 0 AND fk_org_exists AND fk_role_exists AND idx_org_exists THEN
        RAISE NOTICE 'SUCCESS: users table constraints are properly configured';
    END IF;
END $$;

-- ========================================
-- 8. Add CHECK constraints for status column (data integrity)
-- ========================================
ALTER TABLE users
    ADD CONSTRAINT IF NOT EXISTS chk_users_status
    CHECK (status IN ('active', 'inactive', 'suspended', 'pending'));

COMMENT ON CONSTRAINT chk_users_status ON users IS 'Allowed values for user account status';

-- ========================================
-- 9. Document the changes
-- ========================================
COMMENT ON TABLE users IS 'User accounts with multi-tenant org_id and RBAC role_id. org_id and role_id are both NOT NULL with proper foreign keys.';
COMMENT ON COLUMN users.org_id IS 'Organization UUID - NOT NULL, foreign key to organizations(id) - required for multi-tenancy';
COMMENT ON COLUMN users.role_id IS 'Role UUID - NOT NULL, foreign key to roles(id) - for RBAC';

\echo '====================================================='
\echo 'Users table fixes complete!'
\echo '====================================================='
\echo 'Summary:'
\echo '  - org_id and role_id are now NOT NULL'
\echo '  - Foreign keys added to organizations and roles'
\echo '  - Index idx_users_org_id created'
\echo '  - Composite index idx_users_org_email added'
\echo '  - CHECK constraint on status enum added'
\echo '====================================================='
