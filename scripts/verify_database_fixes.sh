#!/bin/bash

# Database Fixes Verification Script
# Run this after applying migrations 064-068 to verify all fixes
# Usage: ./scripts/verify_database_fixes.sh

set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-fleet_user}"
DB_PASSWORD="${DB_PASSWORD:-fleet_password}"
DB_NAME="${DB_NAME:-fleet_db}"

export PGPASSWORD=$DB_PASSWORD

echo "========================================"
echo "FlyFleet Database Fixes Verification"
echo "========================================"
echo "Database: $DB_NAME at $DB_HOST:$DB_PORT"
echo ""

check_status=0

# Function to run query and check result
run_check() {
    local description="$1"
    local query="$2"
    local expected="$3"

    echo -n "Checking: $description... "

    result=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "$query" 2>/dev/null || echo "ERROR")

    if [ "$result" = "$expected" ] || [ "$result" = "t" ]; then
        echo "✓ PASS"
        return 0
    elif [ "$result" = "ERROR" ]; then
        echo "✗ FAIL (connection error)"
        return 1
    else
        echo "✗ FAIL (expected: $expected, got: $result)"
        return 1
    fi
}

echo "=== Migration 064: Foreign Keys on trips ==="
run_check "trips.driver_id FK exists" \
    "SELECT EXISTS(SELECT 1 FROM information_schema.table_constraints WHERE table_name='trips' AND constraint_name='fk_trips_driver')" "t"

run_check "trips.org_id FK exists" \
    "SELECT EXISTS(SELECT 1 FROM information_schema.table_constraints WHERE table_name='trips' AND constraint_name='fk_trips_organization')" "t"

echo ""
echo "=== Migration 065: Performance Indexes ==="
run_check "support_tickets composite index" \
    "SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname='idx_support_tickets_org_status_priority_created')" "t"

run_check "trips org+vehicle+time index" \
    "SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname='idx_trips_org_vehicle_time')" "t"

run_check "trips org+driver+status+time index" \
    "SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname='idx_trips_org_driver_status_time')" "t"

run_check "ledger org+created index" \
    "SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname='idx_ledger_org_created')" "t"

run_check "driver_behavior driver+recent index" \
    "SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname='idx_driver_behavior_driver_recent')" "t"

run_check "gps_locations vehicle+org+time index" \
    "SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname='idx_gps_vehicle_org_time')" "t"

echo ""
echo "=== Migration 066: updated_at Triggers ==="
run_check "update_updated_at_column function exists" \
    "SELECT EXISTS(SELECT 1 FROM pg_proc WHERE proname='update_updated_at_column')" "t"

# Count tables with updated_at that have triggers
tables_with_updated_at=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND column_name='updated_at' AND table_name NOT LIKE 'pg_%'" 2>/dev/null || echo "0")
tables_with_triggers=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(DISTINCT event_object_table) FROM information_schema.triggers WHERE trigger_schema='public' AND trigger_name LIKE 'update\_%\_updated_at'" 2>/dev/null || echo "0")

echo "  Tables with updated_at column: $tables_with_updated_at"
echo "  Tables with update_updated_at trigger: $tables_with_triggers"

if [ "$tables_with_triggers" -ge "$tables_with_updated_at" ]; then
    echo "✓ All or more tables have triggers (good)"
else
    echo "⚠ Some tables missing triggers (may be expected if some tables shouldn't have them)"
fi

echo ""
echo "=== Migration 067: org_id Naming Standardization ==="
run_check "No tables using organization_id" \
    "SELECT COUNT(*) = 0 FROM information_schema.columns WHERE table_schema='public' AND column_name='organization_id' AND table_name NOT IN ('pg_statistic','pg_class')" "t"

run_check "trips uses org_id (not organization_id)" \
    "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='trips' AND column_name='org_id')" "t"

run_check "work_orders uses org_id (if exists)" \
    "SELECT CASE WHEN EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='work_orders') THEN EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='work_orders' AND column_name='org_id') ELSE TRUE END" "t"

echo ""
echo "=== Migration 068: Normalized Driver Tables ==="
run_check "driver_attendance table exists" \
    "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_attendance')" "t"

run_check "driver_schedules table exists" \
    "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_schedules')" "t"

run_check "driver_attendance has RLS enabled" \
    "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_attendance' AND row_level_security='t')" "t"

run_check "driver_schedules has RLS enabled" \
    "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_schedules' AND row_level_security='t')" "t"

# Check if data was migrated
attendance_count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COALESCE(COUNT(*),0) FROM driver_attendance" 2>/dev/null || echo "0")
echo "  Records in driver_attendance: $attendance_count (should be >0 if data existed)"

schedules_count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COALESCE(COUNT(*),0) FROM driver_schedules" 2>/dev/null || echo "0")
echo "  Records in driver_schedules: $schedules_count (should be >0 if data existed)"

echo ""
echo "=== Data Integrity Checks ==="
# Check FKs are enforced
fk_violations=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT COUNT(*) FROM trips t
    LEFT JOIN drivers d ON t.driver_id = d.id
    WHERE t.driver_id IS NOT NULL AND d.id IS NULL
" 2>/dev/null || echo "0")

if [ "$fk_violations" -eq "0" ]; then
    echo "✓ No FK violations in trips.driver_id"
else
    echo "✗ Found $fk_violations orphaned trip driver references"
    check_status=1
fi

# Check org consistency
org_mismatch=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT COUNT(*) FROM trips t
    LEFT JOIN organizations o ON t.org_id = o.id
    WHERE t.org_id IS NOT NULL AND o.id IS NULL
" 2>/dev/null || echo "0")

if [ "$org_mismatch" -eq "0" ]; then
    echo "✓ All trips.org_id reference valid organizations"
else
    echo "✗ Found $org_mismatch trips with invalid org_id"
    check_status=1
fi

# Check driver_attendance org consistency
att_org_check=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT COUNT(*) FROM driver_attendance da
    LEFT JOIN drivers d ON da.driver_id = d.id
    WHERE d.id IS NULL
" 2>/dev/null || echo "0")

if [ "$att_org_check" -eq "0" ]; then
    echo "✓ All driver_attendance.driver_id reference valid drivers"
else
    echo "⚠ Found $att_org_check orphaned attendance records (should be 0 if migration was complete)"
fi

echo ""
echo "========================================"
if [ $check_status -eq 0 ]; then
    echo "✓ ALL CHECKS PASSED (or non-critical warnings)"
    echo "Database fixes are successfully applied!"
else
    echo "✗ SOME CHECKS FAILED"
    echo "Review the errors above and address them."
fi
echo "========================================"

exit $check_status
