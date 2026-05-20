#!/bin/bash
# ==============================================================================
# FlyFleet Native Services Control Deck (No Docker)
# ==============================================================================
# This script handles port cleanup, compiles the Go Monolith backend, 
# launches NATS and Redis, and starts the Vite frontend.
# ==============================================================================

# Set nice colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=======================================================${NC}"
echo -e "${BLUE}       🚀 Starting FlyFleet in Native Mode (No Docker)  ${NC}"
echo -e "${BLUE}=======================================================${NC}"

# 1. Graceful Port Cleanup
echo -e "\n${YELLOW}🧹 Cleaning up existing FlyFleet processes...${NC}"
pkill -f "fleet_backend" || true
pkill -f "vite" || true
pkill -f "nats-server" || true

# Helper to release ports
kill_port() {
    local port=$1
    if command -v lsof &>/dev/null; then
        local pid=$(lsof -t -i:$port)
        if [ -n "$pid" ]; then
            echo -e "   Killing process on port $port (PID: $pid)..."
            kill -9 $pid 2>/dev/null || true
        fi
    elif command -v fuser &>/dev/null; then
        fuser -k $port/tcp &>/dev/null || true
    fi
}

kill_port 8110 # Go Backend API
kill_port 3000 # Dashboard Web
kill_port 4222 # NATS Broker
kill_port 6379 # Redis DB
sleep 2

# 2. Check Database Status
echo -e "\n${YELLOW}🗄️ Checking local PostgreSQL (port 5432)...${NC}"
export PGPASSWORD=fleet_password
if pg_isready -h localhost -p 5432 &>/dev/null; then
    echo -e "   ${GREEN}✓ PostgreSQL is active on port 5432!${NC}"
else
    echo -e "   ${RED}✗ PostgreSQL is not reachable on port 5432!${NC}"
    echo -e "     Please make sure your local PostgreSQL service is started."
    echo -e "     (Command: sudo service postgresql start or pg_ctl)"
    exit 1
fi

# 3. Start NATS JetStream
echo -e "\n${YELLOW}📡 Starting NATS JetStream Broker...${NC}"
NATS_PATH="./nats-server-v2.10.11-linux-amd64/nats-server"
if [ -f "$NATS_PATH" ]; then
    nohup "$NATS_PATH" -js -p 4222 > nats.log 2>&1 &
    echo -e "   ${GREEN}✓ Started NATS from workspace package (Port 4222)${NC}"
elif command -v nats-server &>/dev/null; then
    nohup nats-server -js -p 4222 > nats.log 2>&1 &
    echo -e "   ${GREEN}✓ Started system-wide NATS Broker (Port 4222)${NC}"
else
    echo -e "   ${RED}✗ NATS Server not found! Please download NATS or install it.${NC}"
    exit 1
fi

# 4. Start Redis Cache
echo -e "\n${YELLOW}📦 Starting Redis Cache Server...${NC}"
if redis-server --port 6379 --daemonize yes 2>/dev/null; then
    echo -e "   ${GREEN}✓ Started Redis in daemon mode on port 6379${NC}"
elif command -v redis-server &>/dev/null; then
    nohup redis-server --port 6379 > redis.log 2>&1 &
    echo -e "   ${GREEN}✓ Started Redis in background on port 6379${NC}"
else
    echo -e "   ${RED}✗ Redis Server not found! Please install redis-server.${NC}"
    exit 1
fi

# 5. Database Schema Migration
echo -e "\n${YELLOW}🔄 Applying Database Migrations...${NC}"
chmod +x migrate.sh
./migrate.sh

# 6. Build and Start Backend Monolith
echo -e "\n${YELLOW}⚙️  Compiling & Launching Go Backend Monolith...${NC}"
cd services/fleet-core
echo "   Building binary..."
go build -o fleet_backend ./cmd/main.go
if [ $? -ne 0 ]; then
    echo -e "   ${RED}✗ Failed to compile Go Backend!${NC}"
    exit 1
fi

export DB_URL="postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
export PORT=8110
export APP_ENV=development
export LOG_LEVEL=debug
export REDIS_ADDR=localhost:6379
export NATS_URL=nats://localhost:4222

nohup ./fleet_backend > ../../backend.log 2>&1 &
BACKEND_PID=$!
cd ../..
echo -e "   ${GREEN}✓ Go Backend Monolith launched on port 8110 (PID: $BACKEND_PID)${NC}"

# 7. Start Frontend Dashboard Web
echo -e "\n${YELLOW}💻 Starting Frontend Dashboard Web...${NC}"
cd apps/dashboard-web
nohup npm run dev -- --port 3000 > ../../dashboard.log 2>&1 &
FRONTEND_PID=$!
cd ../..
echo -e "   ${GREEN}✓ Vite Dev Server launched on port 3000 (PID: $FRONTEND_PID)${NC}"

# 8. Services Summary
echo -e "\n${BLUE}=======================================================${NC}"
echo -e "${GREEN}✨ All FlyFleet services are starting up successfully!${NC}"
echo -e "${BLUE}=======================================================${NC}"
echo -e "📍 ${BLUE}Dashboard UI:${NC}    http://localhost:3000"
echo -e "📍 ${BLUE}Backend API Gateway:${NC} http://localhost:8110"
echo -e "📍 ${BLUE}NATS Broker:${NC}     nats://localhost:4222"
echo -e "📍 ${BLUE}Redis Server:${NC}    127.0.0.1:6379"
echo -e "📍 ${BLUE}PostgreSQL:${NC}      127.0.0.1:5432 (DB: fleet_db)"
echo -e "${BLUE}=======================================================${NC}"
echo -e "📝 ${YELLOW}Check logs at any time:${NC}"
echo -e "   - Backend logs:   ${GREEN}tail -f backend.log${NC}"
echo -e "   - Frontend logs:  ${GREEN}tail -f dashboard.log${NC}"
echo -e "   - NATS logs:      ${GREEN}tail -f nats.log${NC}"
echo -e "   - Redis logs:     ${GREEN}tail -f redis.log${NC}"
echo -e "${BLUE}=======================================================${NC}"
echo -e "💡 To stop all services later, run: ${YELLOW}pkill -f fleet_backend && pkill -f vite && pkill -f nats-server${NC}"
