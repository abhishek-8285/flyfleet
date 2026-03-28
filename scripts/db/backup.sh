#!/bin/bash
set -e

# Configuration - require explicit DB_PASSWORD, no defaults
: "${DB_PASSWORD:?ERROR: DB_PASSWORD environment variable is required. Set it with your database password.}"
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5433}
DB_USER=${DB_USER:-fleet_user}
DB_NAME=${DB_NAME:-fleet_db}
BACKUP_DIR=${BACKUP_DIR:-./backups}
DATE=$(date +%Y-%m-%d_%H-%M-%S)
BACKUP_FILE="$BACKUP_DIR/$DB_NAME-$DATE.sql.gz"

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

echo "Starting logical backup of $DB_NAME to $BACKUP_FILE..."

# Execute pg_dump using PGPASSWORD from environment
PGPASSWORD="${DB_PASSWORD}" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"

echo "Backup successful: $BACKUP_FILE"
