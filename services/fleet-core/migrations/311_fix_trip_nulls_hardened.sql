-- Fix Trips NULLs and Harden Schema
-- Migration: 311_fix_trip_nulls_hardened.sql

-- 1. Repair 'planned_distance'
UPDATE trips SET planned_distance = 0.00 WHERE planned_distance IS NULL;

-- 2. Repair 'estimated_duration_min'
UPDATE trips SET estimated_duration_min = 0 WHERE estimated_duration_min IS NULL;

-- 3. Repair 'scheduled_start'
UPDATE trips SET scheduled_start = now() WHERE scheduled_start IS NULL;

-- 4. Repair 'trip_type'
UPDATE trips SET trip_type = 'standard' WHERE trip_type IS NULL OR trip_type = '';

-- 5. Repair 'created_at' and 'updated_at' (just in case)
UPDATE trips SET created_at = now() WHERE created_at IS NULL;
UPDATE trips SET updated_at = now() WHERE updated_at IS NULL;

-- 6. Repair 'id' and 'tenant_id' (should not be NULL, but for completion)
-- No action needed as they are NOT NULL in schema.

-- 7. Apply NOT NULL constraints with defaults for future safety
ALTER TABLE trips ALTER COLUMN planned_distance SET DEFAULT 0.00;
ALTER TABLE trips ALTER COLUMN estimated_duration_min SET DEFAULT 0;
ALTER TABLE trips ALTER COLUMN scheduled_start SET DEFAULT now();
ALTER TABLE trips ALTER COLUMN trip_type SET DEFAULT 'standard';
ALTER TABLE trips ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE trips ALTER COLUMN updated_at SET DEFAULT now();

-- Note: We don't apply NOT NULL to these columns yet to avoid breaking 
-- any legacy code that might still be inserting NULLs, but we provide 
-- defaults at the DB level where possible.
