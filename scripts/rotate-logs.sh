#!/bin/bash
# Manual log rotation script for FlyFleet
# Runs outside of logrotate system (e.g., on Termux or custom setups)

set -e

LOGS_DIR="/data/data/com.termux/files/home/flyfleet/tmp/logs"
RETENTION_DAYS=7
DATE=$(date +%Y-%m-%d)

echo "🗂️  Rotating logs in $LOGS_DIR..."

if [ ! -d "$LOGS_DIR" ]; then
    echo "⚠️  Logs directory does not exist: $LOGS_DIR"
    exit 0
fi

# Create archive directory
ARCHIVE_DIR="$LOGS_DIR/archive/$DATE"
mkdir -p "$ARCHIVE_DIR"

# Find and rotate .log files
find "$LOGS_DIR" -maxdepth 1 -name "*.log" -type f | while read -r logfile; do
    filename=$(basename "$logfile")
    echo "  Rotating: $filename"

    # Compress and move to archive
    gzip -c "$logfile" > "$ARCHIVE_DIR/$filename.gz"

    # Truncate original log file (keep file for continued writing)
    : > "$logfile"

    echo "    ✓ Archived to: $ARCHIVE_DIR/$filename.gz"
done

# Delete old archives (> RETENTION_DAYS)
echo "  Cleaning up archives older than $RETENTION_DAYS days..."
find "$LOGS_DIR/archive" -maxdepth 1 -type d -mtime +$RETENTION_DAYS | while read -r olddir; do
    echo "    Removing: $olddir"
    rm -rf "$olddir"
done

# Summary
ARCHIVE_SIZE=$(du -sh "$LOGS_DIR/archive" 2>/dev/null | cut -f1)
echo ""
echo "✅ Log rotation complete!"
echo "   Current archive: $ARCHIVE_DIR"
echo "   Total archive size: $ARCHIVE_SIZE"
echo "   Active logs: $LOGS_DIR/*.log (truncated, ready for new writes)"
echo ""
