#!/bin/bash
set -e

echo "Stopping stale processes..."
pkill -9 -f air || true
pkill -9 -f fleet-core || true
pkill -9 -f 'go run' || true
fuser -k 8084/tcp || true
fuser -k 8100/tcp || true
sleep 1

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

# Run database migrations
echo "Running database migrations..."
go run ./cmd/migrate up

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

# Detect port from .env
TARGET_PORT=$(grep "^PORT=" .env | cut -d'=' -f2)
TARGET_PORT=${TARGET_PORT:-8084}
export PORT=$TARGET_PORT

echo "Waiting for service to start on port $TARGET_PORT..."
for i in {1..20}; do
    if curl -s http://localhost:$TARGET_PORT/health | grep -q "healthy"; then
        echo "Backend is UP and healthy on port $TARGET_PORT!"
        exit 0
    fi
    echo "Still waiting... ($i/20)"
    sleep 2
done

echo "Backend failed to start or health check timed out."
tail -n 20 air.log
exit 1
