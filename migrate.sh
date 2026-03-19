#!/bin/bash

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-fleet_user}"
DB_PASSWORD="${DB_PASSWORD:-fleet_password}"
DB_NAME="${DB_NAME:-fleet_db}"

# Export PGPASSWORD to avoid prompting
export PGPASSWORD=$DB_PASSWORD

echo "----------------------------------------------------------------"
echo "Starting Database Migration..."
echo "Database: $DB_NAME at $DB_HOST:$DB_PORT"
echo "----------------------------------------------------------------"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql command not found. Please install PostgreSQL client."
    exit 1
fi

# Directory containing migration files
MIGRATION_DIR="infrastructure/migrations"

# Check if directory exists
if [ ! -d "$MIGRATION_DIR" ]; then
    echo "Error: Migration directory $MIGRATION_DIR not found!"
    exit 1
fi

# Loop through SQL files in alphabetical order
for file in $(ls $MIGRATION_DIR/*.sql | sort); do
    filename=$(basename "$file")
    echo "Applying migration: $filename..."
    
    # Run the SQL file
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$file"
    
    if [ $? -eq 0 ]; then
        echo "✓ Successfully applied $filename"
    else
        echo "✗ Failed to apply $filename"
        # We don't exit on failure to allow idempotent reruns (if using IF NOT EXISTS)
        # But in a strict environment, we might want to exit.
        # For now, we continue.
    fi
done

echo "----------------------------------------------------------------"
echo "Migration Complete."
echo "----------------------------------------------------------------"
