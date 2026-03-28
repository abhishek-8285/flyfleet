#!/usr/bin/env bash
#
# FlyFleet Database Backup Script
# Creates compressed backups of PostgreSQL and Redis databases
# Retention: Keep last 7 daily backups
#
# Usage: ./scripts/backup.sh [backup-dir]
#   If no backup-dir provided, creates timestamped directory under ./backups/
#
# Schedule (crontab):
#   0 2 * * * /data/data/com.termux/files/home/flyfleet/scripts/backup.sh >> /var/log/backup.log 2>&1
#
# To restore:
#   ./scripts/restore.sh /path/to/backup_directory

set -euo pipefail

BACKUP_ROOT="${1:-./backups}"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
BACKUP_DIR="${BACKUP_ROOT}/flyfleet-${TIMESTAMP}"

echo "========================================"
echo "FlyFleet Backup - ${TIMESTAMP}"
echo "========================================"
echo ""

# Create backup directory
mkdir -p "$BACKUP_DIR"
echo "✓ Created backup directory: $BACKUP_DIR"

# Load environment variables
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if docker-compose services are running
if ! docker ps | grep -q fleet-postgres; then
    echo "⚠ Warning: fleet-postgres container not running. Attempting to start..."
    docker-compose up -d postgres
    sleep 5
fi

# PostgreSQL Backup
echo ""
echo "📦 Backing up PostgreSQL..."
PG_CONTAINER="fleet-postgres"
if docker ps | grep -q "$PG_CONTAINER"; then
    # Get database credentials from environment or defaults
    PG_DB="${FLYFLEET_DB_NAME:-fleet_db}"
    PG_USER="${FLYFLEET_DB_USER:-fleet_user}"

    docker exec "$PG_CONTAINER" pg_dump -U "$PG_USER" "$PG_DB" \
        | gzip > "$BACKUP_DIR/postgres-${PG_DB}.sql.gz"

    echo "  ✓ PostgreSQL backup saved: postgres-${PG_DB}.sql.gz ($(du -h "$BACKUP_DIR/postgres-${PG_DB}.sql.gz" | cut -f1))"
else
    echo "  ⚠ PostgreSQL container not running, skipping..."
fi

# Redis Backup
echo ""
echo "📦 Backing up Redis..."
REDIS_CONTAINER="fleet-redis"
if docker ps | grep -q "$REDIS_CONTAINER"; then
    # Trigger Redis save
    docker exec "$REDIS_CONTAINER" redis-cli SAVE

    # Copy RDB file
    docker cp "$REDIS_CONTAINER:/data/dump.rdb" "$BACKUP_DIR/redis.rdb"
    echo "  ✓ Redis backup saved: redis.rdb ($(du -h "$BACKUP_DIR/redis.rdb" | cut -f1))"
else
    echo "  ⚠ Redis container not running, skipping..."
fi

# Create manifest
echo ""
echo "📝 Creating backup manifest..."
cat > "$BACKUP_DIR/manifest.json" <<EOF
{
  "timestamp": "${TIMESTAMP}",
  "backup_dir": "$BACKUP_DIR",
  "postgres_file": "postgres-${PG_DB}.sql.gz",
  "redis_file": "redis.rdb",
  "created_by": "FlyFleet Backup Script",
  "version": "1.0"
}
EOF

# Calculate total size
TOTAL_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
echo "  ✓ Manifest created"
echo "  Total backup size: $TOTAL_SIZE"

# Prune old backups (keep last 7)
echo ""
echo "🗑️  Pruning old backups (keeping last 7)..."
cd "$BACKUP_ROOT"
ls -t flyfleet-*/ | tail -n +8 | xargs -r rm -rf
echo "  ✓ Old backups removed, kept latest 7"

echo ""
echo "✅ Backup completed successfully!"
echo "   Location: $BACKUP_DIR"
echo ""
echo "To restore:"
echo "  ./scripts/restore.sh $BACKUP_DIR"
