#!/bin/bash
set -e

# Cleanup first
echo "🧹 Cleaning up existing processes..."
# Kill by port to be sure
for port in {8080..8090} {50051..50061}; do
  lsof -i :$port -t | xargs -r kill -9 || true
done

# Shared Env - require database credentials
: "${DATABASE_URL:?ERROR: DATABASE_URL environment variable is required. Copy .env.example to .env and set your database credentials.}"
export DATABASE_URL="${DATABASE_URL}"
export JWT_SECRET="${JWT_SECRET:-change-me-to-strong-random-secret-$(openssl rand -hex 32 2>/dev/null || echo 'RANDOM_SECRET')}"
export NATS_URL="nats://localhost:4222"
export REDIS_ADDR="localhost:6379"

# Billing Service Specifics - derive from DATABASE_URL or require explicit vars
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-fleet_user}"
# If DB_PASS not set, try to extract from DATABASE_URL (not ideal but supports existing patterns)
if [ -z "${DB_PASS}" ]; then
  # Extract password from DATABASE_URL format: postgres://user:password@host:port/db
  DB_PASS=$(echo "${DATABASE_URL}" | sed -n 's|postgres://[^:]*:\([^@]*\)@.*|\1|p')
  if [ -z "${DB_PASS}" ]; then
    echo "ERROR: DB_PASS must be set or DATABASE_URL must contain password"
    exit 1
  fi
fi
export DB_PASS
export DB_NAME="${DB_NAME:-fleet_db}"
export SSL_MODE="disable"

# Logs
mkdir -p tmp/logs

echo "🚀 Starting services..."

# Helper function to start service
start_service() {
    name=$1
    dir=$2
    http_port=$3
    grpc_port=$4
    
    echo "  - $name Service (HTTP:$http_port, gRPC:$grpc_port)..."
    (
        cd $dir
        export PORT=$http_port
        export GRPC_PORT=$grpc_port
        # Specific overrides for services that use different env var names
        if [ "$name" == "billing" ]; then
             export PORT=$http_port # Billing uses int, but shell export is string, viper handles it
        fi
        
        go run cmd/main.go > ../../tmp/logs/$name.log 2>&1 &
    )
}

# Start services
start_service "auth" "services/auth-service" 8080 50051
start_service "tracking" "services/tracking-service" 8081 50052
start_service "trip" "services/trip-service" 8082 50053
start_service "alert" "services/alert-service" 8083 50054
start_service "billing" "services/billing-service" 8084 50055
start_service "admin" "services/admin-support-service" 8085 50056
start_service "maintenance" "services/maintenance-service" 8086 50057
start_service "driver" "services/driver-service" 8087 50058
start_service "analytics" "services/analytics-service" 8088 50059
start_service "reports" "services/reports-service" 8089 50060
start_service "fleet" "services/fleet-core-service-legacy" 8090 50061

echo "⏳ Waiting for services to stabilize (25s)..."
sleep 25

echo "✅ Local stack started."
echo "   Run 'python3 scripts/validate_grpc_platform.py' to verify."
