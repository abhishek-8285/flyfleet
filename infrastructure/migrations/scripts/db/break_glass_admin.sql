-- break_glass_admin.sql
-- WARNING: DO NOT RUN THIS SCRIPT UNLESS IN AN EMERGENCY RECOVERY SCENARIO.
-- This script creates a high-privilege Postgres role that completely ignores Row-Level Security (RLS).
-- It allows a database administrator to stitch together cross-tenant data without disabling policies globally.

-- ==========================================
-- 1. ENACTING THE EMERGENCY ROLE
-- ==========================================
BEGIN;

-- Create the role with explicit BYPASSRLS privilege.
CREATE ROLE flyfleet_breakglass WITH LOGIN PASSWORD 'CHANGE_ME_IN_PROD' BYPASSRLS;

-- Grant it access to the actual data schemas
GRANT USAGE ON SCHEMA public TO flyfleet_breakglass;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO flyfleet_breakglass;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO flyfleet_breakglass;

COMMIT;

/*
-- ==========================================
-- 2. USAGE (Run queries as the breakglass user)
-- ==========================================
-- Example: Connecting via psql
-- psql -h your-db-host -U flyfleet_breakglass -d fleet_db

SELECT vehicle_id, org_id FROM telemetry_events WHERE org_id = 'xxx' OR org_id = 'yyy';
-- Notice that we DID NOT call set_config('app.current_org_id', 'xxx', false);
*/

-- ==========================================
-- 3. REVOKING THE EMERGENCY ROLE (POST-RECOVERY)
-- ==========================================
/*
BEGIN;

-- Revoke schema bindings
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM flyfleet_breakglass;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM flyfleet_breakglass;
REVOKE USAGE ON SCHEMA public FROM flyfleet_breakglass;

-- Terminate any active sessions holding this role
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.usename = 'flyfleet_breakglass';

-- Drop the role entirely
DROP ROLE IF EXISTS flyfleet_breakglass;

COMMIT;
*/
