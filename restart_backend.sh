#!/bin/bash
set -e

echo "Stopping stale processes..."
pkill -9 -f air || true
pkill -9 -f fleet-core || true
fuser -k 8084/tcp || true

# Install air if missing
GOPATH=$(go env GOPATH)
GOBIN=${GOPATH}/bin
PATH=$PATH:$GOBIN

if ! command -v air &> /dev/null; then
    if [ -f "$GOBIN/air" ]; then
        AIR_CMD="$GOBIN/air"
    else
        echo "Installing air..."
        go install github.com/air-verse/air@latest
        AIR_CMD="$GOBIN/air"
    fi
else
    AIR_CMD=$(command -v air)
fi

echo "Verifying environment (Air: $AIR_CMD)..."
cd services/fleet-core

# Force a clean build to verify compilation
echo "Checking compilation..."
go build -o bin/fleet-core ./cmd

if [ $? -eq 0 ]; then
    echo "Compilation successful. Launching with hot-reload..."
    $AIR_CMD > air.log 2>&1 &
else
    echo "Compilation FAILED. Check errors in terminal."
    exit 1
fi

echo "Waiting for service to start on port 8084..."
for i in {1..10}; do
    if curl -s http://localhost:8084/health | grep -q "healthy"; then
        echo "Backend is UP and healthy!"
        exit 0
    fi
    echo "Still waiting... ($i/10)"
    sleep 2
done

echo "Backend failed to start or health check timed out."
tail -n 20 air.log
exit 1
