#!/bin/bash
# scripts/check_concurrent_index.sh
# Ensures all CREATE INDEX commands on the monolithic database use the CONCURRENTLY keyword.

MIGRATIONS_DIR="infrastructure/migrations"

# Find any CREATE INDEX that doesn't include CONCURRENTLY
# We use regex to handle variable whitespace and case.
# Note: CONCURRENTLY must be built outside a transaction block.
BAD_INDEXES=$(grep -Ei "CREATE\s+(UNIQUE\s+)?INDEX" $MIGRATIONS_DIR/*.sql 2>/dev/null | grep -Evi "CONCURRENTLY")

if [ -n "$BAD_INDEXES" ]; then
    echo "❌ CRITICAL: Found CREATE INDEX without CONCURRENTLY!"
    echo "Creating a standard index acquires a ShareLock on the table, blocking all writes (INSERT/UPDATE/DELETE)."
    echo "For hyper-scale tables like gps_locations or fastag_transactions, this causes complete downtime."
    echo ""
    echo "Offending files:"
    echo "$BAD_INDEXES"
    echo ""
    echo "Fix: Use CREATE INDEX CONCURRENTLY."
    echo "Note: CONCURRENTLY builds cannot run inside a standard BEGIN/COMMIT transaction block."
    echo "If your migration tool wraps statements in transactions by default, you may need to disable transactions for this specific migration file (e.g., using a tool-specific annotation like '-- +goose NO TRANSACTION')."
    exit 1
fi

echo "✅ Concurrent Index linter passed."
exit 0
