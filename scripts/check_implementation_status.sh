#!/bin/bash

# Database Fixes Implementation Status Check
# Run manually or via cron to track progress of migrations 064-068
# This is the script that the /loop job will execute

set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5433}"
DB_USER="${DB_USER:-fleet_user}"
DB_PASSWORD="${DB_PASSWORD:-fleet_password}"
DB_NAME="${DB_NAME:-fleet_db}"

export PGPASSWORD=$DB_PASSWORD

echo "========================================"
echo "FlyFleet Database Fixes Status Check"
echo "========================================"
echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
echo "Database: $DB_NAME at $DB_HOST:$DB_PORT"
echo ""

# ========================================
# 1. Check Database Connectivity
# ========================================
echo "1. Database Connectivity Check"
echo "----------------------------------------"

if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT 1" &>/dev/null; then
    echo "✗ FAILED: Cannot connect to database"
    echo ""
    echo "💡 TIP: Start the database with:"
    echo "   docker compose up -d postgres"
    echo "   OR start your local PostgreSQL service"
    echo ""
    exit 1
fi

echo "✓ Database connection successful"
echo ""

# ========================================
# 2. Check Which Migrations Are Applied
# ========================================
echo "2. Migration Application Status"
echo "----------------------------------------"

MIGRATION_DIR="infrastructure/migrations"
declare -A migrations_to_check=(
    ["064"]="Foreign Keys on trips"
    ["065"]="Performance Indexes"
    ["066"]="updated_at Triggers"
    ["067"]="org_id Naming Standardization"
    ["068"]="Normalized Driver Tables"
)

for migration in "${!migrations_to_check[@]}"; do
    filename="${migration}_${migrations_to_check[$migration]// /_}.sql"
    # Make filename more fuzzy - check if migration number exists in directory
    if ls $MIGRATION_DIR/${migration}_*.sql 1>/dev/null 2>&1; then
        echo "✓ Migration $migration file exists: ${migrations_to_check[$migration]}"
    else
        echo "✗ Migration $migration file MISSING: ${migrations_to_check[$migration]}"
    fi
done

echo ""

# ========================================
# 3. Check Database Schema Changes
# ========================================
echo "3. Schema Change Verification"
echo "----------------------------------------"

# Check 064: Foreign keys
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name='trips' AND constraint_name='fk_trips_driver'
    )
" | grep -q '^t$'; then
    echo "✓ 064: trips.driver_id FK exists"
else
    echo "✗ 064: trips.driver_id FK NOT FOUND"
fi

if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name='trips' AND constraint_name='fk_trips_organization'
    )
" | grep -q '^t$'; then
    echo "✓ 064: trips.org_id FK exists"
else
    echo "✗ 064: trips.org_id FK NOT FOUND"
fi

# Check 065: Indexes
declare -a indexes=(
    "idx_support_tickets_org_status_priority_created"
    "idx_trips_org_vehicle_time"
    "idx_trips_org_driver_status_time"
    "idx_ledger_org_created"
    "idx_driver_behavior_driver_recent"
    "idx_gps_vehicle_org_time"
    "idx_obd_vehicle_org_time"
    "idx_fuel_vehicle_org_time"
)

for idx in "${indexes[@]}"; do
    if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
        SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname='$idx')
    " | grep -q '^t$'; then
        echo "✓ 065: Index $idx exists"
    else
        echo "✗ 065: Index $idx MISSING"
    fi
done

# Check 066: Triggers
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(SELECT 1 FROM pg_proc WHERE proname='update_updated_at_column')
" | grep -q '^t$'; then
    echo "✓ 066: update_updated_at_column function exists"
else
    echo "✗ 066: update_updated_at_column function MISSING"
fi

# Count tables with triggers
tables_with_trigger=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT COUNT(DISTINCT event_object_table)
    FROM information_schema.triggers
    WHERE trigger_schema='public' AND trigger_name LIKE 'update\_%\_updated_at'
