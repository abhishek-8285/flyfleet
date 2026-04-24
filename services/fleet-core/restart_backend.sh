#!/bin/bash
# Force kill any process on 8085
fuser -k 8085/tcp || true
pkill -f "go run cmd/main.go" || true

echo "Cleaning up port 8084..."
fuser -k 8084/tcp || true

echo "Starting fleet-core on port 8084 using go run..."
export PORT=8084
export TELEMETRY_HMAC_SECRET=FF-SECRET-2026
export JWT_SECRET=CHANGE_ME_GENERATE_A_STRONG_KEY_32
nohup go run ./cmd > fleet-manual.log 2>&1 &

echo "Service started on port 8084. Check fleet-manual.log for progress."
