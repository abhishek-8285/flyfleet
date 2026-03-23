#!/bin/bash
# scripts/check_migrations.sh
# Enforces safe PostgreSQL migration patterns for the Monolith.
# Guardrails:
# 1. EVERY migration file must contain `SET lock_timeout = '2s';`
# 2. No `ALTER TABLE ... ADD COLUMN ... DEFAULT` on hyper-scale tables.

MIGRATIONS_DIR="infrastructure/migrations"
TARGETS="(gps_locations|vehicles|telemetry|fastag_transactions)"

# If directory doesn't exist, just exit cleanly (maybe no migrations yet)
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "No migrations directory found. Skipping."
    exit 0
fi

# 1. Check for missing lock timeouts
for idx_file in $MIGRATIONS_DIR/*.sql; do
    # Skip if wildcard didn't match anything
    [ -e "$idx_file" ] || continue
    
    if ! grep -Eq "SET\s+lock_timeout\s*=\s*'2s';" "$idx_file"; then
        echo "❌ CRITICAL: Missing lock_timeout in $idx_file!"
        echo "Every migration must explicitly enforce a 2-second lock timeout."
        echo "If a migration cannot instantly acquire an ACCESS EXCLUSIVE lock on a heavily used table,"
        echo "it will queue up, and subsequently block ALL other queries attempting to read from that table."
        echo "This causes immediate, cascading database connection exhaustion."
        echo ""
        echo "Fix: Add the following line at the top of your migration file:"
        echo "SET lock_timeout = '2s';"
        echo ""
        echo "If the migration fails in CI due to the timeout, it means the table is perfectly hot."
        echo "The CI pipeline should retry the deployment automatically. Do NOT increase the timeout!"
        exit 1
    fi
done

# 2. Check for full-table rewrites (Lazy Default Guardrail)
BAD_MIGRATIONS=$(grep -Ei "ALTER\s+TABLE\s+${TARGETS}.*ADD\s+COLUMN.*DEFAULT" $MIGRATIONS_DIR/*.sql 2>/dev/null)

if [ -n "$BAD_MIGRATIONS" ]; then
    echo "❌ CRITICAL: Found full-table rewrite in migration!"
    echo "Adding a DEFAULT value to a new column requires rewriting the entire table."
    echo "This will acquire an ACCESS EXCLUSIVE LOCK on hyper-scale tables, causing global downtime."
    echo "Offending files:"
    echo "$BAD_MIGRATIONS"
    echo ""
    echo "Fix using the Lazy Default Pattern:"
    echo "  -- Step 1: Add the column allowing NULLs."
    echo "  ALTER TABLE my_table ADD COLUMN new_col VARCHAR(50);"
    echo "  -- Step 2: Set the default for FUTURE inserts only."
    echo "  ALTER TABLE my_table ALTER COLUMN new_col SET DEFAULT 'value';"
    echo "  -- Step 3: Backfill old rows asynchronously in 5k chunks via a Go worker."
    exit 1
fi

echo "✅ Standard Migration lock-guards passed."
exit 0
