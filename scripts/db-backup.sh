#!/bin/bash
# Database backup script - creates a portable dump with schema + seed data
# Usage: ./scripts/db-backup.sh [output-file]

set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-fleet_db}"

# Use superuser for backup (fleet_user may lack LOCK permissions on some tables)
# Falls back to fleet_user if superuser not available
DB_SUPER_USER="${DB_SUPER_USER:-$(whoami)}"
DB_SUPER_PASS="${DB_SUPER_PASS:-}"
DB_USER="${DB_USER:-fleet_user}"
DB_PASSWORD="${DB_PASSWORD:-fleet_password}"

if [ -n "$DB_SUPER_PASS" ]; then
    export PGPASSWORD="$DB_SUPER_PASS"
    BACKUP_USER="$DB_SUPER_USER"
else
    export PGPASSWORD="$DB_PASSWORD"
    BACKUP_USER="$DB_USER"
fi

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="${1:-backups/flyfleet_${TIMESTAMP}.sql.gz}"

mkdir -p "$(dirname "$OUTPUT_FILE")"

echo "----------------------------------------------------------------"
echo "Database Backup"
echo "----------------------------------------------------------------"
echo "Source:      $DB_NAME @ $DB_HOST:$DB_PORT (user: $BACKUP_USER)"
echo "Output:      $OUTPUT_FILE"
echo "----------------------------------------------------------------"

if ! command -v pg_dump &> /dev/null; then
    echo "Error: pg_dump not found. Install PostgreSQL client or use Docker."
    echo ""
    echo "Docker fallback:"
    echo "  docker exec fleet-postgres pg_dump -U $DB_USER -d $DB_NAME --clean --if-exists | gzip > $OUTPUT_FILE"
    exit 1
fi

echo "Dumping schema + data (compressed)..."
pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$BACKUP_USER" -d "$DB_NAME" \
    --clean --if-exists \
    --no-owner --no-privileges \
    --exclude-table-data="*_audit_log" \
    --exclude-table-data="http_request_logs" \
    --exclude-table-data="*_telemetry*" \
    2>/dev/null | gzip > "$OUTPUT_FILE"

SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
echo "✓ Backup complete: $OUTPUT_FILE ($SIZE)"
echo ""
echo "To restore on another device:"
echo "  ./scripts/db-restore.sh $OUTPUT_FILE"
