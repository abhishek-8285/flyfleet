#!/bin/bash

# FlyFleet Migration Script: Docker -> Native Local
# This script handles the installation, configuration, and data migration.

set -e

echo "🚀 Starting FlyFleet Native Migration..."

# 1. System Updates & Repository Fixes
echo "📦 Fixing APT repositories..."
sudo rm -f /etc/apt/sources.list.d/timescaledb.list /etc/apt/sources.list.d/timescale_timescaledb.list
(curl -s https://packagecloud.io/install/repositories/timescale/timescaledb/script.deb.sh | sudo bash) || true

# 2. Extension Installation (PG 16)
echo "🧱 Fixing package manager and resolving conflicts..."
# Aggressively remove any standard repo versions
sudo apt-get purge -y postgresql-16-timescaledb postgresql-18-timescaledb || true
sudo apt-get autoremove -y || true
sudo apt-get --fix-broken install -y

echo "🧱 Installing clean TimescaleDB and Toolkit for Postgres 16..."
# Install for the ACTIVE version (16)
sudo apt-get install -y timescaledb-2-2.26.3-postgresql-16 timescaledb-2-loader-postgresql-16 timescaledb-toolkit-postgresql-16

# Enable TimescaleDB in postgresql.conf if not already there
PG_CONF="/etc/postgresql/16/main/postgresql.conf"
if ! grep -q "timescaledb" "$PG_CONF"; then
    echo "Configuring Postgres 16 to load timescaledb..."
    sudo sed -i "s/#shared_preload_libraries = ''/shared_preload_libraries = 'timescaledb'/g" "$PG_CONF"
    sudo systemctl restart postgresql
fi

# 3. Role & Database Preparation
echo "🐘 Preparing Postgres environment..."
sudo -u postgres psql -c "REVOKE CONNECT ON DATABASE fleet_db FROM public;" || true
sudo -u postgres psql -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'fleet_db' AND pid <> pg_backend_pid();" || true
sudo -u postgres psql -c "DROP DATABASE IF EXISTS fleet_db;" || true
sudo -u postgres psql -c "DROP USER IF EXISTS fleet_user;" || true
sudo -u postgres psql -c "DROP USER IF EXISTS flyfleet_app;" || true

sudo -u postgres psql -c "CREATE USER flyfleet_app WITH LOGIN;" || true
sudo -u postgres psql -c "CREATE USER fleet_user WITH PASSWORD 'fleet_password' SUPERUSER;"
sudo -u postgres psql -c "CREATE DATABASE fleet_db OWNER fleet_user;"

# 4. Dump Sanitization & Restore
echo "💾 Sanity checking and restoring data (v3 sanitization)..."
BACKUP_FILE="backups/last/fleet_db-20260313-000000.sql.gz"
SANITIZED_FILE="/tmp/fleet_db_sanitized.sql"

if [ -f "$BACKUP_FILE" ]; then
    echo "Stripping restrictions and incompatible 'ONLY' keywords..."
    # 1. Remove \restrict line
    # 2. Remove 'ALTER TABLE ONLY' (as v18 rejects ONLY on hypertables)
    # 3. Create necessary extensions
    zcat "$BACKUP_FILE" | grep -v "^\\\\restrict" | sed 's/ALTER TABLE ONLY/ALTER TABLE/g' > "$SANITIZED_FILE"
    
    echo "Restoring to local fleet_db..."
    PGPASSWORD=fleet_password psql -h localhost -U fleet_user -d fleet_db -f "$SANITIZED_FILE" -v ON_ERROR_STOP=0 -q
    rm -f "$SANITIZED_FILE"
else
    echo "⚠️ Warning: Backup file not found at $BACKUP_FILE."
fi

# 5. Native Services Setup
echo "⚡ Starting Redis and NATS..."
sudo systemctl enable --now redis-server
echo "🚀 Starting NATS server with JetStream..."
nohup nats-server -js -p 4222 > nats.log 2>&1 &

# 6. Docker Total Purge
echo "🗑️ Purging Docker.io and Docker-Compose..."
sudo apt-get purge -y docker.io docker-compose python3-compose python3-docker
sudo apt-get autoremove -y

# 7. Verification
echo "------------------------------------------------"
echo "✅ Migration Complete!"
echo "👉 Postgres: localhost:5432 (fleet_db)"
echo "👉 Redis: localhost:6379"
echo "👉 NATS: localhost:4222"
echo "------------------------------------------------"
echo "Please restart your terminal to ensure all changes take effect."
