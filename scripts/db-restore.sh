#!/bin/bash
# Database restore script - recreates schema and loads data from backup
# Usage: ./scripts/db-restore.sh <backup-file>

set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5433}"
DB_USER="${DB_USER:-fleet_user}"
DB_PASSWORD="${DB_PASSWORD:-fleet_password}"
DB_NAME="${DB_NAME:-fleet_db}"

export PGPASSWORD=$DB_PASSWORD

BACKUP_FILE="${1:-}"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file>"
    echo ""
    echo "Available backups:"
    ls -lh backups/*.sql.gz 2>/dev/null || echo "  (none found in backups/)"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: File not found: $BACKUP_FILE"
    exit 1
fi

echo "----------------------------------------------------------------"
echo "Database Restore"
echo "----------------------------------------------------------------"
echo "Backup:      $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"
echo "Target:      $DB_NAME @ $DB_HOST:$DB_PORT"
echo "----------------------------------------------------------------"
echo ""
echo "⚠  This will DROP and recreate all tables in $DB_NAME"
echo "⚠  All current data will be lost!"
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 0
fi

if ! command -v psql &> /dev/null; then
    echo "Error: psql not found. Install PostgreSQL client or use Docker."
    exit 1
fi

echo ""
echo "Step 1/3: Dropping existing tables..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
DO \$\$
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END \$\$;
" 2>/dev/null
echo "✓ Tables dropped"

echo ""
echo "Step 2/3: Restoring schema + data..."
gunzip -c "$BACKUP_FILE" | psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" 2>&1 | tail -5
echo "✓ Restore complete"

echo ""
echo "Step 3/3: Running remaining migrations (if any)..."
./migrate.sh 2>&1 | tail -3
echo "✓ Migrations applied"

echo ""
echo "----------------------------------------------------------------"
echo "✓ Database restored successfully!"
echo "----------------------------------------------------------------"