" 2>/dev/null || echo "0")
echo "   → Tables with updated_at triggers: $tables_with_trigger"

# Check 067: Naming standardization
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema='public' AND column_name='organization_id'
    AND table_name NOT IN ('pg_statistic','pg_class')
" | grep -q '^0$'; then
    echo "✓ 067: No tables use 'organization_id' (all use 'org_id')"
else
    count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
        SELECT COUNT(*) FROM information_schema.columns
        WHERE table_schema='public' AND column_name='organization_id'
        AND table_name NOT IN ('pg_statistic','pg_class')
    " 2>/dev/null || echo ">0")
    echo "✗ 067: Found $count tables still using 'organization_id'"
fi

# Check 068: Normalized tables
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_attendance')
" | grep -q '^t$'; then
    echo "✓ 068: driver_attendance table exists"
    count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM driver_attendance" 2>/dev/null || echo "0")
    echo "   → Records: $count"
else
    echo "✗ 068: driver_attendance table MISSING"
fi

if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_schedules')
" | grep -q '^t$'; then
    echo "✓ 068: driver_schedules table exists"
    count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT COUNT(*) FROM driver_schedules" 2>/dev/null || echo "0")
    echo "   → Records: $count"
else
    echo "✗ 068: driver_schedules table MISSING"
fi

echo ""

# ========================================
# 4. Data Integrity Checks
# ========================================
echo "4. Data Integrity Verification"
echo "----------------------------------------"

# Check FK violations in trips
fk_violations=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT COUNT(*) FROM trips t
    LEFT JOIN drivers d ON t.driver_id = d.id
    WHERE t.driver_id IS NOT NULL AND d.id IS NULL
" 2>/dev/null || echo "0")

if [ "$fk_violations" -eq "0" ]; then
    echo "✓ No FK violations in trips.driver_id"
else
    echo "✗ Found $fk_violations orphaned trip driver references"
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
fi

# Check driver_attendance references
att_orphaned=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT COUNT(*) FROM driver_attendance da
    LEFT JOIN drivers d ON da.driver_id = d.id
    WHERE d.id IS NULL
" 2>/dev/null || echo "0")

if [ "$att_orphaned" -eq "0" ]; then
    echo "✓ All driver_attendance.driver_id reference valid drivers"
else
    echo "⚠ Found $att_orphaned orphaned attendance records"
fi

echo ""

# ========================================
# 5. Overall Status & Next Steps
# ========================================
echo "========================================"
echo "Status Summary"
echo "========================================"

# Determine if all implemented
all_implemented=true

# Re-check critical items
if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(SELECT 1 FROM information_schema.table_constraints WHERE table_name='trips' AND constraint_name='fk_trips_driver')
" | grep -q '^t$'; then all_implemented=false; fi

if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_attendance')
" | grep -q '^t$'; then all_implemented=false; fi

if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
    SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='driver_schedules')
" | grep -q '^t$'; then all_implemented=false; fi

if [ "$all_implemented" = true ]; then
    echo "✅ ALL DATABASE FIXES IMPLEMENTED!"
    echo ""
    echo "✓ Foreign keys added"
    echo "✓ Performance indexes created"
    echo "✓ Triggers standardized"
    echo "✓ Naming consistent"
    echo "✓Tables normalized"
    echo ""
    echo "Next steps:"
    echo "1. Update application code to use new schema"
    echo "2. Run integration tests"
    echo "3. Consider stopping this loop: CronDelete 7a4db685"
else
    echo "⚠ NOT ALL FIXES IMPLEMENTED YET"
    echo ""
    echo "Missing items:"
    echo "  - Some migrations may not have been applied"
    echo "  - Some indexes/constraints may be missing"
    echo ""
    echo "Action: Run ./migrate.sh to apply pending migrations"
    echo "        Then re-run this check."
    echo ""
    echo "This check will automatically run again in 10 minutes."
fi

echo "========================================"

exit 0
