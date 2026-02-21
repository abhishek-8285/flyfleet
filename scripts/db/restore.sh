#!/bin/bash
set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5433}
DB_USER=${DB_USER:-fleet_user}
DB_NAME=${DB_NAME:-fleet_db}

if [ -z "$1" ]; then
  echo "Usage: $0 <backup_file.sql.gz>"
  exit 1
fi

BACKUP_FILE="$1"

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Error: Backup file $BACKUP_FILE does not exist."
  exit 1
fi

echo "Starting restoration of $DB_NAME from $BACKUP_FILE..."

# Restore data via psql utilizing zcat since it's gzipped
zcat "$BACKUP_FILE" | PGPASSWORD=${DB_PASSWORD:-fleet_password} psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"

echo "Restoration completed successfully."
