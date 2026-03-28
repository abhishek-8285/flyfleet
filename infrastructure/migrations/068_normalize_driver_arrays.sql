-- Migration: 068_normalize_driver_arrays.sql
-- Description: Extract attendance_records and schedules JSONB arrays into normalized tables
-- Created: 2026-03-28

SET search_path TO public;

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ========================================
-- 1. Create driver_attendance table
-- Replaces: drivers.attendance_records (JSONB array of AttendanceRecord)
-- Matches Go struct: models.AttendanceRecord
-- ========================================
CREATE TABLE IF NOT EXISTS driver_attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Attendance details (matching AttendanceRecord struct)
    attendance_date DATE NOT NULL,  -- maps to "date" field (date only)
    status VARCHAR(20) NOT NULL DEFAULT 'present', -- present, absent, leave, holiday
    check_in_time TIMESTAMPTZ,  -- nullable (*time.Time)
    check_out_time TIMESTAMPTZ, -- nullable (*time.Time)
    remarks TEXT,  -- maps to "remarks"
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints: One attendance record per driver per date
    UNIQUE(driver_id, attendance_date)
);

-- Indexes for driver_attendance
CREATE INDEX IF NOT EXISTS idx_driver_attendance_driver ON driver_attendance(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_attendance_org_date ON driver_attendance(org_id, attendance_date DESC);
CREATE INDEX IF NOT EXISTS idx_driver_attendance_date_status ON driver_attendance(attendance_date, status);

-- RLS for multi-tenancy
ALTER TABLE driver_attendance ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS driver_attendance_tenant_policy ON driver_attendance;
CREATE POLICY driver_attendance_tenant_policy ON driver_attendance
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Trigger for updated_at (for future updates if needed)
CREATE TRIGGER update_driver_attendance_updated_at
    BEFORE UPDATE ON driver_attendance
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- 2. Create driver_schedules table
-- Replaces: drivers.schedules (JSONB array of WorkSchedule)
-- Matches Go struct: models.WorkSchedule
-- ========================================
CREATE TABLE IF NOT EXISTS driver_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Schedule template details (matching WorkSchedule struct)
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday, 6=Saturday
    start_time VARCHAR(10) NOT NULL,  -- "HH:MM" format (e.g., "09:00")
    end_time VARCHAR(10) NOT NULL,    -- "HH:MM" format (e.g., "17:00")
    is_working BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints: One schedule per driver per day of week
    UNIQUE(driver_id, day_of_week)
);

