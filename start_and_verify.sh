#!/bin/bash

# Cleanup
echo "Cleaning up..."
pkill -f web-bff || true
pkill -f fleet-core-service || true
pkill -f admin-support-service || true

# Build Services (Ensure binaries are fresh)
echo "Building Services..."
(cd services/microservices/web-bff && go build -o web-bff cmd/main.go)
(cd services/fleet-core-service && go build -o fleet-core-service cmd/main.go)
(cd services/microservices/admin-support-service && go build -o admin-support-service cmd/main.go)


# Run Database Migrations
echo "Running Database Migrations..."
./migrate.sh

# Start Web-BFF
echo "Starting Web-BFF on port 8088..."
export PORT=8088
export FLEET_CORE_SERVICE_URL="http://localhost:8084"
export ADMIN_SUPPORT_SERVICE_URL="http://localhost:8086"
export TRACKING_SERVICE_URL="http://localhost:8082"
export ALERT_SERVICE_URL="http://localhost:8085"
export BILLING_SERVICE_URL="http://localhost:8087"
export AUTH_SERVICE_URL="http://localhost:8080" # Not running, but avoids DNS lookup delay

cd services/microservices/web-bff
./web-bff > web-bff.log 2>&1 &
PID=$!
echo "Web-BFF PID: $PID"

# Wait for startup
sleep 5

# Check if running
if ps -p $PID > /dev/null; then
   echo "Web-BFF is running."
   echo "Checking Health..."
   curl -v http://localhost:8088/health
   echo ""
   
   # Run verification
   echo "Running full verification..."
   cd ../../
   ./verify_all.sh
else
   echo "Web-BFF died. Logs:"
   cat web-bff.log
fi
