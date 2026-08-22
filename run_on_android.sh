#!/bin/bash
# FlyFleet startup script for running on Android via Termux
# Run as the Termux user (u0_a231), NOT as root

# Set up Termux environment paths
export PREFIX=/data/data/com.termux/files/usr
export PATH="$PREFIX/go1.26/go/bin:$PREFIX/bin:$PATH"
export HOME=/data/data/com.termux/files/home
export GOCACHE="$PREFIX/tmp/go-build-cache"
export GOMODCACHE="$PREFIX/tmp/pkg/mod"
export GOPATH="$PREFIX/tmp/gopath"
export GOTOOLCHAIN=local
export GOTMPDIR="$PREFIX/tmp"
export GOSUMDB=off
export GONOSUMCHECK=*
export SSL_CERT_FILE="$PREFIX/etc/tls/cert.pem"
export CGO_ENABLED=0
export GOWORK=off

# Detect the current Termux user
export USER=$(whoami)
export PG_SOCKET_DIR="$PREFIX/tmp"

# Project paths
export FLYFLEET_ROOT="$HOME/flyfleet"
export FLEET_CORE_DIR="$FLYFLEET_ROOT/services/fleet-core"
export DASHBOARD_DIR="$FLYFLEET_ROOT/apps/dashboard-web"
export DATABASE_URL="host=$PG_SOCKET_DIR user=$USER dbname=fleet_db sslmode=disable"

echo ""
echo "🚀 Starting FlyFleet on Android/Termux..."
echo "📂 Project root: $FLYFLEET_ROOT"
echo "👤 User: $USER"

# Check if flyfleet exists
if [ ! -d "$FLYFLEET_ROOT" ]; then
    echo "❌ Error: FlyFleet not found at $FLYFLEET_ROOT"
    echo "   Clone it first:"
    echo "   git clone https://github.com/abhishek-8285/flyfleet.git $HOME/flyfleet"
    exit 1
fi

# Ensure directories exist
mkdir -p "$GOCACHE" "$GOMODCACHE" "$GOPATH" "$GOTMPDIR"

# Step 1: Start PostgreSQL
echo ""
echo "🗄️  Starting PostgreSQL..."
if ! pgrep postgres > /dev/null 2>&1; then
    PG_DATA_DIR="$PREFIX/var/lib/postgresql"
    if [ ! -d "$PG_DATA_DIR" ] || [ ! -f "$PG_DATA_DIR/PG_VERSION" ]; then
        echo "   Initializing PostgreSQL database..."
        mkdir -p "$PG_DATA_DIR"
        initdb -D "$PG_DATA_DIR" >/dev/null 2>&1
    fi
    pg_ctl -D "$PG_DATA_DIR" -o "-c listen_addresses= -c unix_socket_directories=$PG_SOCKET_DIR" -l "$PREFIX/var/log/postgres.log" start
    sleep 3
fi
echo "   ✓ PostgreSQL running"

# Step 2: Prepare database
echo ""
echo "🗄️  Preparing database..."
if ! psql -h "$PG_SOCKET_DIR" -d fleet_db -c "SELECT 1" >/dev/null 2>&1; then
    createdb -h "$PG_SOCKET_DIR" fleet_db 2>/dev/null || echo "   (database may already exist)"
fi
echo "   ✓ Database fleet_db exists"

# Step 3: Start Redis (with ARM64 kernel workaround)
echo ""
echo "🔴 Starting Redis..."
if ! pgrep redis-server > /dev/null 2>&1; then
    redis-server --port 6380 --daemonize yes --ignore-warnings ARM64-COW-BUG 2>&1 | tail -2
    sleep 1
fi
if pgrep redis-server > /dev/null 2>&1; then
    echo "   ✓ Redis running on port 6380"
else
    echo "   ⚠️  Redis failed to start, trying default port..."
    redis-server --daemonize yes --ignore-warnings ARM64-COW-BUG 2>&1 | tail -2
    sleep 1
fi

# Step 4: Start NATS (if available)
echo ""
echo "🟣 Starting NATS..."
if command -v nats-server > /dev/null 2>&1; then
    if ! pgrep nats-server > /dev/null 2>&1; then
        nats-server -js -p 4222 > /dev/null 2>&1 &
        sleep 2
    fi
    echo "   ✓ NATS running on port 4222"
else
    echo "   ⚠️  NATS not installed. Install with: pkg install -y nats-server"
    echo "   (Messaging features may not work without NATS)"
fi

# Step 5: Build fleet-core backend
echo ""
echo "🏗️  Building backend..."
cd "$FLEET_CORE_DIR"
go build -o fleet-core ./cmd 2>&1 | tail -5
echo "   ✓ Backend built"

# Step 6: Run database migrations directly (skips schema harmonizer)
echo ""
echo "📦 Running database migrations..."
go run ./cmd/migrate up 2>&1 | tail -10
echo "   ✓ Migrations done"

# Step 7: Start backend (harmonizer runs at startup, no-op on existing schema)
echo ""
echo "📦 Starting Fleet Core backend on port 8110..."
mkdir -p tmp/logs
if pgrep fleet-core > /dev/null 2>&1; then
    pkill -f fleet-core 2>/dev/null || true
    sleep 2
fi
nohup ./fleet-core > tmp/logs/backend_persist.log 2>&1 &
echo $! > tmp/backend.pid
sleep 5

# Check if backend started
if kill -0 $(cat tmp/backend.pid) 2>/dev/null; then
    echo "   ✓ Backend running (PID: $(cat tmp/backend.pid))"
    tail -5 tmp/logs/backend_persist.log
else
    echo "   ❌ Backend failed to start. Check logs:"
    echo "      tail -50 tmp/logs/backend_persist.log"
    exit 1
fi

# Step 8: Build and start UI
echo ""
echo "🎨 Starting UI on port 3005..."
cd "$DASHBOARD_DIR"

# Install npm dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "   Installing npm dependencies..."
    npm install 2>&1 | tail -5
fi

# Start UI dev server
if pgrep -f vite > /dev/null 2>&1; then
    pkill -f vite 2>/dev/null || true
    sleep 1
fi
mkdir -p tmp
nohup npx vite --host 0.0.0.0 --port 3005 > tmp/ui_persist.log 2>&1 &
echo $! > tmp/ui.pid
sleep 2

echo ""
echo "------------------------------------------------"
echo "✅ FlyFleet is now running!"
echo "👉 Dashboard:          http://localhost:3005/billing/designer"
echo "👉 Backend API:        http://localhost:8110"
echo "👉 Health Check:       http://localhost:8110/health"
echo ""
echo "📁 Logs:"
echo "   Backend:  $FLEET_CORE_DIR/tmp/logs/backend_persist.log"
echo "   UI:       $DASHBOARD_DIR/tmp/ui_persist.log"
echo ""
echo "🛠️  To stop all services:"
echo "   pkill -f fleet-core"
echo "   pkill -f vite"
echo "   pkill -f redis-server"
echo "   pkill -f postgres"
echo "   pkill -f nats-server"
echo ""
echo "📋 Keep this Termux window open while running."
echo "------------------------------------------------"
