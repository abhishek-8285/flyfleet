#!/bin/bash

# FlyFleet Native Startup Script (No Docker)
# This script starts all backend, frontend, and infrastructure services natively.

echo "🚀 Starting FlyFleet in Native Mode..."

# 1. Cleanup
echo "🧹 Cleaning up old processes..."
pkill -f "fleet-core" || true
pkill -f "fleet_backend" || true
pkill -f "vite" || true
pkill -f "nats-server" || true
# Force kill whatever is holding our ports
fuser -k 8084/tcp 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 2

# 2. Start Infrastructure
echo "📡 Starting NATS..."
nohup ./nats-server-v2.10.11-linux-amd64/nats-server -js -p 4222 > nats.log 2>&1 &

echo "📦 Starting Redis..."
# Try to start redis on standard port 6379 as per project configs
redis-server --port 6379 --daemonize yes || echo "Redis already running or failed to start"

# Wait for infra to initialize
sleep 2

# 3. Run Database Migrations
echo "🗄️ Running Database Migrations (in background)..."
chmod +x migrate.sh
export DB_USER=fleet_user
export DB_PASSWORD=fleet_password
export DB_NAME=fleet_db
nohup ./migrate.sh > migrate.log 2>&1 &

# 4. Start Backend (fleet-core)
echo "⚙️ Starting Backend (fleet-core)..."
export DB_URL="postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
export PORT=8084
export APP_ENV=development
export LOG_LEVEL=debug
export REDIS_ADDR=localhost:6379
export NATS_URL=nats://localhost:4222

cd services/fleet-core
# Build and run the backend (using build to avoid 'go run' internal package issues in workspaces)
go build -o fleet_backend ./cmd
nohup ./fleet_backend > ../../backend.log 2>&1 &
cd ../..

# 5. Start Frontend (dashboard-web)
echo "💻 Starting Frontend (dashboard-web)..."
cd apps/dashboard-web
# Using 'npm run dev' to start Vite server
nohup npm run dev -- --port 3000 > ../../dashboard.log 2>&1 &
cd ../..

echo "------------------------------------------------"
echo "✅ FlyFleet services are starting!"
echo "📍 Backend API:  http://localhost:8084"
echo "📍 Dashboard:    http://localhost:3000"
echo "------------------------------------------------"
echo "Check logs for details:"
echo "  - Backend:  tail -f backend.log"
echo "  - Frontend: tail -f dashboard.log"
echo "  - NATS:     tail -f nats.log"
echo "------------------------------------------------"