-- Indexes for driver_schedules
CREATE INDEX IF NOT EXISTS idx_driver_schedules_driver ON driver_schedules(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_schedules_org_day ON driver_schedules(org_id, day_of_week);
CREATE INDEX IF NOT EXISTS idx_driver_schedules_day_working ON driver_schedules(day_of_week, is_working);

-- RLS for multi-tenancy
ALTER TABLE driver_schedules ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS driver_schedules_tenant_policy ON driver_schedules;
CREATE POLICY driver_schedules_tenant_policy ON driver_schedules
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Trigger for updated_at
CREATE TRIGGER update_driver_schedules_updated_at
    BEFORE UPDATE ON driver_schedules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- 3. Data Migration: Transfer existing JSONB data to new tables
-- ========================================

-- Migrate attendance_records from drivers.attendance_records JSONB
-- JSON structure per item: {"id":"uuid","date":"ISO8601","status":"string","check_in_time":"ISO8601|null","check_out_time":"ISO8601|null","remarks":"string","created_at":"ISO8601"}
DO $$
DECLARE
    driver_record RECORD;
    attendance_data JSONB;
    attendance_item JSONB;
    migrated_attendance INT := 0;
    skipped_null INT := 0;
BEGIN
    FOR driver_record IN
        SELECT id, org_id, attendance_records
        FROM drivers
        WHERE attendance_records IS NOT NULL
          AND jsonb_array_length(attendance_records) > 0
    LOOP
        FOR attendance_item IN SELECT jsonb_array_elements(driver_record.attendance_records)
        LOOP
            BEGIN
                -- Extract date part (time.Time stores full timestamp but we only need date)
                -- If stored as full timestamp, cast to date
                INSERT INTO driver_attendance (
                    id,
                    driver_id,
                    org_id,
                    attendance_date,
                    status,
                    check_in_time,
                    check_out_time,
                    remarks,
                    created_at
                ) VALUES (
                    COALESCE((attendance_item->>'id')::UUID, gen_random_uuid()),
                    driver_record.id,
                    driver_record.org_id,
                    (attendance_item->>'date')::DATE,
                    COALESCE(attendance_item->>'status', 'present'),
                    (attendance_item->>'check_in_time')::TIMESTAMPTZ,
                    (attendance_item->>'check_out_time')::TIMESTAMPTZ,
                    attendance_item->>'remarks',
                    COALESCE((attendance_item->>'created_at')::TIMESTAMPTZ, NOW())
                )
                ON CONFLICT (driver_id, attendance_date) DO NOTHING;

                migrated_attendance := migrated_attendance + 1;
            EXCEPTION
                WHEN others THEN
                    RAISE WARNING 'Failed to migrate attendance for driver %: %', driver_record.id, SQLERRM;
            END;
        END LOOP;
    END LOOP;

    -- Count drivers that had NULL attendance_records
    SELECT COUNT(*) INTO skipped_null
    FROM drivers
    WHERE attendance_records IS NULL OR jsonb_array_length(attendance_records) = 0;

    RAISE NOTICE 'Migrated % attendance records from drivers.attendance_records (skipped % drivers with empty/null)', migrated_attendance, skipped_null;
END $$;

-- Migrate schedules from drivers.schedules JSONB
-- JSON structure per item: {"id":"uuid","day_of_week":0-6,"start_time":"HH:MM","end_time":"HH:MM","is_working":true,"created_at":"ISO8601","updated_at":"ISO8601"}
DO $$
DECLARE
    driver_record RECORD;
    schedule_data JSONB;
    schedule_item JSONB;
    migrated_schedules INT := 0;
    skipped_null INT := 0;
BEGIN
    FOR driver_record IN
        SELECT id, org_id, schedules
        FROM drivers
        WHERE schedules IS NOT NULL
          AND jsonb_array_length(schedules) > 0
    LOOP
        FOR schedule_item IN SELECT jsonb_array_elements(driver_record.schedules)
        LOOP
            BEGIN
                INSERT INTO driver_schedules (
                    id,
                    driver_id,
                    org_id,
                    day_of_week,
                    start_time,
                    end_time,
                    is_working,
                    created_at,
                    updated_at
                ) VALUES (
                    COALESCE((schedule_item->>'id')::UUID, gen_random_uuid()),
                    driver_record.id,
                    driver_record.org_id,
                    (schedule_item->>'day_of_week')::INTEGER,
                    schedule_item->>'start_time',
                    schedule_item->>'end_time',
                    COALESCE((schedule_item->>'is_working')::BOOLEAN, true),
                    COALESCE((schedule_item->>'created_at')::TIMESTAMPTZ, NOW()),
                    COALESCE((schedule_item->>'updated_at')::TIMESTAMPTZ, NOW())
                )
                ON CONFLICT (driver_id, day_of_week) DO NOTHING;

                migrated_schedules := migrated_schedules + 1;
            EXCEPTION
                WHEN others THEN
                    RAISE WARNING 'Failed to migrate schedule for driver %: %', driver_record.id, SQLERRM;
            END;
        END LOOP;
    END LOOP;

    -- Count drivers that had NULL schedules
    SELECT COUNT(*) INTO skipped_null
    FROM drivers
    WHERE schedules IS NULL OR jsonb_array_length(schedules) = 0;

    RAISE NOTICE 'Migrated % schedule records from drivers.schedules (skipped % drivers with empty/null)', migrated_schedules, skipped_null;
END $$;

-- ========================================
-- 4. Keep JSONB columns for backward compatibility but add deprecation comments
-- ========================================

COMMENT ON COLUMN drivers.attendance_records IS 'DEPRECATED: Use driver_attendance table. Kept for backward compatibility during migration.';
COMMENT ON COLUMN drivers.schedules IS 'DEPRECATED: Use driver_schedules table. Kept for backward compatibility during migration.';

-- ========================================
-- 5. Document the new tables
-- ========================================
COMMENT ON TABLE driver_attendance IS 'Normalized driver attendance records (daily attendance tracking). Replaces attendance_records JSONB in drivers table.';
COMMENT ON TABLE driver_schedules IS 'Normalized driver work schedule templates (weekly patterns). Replaces schedules JSONB in drivers table.';

COMMENT ON COLUMN driver_attendance.attendance_date IS 'Date of the attendance record (date only, no time)';
COMMENT ON COLUMN driver_attendance.status IS 'Attendance status: present, absent, leave, holiday';
COMMENT ON COLUMN driver_attendance.check_in_time IS 'Actual check-in timestamp (null if not checked in)';
COMMENT ON COLUMN driver_attendance.check_out_time IS 'Actual check-out timestamp (null if not checked out)';

COMMENT ON COLUMN driver_schedules.day_of_week IS 'Day of week: 0=Sunday, 1=Monday, ..., 6=Saturday';
COMMENT ON COLUMN driver_schedules.start_time IS 'Scheduled shift start time in HH:MM format (24h)';
COMMENT ON COLUMN driver_schedules.end_time IS 'Scheduled shift end time in HH:MM format (24h)';
COMMENT ON COLUMN driver_schedules.is_working IS 'Whether the driver is scheduled to work on this day';

-- ========================================
-- 6. Verification
-- ========================================
DO $$
DECLARE
    new_tables_count INT;
BEGIN
    SELECT COUNT(*) INTO new_tables_count
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name IN ('driver_attendance', 'driver_schedules');

    IF new_tables_count = 2 THEN
        RAISE NOTICE 'Normalization complete: driver_attendance and driver_schedules tables created and populated';
    ELSE
        RAISE WARNING 'Expected 2 new tables, found %', new_tables_count;
    END IF;
END $$;
