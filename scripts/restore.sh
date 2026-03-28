#!/usr/bin/env bash
#
# FlyFleet Database Restore Script
#
# Usage: ./scripts/restore.sh <backup_directory>
# Example: ./scripts/restore.sh ./backups/flyfleet-2025-03-26_02-30-00
#
# WARNING: This will OVERWRITE existing data!
# Make sure services are stopped or in maintenance mode.

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 <backup_directory>"
    echo "Example: $0 ./backups/flyfleet-2025-03-26_02-30-00"
    exit 1
fi

BACKUP_DIR="$1"

if [ ! -d "$BACKUP_DIR" ]; then
    echo "Error: Backup directory not found: $BACKUP_DIR"
    exit 1
fi

# Check for manifest
if [ ! -f "$BACKUP_DIR/manifest.json" ]; then
    echo "Error: Invalid backup directory (missing manifest.json)"
    exit 1
fi

echo "========================================"
echo "FlyFleet Database Restore"
echo "========================================"
echo "Backup: $BACKUP_DIR"
echo ""

# Load environment variables
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Confirmation
echo "⚠️  WARNING: This will DESTROY current data and replace it with backup!"
echo ""
read -p "Are you absolutely sure? (type 'YES' to confirm): " CONFIRM
if [ "$CONFIRM" != "YES" ]; then
    echo "Restore cancelled."
    exit 0
fi

# Stop services to prevent writes during restore
echo ""
echo "Stopping services..."
docker-compose stop
echo "✓ Services stopped"

# Restore PostgreSQL
PG_FILE="$BACKUP_DIR/postgres-*.sql.gz"
if [ -f "$PG_FILE" ]; then
    echo ""
    echo "📦 Restoring PostgreSQL..."
    PG_CONTAINER="fleet-postgres"
    PG_DB="${FLYFLEET_DB_NAME:-fleet_db}"
    PG_USER="${FLYFLEET_DB_USER:-fleet_user}"

    # Wait for PostgreSQL to be ready
    echo "  Waiting for PostgreSQL to start..."
    docker-compose start postgres
    sleep 10

    # Drop and recreate database
    echo "  Dropping and recreating database..."
    docker exec "$PG_CONTAINER" psql -U postgres -c "DROP DATABASE IF EXISTS $PG_DB;"
    docker exec "$PG_CONTAINER" psql -U postgres -c "CREATE DATABASE $PG_DB OWNER $PG_USER;"

    # Restore from backup
    echo "  Restoring from backup..."
    gunzip -c "$PG_FILE" | docker exec -i "$PG_CONTAINER" psql -U "$PG_USER" "$PG_DB"
    echo "  ✓ PostgreSQL restored"
else
    echo "  ⚠ No PostgreSQL backup found, skipping..."
fi

# Restore Redis
REDIS_FILE="$BACKUP_DIR/redis.rdb"
if [ -f "$REDIS_FILE" ]; then
    echo ""
    echo "📦 Restoring Redis..."
    REDIS_CONTAINER="fleet-redis"

    # Copy RDB file
    docker cp "$REDIS_FILE" "$REDIS_CONTAINER:/data/dump.rdb"

    # Fix permissions
    docker exec "$REDIS_CONTAINER" chown redis:redis /data/dump.rdb

    echo "  ✓ Redis RDB file restored"
else
    echo "  ⚠ No Redis backup found, skipping..."
fi

# Start services
echo ""
echo "Starting services..."
docker-compose up -d
echo "✓ Services started"

echo ""
echo "✅ Restore completed successfully!"
echo ""
echo "Verify:"
echo "  - Check logs: docker-compose logs -f"
echo "  - Test API: curl http://localhost:8080/healthz"
echo "  - Check data: connect to database and verify records"
