#!/bin/bash
# FlyFleet Artifact Cleanup Script
# Safely removes build artifacts, old binaries, and temporary files
# to free up disk space without affecting source code

set -e

echo "🧹 FlyFleet Artifact Cleanup"
echo "============================"
echo ""

# Configuration
BASE_DIR="/data/data/com.termux/files/home/flyfleet"
BACKUP_DIR="$BASE_DIR/backups"
UPLOADS_DIR="$BASE_DIR/uploads"
TMP_DIR="$BASE_DIR/tmp"
PGDATA_DIR="$BASE_DIR/pgdata"
SERVICES_DIR="$BASE_DIR/services"
NATS_DIR="$BASE_DIR/nats-server-v2.10.11-linux-amd64"

# Statistics
TOTAL_FREED=0

# Function to calculate size in human readable format
human_size() {
    du -sh "$1" 2>/dev/null | cut -f1
}

# Function to add to total
add_to_total() {
    local size_bytes=$(du -sb "$1" 2>/dev/null | cut -f1)
    if [ -n "$size_bytes" ] && [ "$size_bytes" -gt 0 ]; then
        TOTAL_FREED=$((TOTAL_FREED + size_bytes))
    fi
}

# 1. Clean up log files (except active ones)
echo "📋 1/6 Cleaning up old log files..."
if [ -d "$TMP_DIR/logs" ]; then
    LOG_SIZE_BEFORE=$(du -sb "$TMP_DIR/logs" 2>/dev/null | cut -f1)
    # Remove compressed logs older than 7 days
    find "$TMP_DIR/logs" -name "*.log.gz" -mtime +7 -delete 2>/dev/null || true
    # Remove empty log files
    find "$TMP_DIR/logs" -name "*.log" -size 0 -delete 2>/dev/null || true
    LOG_SIZE_AFTER=$(du -sb "$TMP_DIR/logs" 2>/dev/null | cut -f1)
    LOG_FREED=$((LOG_SIZE_BEFORE - LOG_SIZE_AFTER))
    echo "   Freed: $(human_size $LOG_FREED)"
    TOTAL_FREED=$((TOTAL_FREED + LOG_FREED))
else
    echo "   ⚠️  No logs directory found at $TMP_DIR/logs"
fi

# 2. Clean up old binary artifacts in services/_archive
echo "🗃️  2/6 Checking archived binaries..."
if [ -d "$SERVICES_DIR/_archive" ]; then
    ARCHIVE_SIZE=$(du -sh "$SERVICES_DIR/_archive" 2>/dev/null | cut -f1)
    echo "   Current size: $ARCHIVE_SIZE"
    echo "   ⚠️  These are legacy binaries - consider removing if no longer needed"
    read -p "   Delete entire _archive directory? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$SERVICES_DIR/_archive"
        ARCHIVE_FREED=$(du -sb "$SERVICES_DIR/_archive" 2>/dev/null | cut -f1 || echo "0")
        TOTAL_FREED=$((TOTAL_FREED + ARCHIVE_FREED))
        echo "   ✅ Deleted _archive"
    else
        echo "   ⏭️  Skipped"
    fi
else
    echo "   ⚠️  No _archive directory found"
fi

# 3. Clean up PostgreSQL data (WAL files, temporary files)
echo "🗄️  3/6 Checking PostgreSQL data..."
if [ -d "$PGDATA_DIR" ]; then
    PG_SIZE=$(du -sh "$PGDATA_DIR" 2>/dev/null | cut -f1)
    echo "   PostgreSQL data size: $PG_SIZE"
    echo "   ⚠️  This contains your database - NOT recommended to delete"
    echo "   Consider: vacuumdb --analyze, or pg_wal cleanup if safe"
    read -p "   Run 'pg_wal' cleanup (removes old WAL segments)? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Only if postgres is running and we're in the container
        echo "   Skipping - requires PostgreSQL admin privileges"
        # docker exec fleet-postgres pg_archivecleanup /var/lib/postgresql/data/pg_wal 00000001 2>/dev/null || true
    else
        echo "   ⏭️  Skipped"
    fi
else
    echo "   ⚠️  No pgdata directory found"
fi

# 4. Clean up main binary and build artifacts in service directory
echo "🔨 4/6 Cleaning build artifacts..."
if [ -f "$BASE_DIR/services/fleet-core/main" ]; then
    BIN_SIZE=$(du -sh "$BASE_DIR/services/fleet-core/main" | cut -f1)
    echo "   main binary size: $BIN_SIZE"
    read -p "   Delete main binary? (You'll need to rebuild) (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -f "$BASE_DIR/services/fleet-core/main"
        add_to_total "$BASE_DIR/services/fleet-core/main"
        echo "   ✅ Deleted main binary"
    else
        echo "   ⏭️  Skipped"
    fi
fi

# Clean up bin directory
if [ -d "$BASE_DIR/services/fleet-core/bin" ]; then
    BIN_DIR_SIZE=$(du -sh "$BASE_DIR/services/fleet-core/bin" | cut -f1)
    echo "   bin/ directory size: $BIN_DIR_SIZE"
    read -p "   Delete bin/ directory? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$BASE_DIR/services/fleet-core/bin"
        add_to_total "$BASE_DIR/services/fleet-core/bin"
        echo "   ✅ Deleted bin/ directory"
    else
        echo "   ⏭️  Skipped"
    fi
fi

# 5. Clean up NATS downloaded binary
echo "📦 5/6 Checking downloaded binaries..."
if [ -d "$NATS_DIR" ]; then
    NATS_SIZE=$(du -sh "$NATS_DIR" | cut -f1)
    echo "   NATS server size: $NATS_SIZE"
    read -p "   Delete NATS server binary? (Will need to re-download) (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$NATS_DIR"
        add_to_total "$NATS_DIR"
        echo "   ✅ Deleted NATS server"
    else
        echo "   ⏭️  Skipped"
    fi
fi

# 6. Clean up node_modules/.cache if huge
echo "📦 6/6 Checking npm cache..."
if [ -d "$BASE_DIR/node_modules" ]; then
    NPM_SIZE=$(du -sh "$BASE_DIR/node_modules" 2>/dev/null | cut -f1)
    echo "   node_modules size: $NPM_SIZE"
    echo "   ⚠️  Deleting node_modules requires npm install to rebuild"
    read -p "   Delete node_modules? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$BASE_DIR/node_modules"
        add_to_total "$BASE_DIR/node_modules"
        echo "   ✅ Deleted node_modules (run 'npm install' to restore)"
    else
        echo "   ⏭️  Skipped"
    fi
fi

# Summary
echo ""
echo "✅ Cleanup complete!"
echo ""
HUMAN_TOTAL=$(numfmt --to=iec-i --suffix=B $TOTAL_FREED 2>/dev/null || echo "${TOTAL_FREED}B")
echo "📊 Total space freed: $HUMAN_TOTAL"
echo ""
echo "💾 Additional recommendations:"
echo "   - Move large directories to external storage if space is critical"
echo "   - Use 'find . -size +100M -type f' to locate other large files"
echo "   - Consider using Docker volumes for pgdata to keep on separate disk"
echo ""
