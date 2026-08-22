#!/bin/bash
# Preflight script: Run backend to trigger harmonizer, then apply fixes
set -e

export PREFIX=/data/data/com.termux/files/usr
export PATH="$PREFIX/go1.26/go/bin:$PREFIX/bin:$PATH"
export HOME=/data/data/com.termux/files/home
export GOCACHE="$PREFIX/tmp/go-build-cache"
export GOMODCACHE="$PREFIX/tmp/pkg/mod"
export GOPATH="$PREFIX/tmp/gopath"
export GOTOOLCHAIN=local
export GOTMPDIR="$PREFIX/tmp"
export GOSUMDB=off
export GONOSUMCHECK=*
export SSL_CERT_FILE="$PREFIX/etc/tls/cert.pem"
export CGO_ENABLED=0
export GOWORK=off
export DATABASE_URL="host=$PREFIX/tmp user=u0_a231 dbname=fleet_db sslmode=disable"
export PG_SOCKET_DIR="$PREFIX/tmp"
FLEET_CORE_DIR="$HOME/flyfleet/services/fleet-core"

# Function to run harmonizer and return exit code
run_backend_briefly() {
    cd "$FLEET_CORE_DIR"
    timeout 30 ./fleet-core 2>&1 || true
}

# Step 1: Reset database
echo "Step 1: Resetting database..."
psql -h "$PG_SOCKET_DIR" -d fleet_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO u0_a231; GRANT ALL ON SCHEMA public TO public;" 2>&1 | tail -1

# Step 2: Run backend briefly to trigger harmonizer
echo "Step 2: Running harmonizer..."
run_backend_briefly 2>&1 | grep -E "harmoniz|Stabiliz|migration" | head -5

# Step 3: Apply comprehensive preflight SQL
echo "Step 3: Applying preflight schema..."
psql -h "$PG_SOCKET_DIR" -d fleet_db -f "$FLEET_CORE_DIR/preflight_schema.sql" 2>&1 | grep -c "ALTER TABLE" | head -1

# Step 4: Clean up failed migration markers
echo "Step 4: Cleaning up migration state..."
psql -h "$PG_SOCKET_DIR" -d fleet_db -c "TRUNCATE schema_migrations_legacy, schema_migrations;" 2>&1 | tail -1

# Step 5: Run migrations
echo "Step 5: Running migrations..."
cd "$FLEET_CORE_DIR"
go run ./cmd/migrate up 2>&1 | tail -10

echo "Done"
