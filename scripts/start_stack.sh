#!/bin/bash

# Termux Stack Starter - Fixed for Conflict Handling
NATS_PORT=4222
REDIS_PORT=6380
API_PORT=8081

mkdir -p tmp/logs tmp/invoices

echo "[$(date)] Checking for conflicts on port $API_PORT..."
# Find and kill anything on our port
CONF_PID=$(lsof -t -i:$API_PORT)
if [ ! -z "$CONF_PID" ]; then
    echo "Killing conflicting process $CONF_PID"
    kill -9 $CONF_PID || true
fi

echo "[$(date)] Starting NATS..."
nats-server -js -p $NATS_PORT > tmp/logs/nats.log 2>&1 &

echo "[$(date)] Starting Redis..."
redis-server --port $REDIS_PORT --daemonize yes

sleep 2
go run scripts/setup_nats.go

echo "[$(date)] Launching Optimized Backend..."
export PORT=$API_PORT
export DATABASE_URL="host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
export NATS_URL="nats://localhost:$NATS_PORT"
export REDIS_ADDR="localhost:$REDIS_PORT"
export INVOICE_STORAGE_PATH="./tmp/invoices"

# Run direct binary for speed and stability
./tmp/fleet-backend > tmp/logs/backend_persist.log 2>&1
