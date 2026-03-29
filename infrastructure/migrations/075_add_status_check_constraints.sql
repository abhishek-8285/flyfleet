-- Migration: 075_add_status_check_constraints.sql
-- Description: Add CHECK constraints for status and enum-like columns
-- Created: 2026-03-29
-- Purpose: Enforce data integrity on status fields

SET search_path TO public;

\echo 'Adding CHECK constraints for status columns...';

-- ========================================
-- Add status constraints using a loop over predefined table/allowed-values pairs
-- ========================================
DO $$
DECLARE
    t RECORD;
    con_name TEXT;
    allowed_values TEXT;
BEGIN
    -- Iterate over all tables that have a 'status' column and need an enum constraint
    FOR t IN
        SELECT
            c.table_name,
            -- Determine allowed values based on table name
            CASE c.table_name
                -- Organizations
                WHEN 'organizations' THEN '''active'',''inactive'',''suspended'',''pending'',''archived'''
                -- Users
                WHEN 'users' THEN '''active'',''inactive'',''suspended'',''pending'',''locked'''
                -- Vehicles
                WHEN 'vehicles' THEN '''active'',''inactive'',''maintenance'',''retired'',''out_of_service'''
                -- Drivers
                WHEN 'drivers' THEN '''active'',''inactive'',''suspended'',''on_leave'',''terminated'''
                -- Trips
                WHEN 'trips' THEN '''planned'',''scheduled'',''in_progress'',''completed'',''cancelled'',''delayed'',''no_show'',''emergency'''
                -- Alerts
                WHEN 'alerts' THEN '''new'',''acknowledged'',''resolved'',''dismissed'',''escalated'''
                -- Support Tickets
                WHEN 'support_tickets' THEN '''open'',''in_progress'',''waiting_customer'',''resolved'',''closed'',''reopened'''
                -- Devices
                WHEN 'devices' THEN '''active'',''inactive'',''maintenance'',''offline'',''error'''
                -- MFA Codes
                WHEN 'mfa_codes' THEN '''pending'',''verified'',''expired'',''failed'''
                -- Bulk Jobs
                WHEN 'bulk_jobs' THEN '''pending'',''processing'',''completed'',''failed'',''cancelled'''
                -- Firmware Rollouts
                WHEN 'firmware_rollouts' THEN '''draft'',''scheduled'',''in_progress'',''paused'',''completed'',''cancelled'''
                -- Billing Subscriptions
                WHEN 'billing_subscriptions' THEN '''active'',''inactive'',''past_due'',''canceled'',''expired'''
                -- Invoices
                WHEN 'invoices' THEN '''draft'',''sent'',''viewed'',''paid'',''overdue'',''void'',''refunded'''
                -- Route Optimization Jobs
                WHEN 'route_optimization_jobs' THEN '''pending'',''processing'',''completed'',''failed'',''cancelled'''
                -- Default: skip (no constraint)
                ELSE NULL
            END AS allowed_values
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.column_name = 'status'
    LOOP
        -- Skip if we don't have allowed values defined for this table
        IF t.allowed_values IS NULL THEN
            CONTINUE;
        END IF;

        con_name := 'chk_' || t.table_name || '_status';

        -- Check if constraint already exists
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints tc
            WHERE tc.table_schema = 'public'
              AND tc.table_name = t.table_name
              AND tc.constraint_name = con_name
        ) THEN
            BEGIN
                EXECUTE format(
                    'ALTER TABLE %I ADD CONSTRAINT %I CHECK (status IN (%s))',
                    t.table_name,
                    con_name,
                    t.allowed_values
                );
                RAISE NOTICE 'Added constraint: % on %', con_name, t.table_name;
            EXCEPTION
                WHEN others THEN
                    RAISE WARNING 'Failed to add constraint % on %: %', con_name, t.table_name, SQLERRM;
            END;
        ELSE
            RAISE NOTICE 'Constraint % already exists on %', con_name, t.table_name;
        END IF;
    END LOOP;
END $$;

-- ========================================
-- Additional column-specific constraints: fuel_type, license_type
-- ========================================

-- vehicles.fuel_type
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'vehicles' AND column_name = 'fuel_type'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_schema = 'public' AND table_name = 'vehicles' AND constraint_name = 'chk_vehicles_fuel_type'
        ) THEN
            BEGIN
                EXECUTE 'ALTER TABLE vehicles ADD CONSTRAINT chk_vehicles_fuel_type CHECK (fuel_type IN (''petrol'',''diesel'',''electric'',''hybrid'',''cng'',''lpg''))';
                RAISE NOTICE 'Added constraint: chk_vehicles_fuel_type';
            EXCEPTION
                WHEN others THEN RAISE WARNING 'Failed to add chk_vehicles_fuel_type: %', SQLERRM;
            END;
        END IF;
    END IF;
END $$;

-- drivers.license_type
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'drivers' AND column_name = 'license_type'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_schema = 'public' AND table_name = 'drivers' AND constraint_name = 'chk_drivers_license_type'
        ) THEN
            BEGIN
                EXECUTE 'ALTER TABLE drivers ADD CONSTRAINT chk_drivers_license_type CHECK (license_type IN (''LMV'',''HMV'',''Transport'',''Heavy'',''Medium'',''Light''))';
                RAISE NOTICE 'Added constraint: chk_drivers_license_type';
            EXCEPTION
                WHEN others THEN RAISE WARNING 'Failed to add chk_drivers_license_type: %', SQLERRM;
            END;
        END IF;
    END IF;
END $$;

-- ========================================
-- Add comments for constraints
-- ========================================
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN
        SELECT table_name, constraint_name
        FROM information_schema.table_constraints
        WHERE table_schema = 'public'
          AND constraint_name LIKE 'chk_%_status'
    LOOP
        BEGIN
            EXECUTE format(
                'COMMENT ON CONSTRAINT %I ON %I IS ''Enforces allowed values for status column''',
                r.constraint_name,
                r.table_name
            );
        EXCEPTION
            WHEN others THEN NULL;
        END;
    END LOOP;
END $$;

\echo '====================================================='
\echo 'Status CHECK constraints added!'
\echo '====================================================='
\echo 'Verification:'
\echo '  SELECT table_name, constraint_name FROM information_schema.table_constraints WHERE constraint_name LIKE ''chk_%_status'';'
\echo '====================================================='
