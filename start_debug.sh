#!/bin/bash

# FlyFleet Debug Mode Start Script
# This script starts the backend and frontend services in high-visibility debug mode.

echo "🚀 Starting FlyFleet Debug Mode..."

# 1. Cleanup existing processes
echo "🧹 Cleaning up existing FlyFleet processes..."
pkill -f "fleet-core"
pkill -f "vite"
sleep 2

# 2. Start Backend (fleet-core)
echo "📦 Starting Backend (fleet-core) on port 8084..."
export APP_ENV=development
export LOG_LEVEL=debug
export LOG_FORMAT=console
export ENABLE_PPROF=true
export GIN_MODE=debug
export DB_URL="postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
export PORT=8084

cd services/fleet-core
nohup go run cmd/main.go > /tmp/backend_debug.log 2>&1 &
BACKEND_PID=$!
cd ../..

# 3. Start Frontend (dashboard-web)
echo "💻 Starting Frontend (dashboard-web) on port 3000..."
cd apps/dashboard-web
nohup npm run dev -- --port 3000 > /tmp/frontend_debug.log 2>&1 &
FRONTEND_PID=$!
cd ../..

# 4. Final Status
echo ""
echo "✅ FlyFleet Debug Mode is active!"
echo "--------------------------------------------------"
echo "📍 Backend Log:  /tmp/backend_debug.log"
echo "📍 Frontend Log: /tmp/frontend_debug.log"
echo "📍 PProf:        http://localhost:6060/debug/pprof/"
echo "--------------------------------------------------"
echo "Use 'tail -f /tmp/backend_debug.log' to monitor backend."
echo "Use 'tail -f /tmp/frontend_debug.log' to monitor frontend."
