#!/usr/bin/env bash
#
# FlyFleet Admin User Creation Script
# Creates a super_admin user in the database
#
# Usage: ./scripts/create_admin.sh [email] [password] [name]
#   Default: admin@flyfleet.com admin123 "FlyFleet Admin"
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$PROJECT_ROOT"

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5433}"
DB_USER="${POSTGRES_USER:-fleet_user}"
DB_PASSWORD="${POSTGRES_PASSWORD:-fleet_password}"
DB_NAME="${POSTGRES_DB:-fleet_db}"

# Defaults
EMAIL="${1:-admin@flyfleet.com}"
PASSWORD="${2:-Admin123!}"
NAME="${3:-FlyFleet Admin}"

echo "========================================"
echo "FlyFleet Admin User Creation"
echo "========================================"
echo ""
echo "Database: $DB_NAME @ $DB_HOST:$DB_PORT"
echo "User:     $EMAIL"
echo "Name:     $NAME"
echo ""

# Check if docker is running and container exists
if command -v docker &>/dev/null; then
  if ! docker ps | grep -q fleet-postgres; then
    echo "Error: PostgreSQL container is not running."
    echo "Please start it with: docker-compose up -d postgres"
    exit 1
  fi
  # Execute inside container
  EXEC="docker exec fleet-postgres psql -U $DB_USER -d $DB_NAME -v ON_ERROR_STOP=1"
else
  # Use local psql
  EXEC="PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -v ON_ERROR_STOP=1"
fi

echo "Creating admin user..."

# SQL to create admin user
$EXEC <<SQL
-- Create a default organization if none exists
DO \$\$
DECLARE
  org_id UUID;
BEGIN
  -- Get or create default organization
  SELECT id INTO org_id FROM organizations WHERE name = 'Default Organization' LIMIT 1;
  IF org_id IS NULL THEN
    INSERT INTO organizations (name) VALUES ('Default Organization') RETURNING id INTO org_id;
  END IF;

  -- Check if user already exists
  IF EXISTS (SELECT 1 FROM users WHERE email = '${EMAIL}') THEN
    RAISE NOTICE 'User % already exists', '${EMAIL}';
  ELSE
    -- Insert admin user with super_admin role
    -- Password will be hashed by the application, but for direct DB insert we need to provide a placeholder
    -- The user should change password on first login via /admin/users/{id}/reset-password
    INSERT INTO users (
      id,
      org_id,
      email,
      password_hash,
      role,
      is_active,
      created_at,
      updated_at,
      mfa_enabled,
      mfa_methods,
      last_password_change,
      permission_mask
    ) VALUES (
      gen_random_uuid(),
      org_id,
      '${EMAIL}',
      'CHANGE_ME_INITIAL_PLACEHOLDER',  -- Placeholder - must be set via API
      'super_admin',
      true,
      CURRENT_TIMESTAMP,
      CURRENT_TIMESTAMP,
      false,
      '{}',
      CURRENT_TIMESTAMP,
      9223372036854775807  -- All permissions (max BIGINT)
    );

    RAISE NOTICE 'Admin user % created with role super_admin', '${EMAIL}';
    RAISE NOTICE 'You must now set the password via the API or application.';
  END IF;
END
\$\$;
SQL

echo ""
echo "✅ Admin user creation initiated."
echo ""
echo "⚠️  IMPORTANT: The password is set to a placeholder. You must:"
echo "   1. Start the application: docker-compose up -d"
echo "   2. Use the password reset API:"
echo "      POST /api/v1/auth/password/reset/request"
echo "      { \"email\": \"${EMAIL}\" }"
echo "   3. Follow the reset link sent to the email (check logs for token)"
echo ""
echo "To view the user:"
echo "  docker exec fleet-postgres psql -U $DB_USER -d $DB_NAME -c \"SELECT id, email, role, org_id FROM users WHERE email='${EMAIL}';\""
echo ""
