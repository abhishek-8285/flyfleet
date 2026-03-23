#!/bin/bash

# Termux Stack Starter - Full Persistence
NATS_PORT=4222
REDIS_PORT=6380
API_PORT=8081
UI_PORT=3000

mkdir -p tmp/logs tmp/invoices

echo "[$(date)] Cleaning up old processes..."
pkill -f nats-server || true
pkill -f redis-server || true
pkill -f fleet-backend || true
pkill -f serve_build || true

echo "[$(date)] Starting NATS..."
nohup nats-server -js -p $NATS_PORT > tmp/logs/nats.log 2>&1 &

echo "[$(date)] Starting Redis..."
redis-server --port $REDIS_PORT --daemonize yes

sleep 2
go run scripts/setup_nats.go

echo "[$(date)] Launching Backend..."
export PORT=$API_PORT
export DATABASE_URL="host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
export NATS_URL="nats://localhost:$NATS_PORT"
export REDIS_ADDR="localhost:$REDIS_PORT"
export INVOICE_STORAGE_PATH="./tmp/invoices"

# Start backend in a loop
nohup bash -c "while true; do ./tmp/fleet-backend >> tmp/logs/backend_persist.log 2>&1; sleep 1; done" &

echo "[$(date)] Launching UI (Go Static Server)..."
cd apps/dashboard-web
# Rebuild the Go server if needed
go build -o ../../tmp/ui-server serve_build.go
cd ../..
nohup bash -c "while true; do PORT=$UI_PORT ./tmp/ui-server >> tmp/logs/ui_persist.log 2>&1; sleep 1; done" &

echo "✅ Stack is self-healing in the background."
