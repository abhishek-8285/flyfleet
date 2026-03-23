#!/bin/bash

# 1. Cleanup old processes
echo "🧹 Cleaning up..."
pkill -9 fleet-backend || true
pkill -9 ui-server || true
pkill -9 nats-server || true
pkill -9 redis-server || true
sleep 2

# 2. Start Infrastructure
echo "🚀 Starting NATS and Redis..."
nats-server -js -p 4222 > /dev/null 2>&1 &
redis-server --port 6380 --daemonize yes
sleep 2
go run scripts/setup_nats.go

# 3. Define Database Connection
export DATABASE_URL="host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"

# 4. Start Backend (Port 8081)
echo "📦 Starting Backend on port 8081..."
nohup ./tmp/fleet-backend > tmp/logs/backend_persist.log 2>&1 &

# 5. Start UI Server (Port 3005)
echo "🎨 Starting UI on port 3005..."
nohup ./tmp/ui-server > tmp/logs/ui_persist.log 2>&1 &

echo "------------------------------------------------"
echo "✅ FlyFleet is now running!"
echo "👉 Dashboard & Designer: http://localhost:3005/billing/designer"
echo "👉 Backend API: http://localhost:8081"
echo "------------------------------------------------"
echo "Keep this Termux window open."
