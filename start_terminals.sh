#!/bin/bash

echo "🧹 Cleaning up any existing backend/frontend ports..."
PORTS="8084 3000 3001 4222 8222 6379 5432"
for port in $PORTS; do
  pid=$(lsof -t -i:$port 2>/dev/null)
  if [ -n "$pid" ]; then
    kill -9 $pid 2>/dev/null || true
  fi
done

echo "💀 Force-killing any lingering Node/Go/NATS background instances..."
pkill -9 -f "next" 2>/dev/null || true
pkill -9 -f "react-scripts" 2>/dev/null || true
pkill -9 -f "go run cmd/main.go" 2>/dev/null || true
pkill -9 -f "nats-server" 2>/dev/null || true
sudo killall -9 nats-server 2>/dev/null || true

echo "🐳 Turning off any conflicting Docker containers from old runs..."
docker-compose down 2>/dev/null || true

echo "⚙️  Generating separate launch scripts..."
mkdir -p .flyfleet_startup

cat > .flyfleet_startup/pg.sh << 'EOF'
#!/bin/bash
echo -e "\033]0;PostgreSQL\007"
echo "Connecting to PostgreSQL Native Service..."
echo "Tailing pg_log (if accessible) or launching psql..."
sleep 2
PGPASSWORD=fleet_password psql -U fleet_user -d fleet_db -h localhost
exec bash
EOF

cat > .flyfleet_startup/redis.sh << 'EOF'
#!/bin/bash
echo -e "\033]0;Redis Cache\007"
echo "Connecting to Redis Server..."
redis-cli ping
exec bash
EOF

cat > .flyfleet_startup/nats.sh << 'EOF'
#!/bin/bash
echo -e "\033]0;NATS Jetstream\007"
echo "Starting NATS Broker natively..."
cd nats-server-v2.10.11-linux-amd64 && ./nats-server -js -m 8222
exec bash
EOF

cat > .flyfleet_startup/go.sh << 'EOF'
#!/bin/bash
echo -e "\033]0;Go Backend\007"
echo "Waiting for DB/NATS/Redis to initialize... (5s)"
sleep 5
echo "Starting Go Backend natively..."
cd services/fleet-core-service-legacy
export PORT=8084
export DATABASE_URL="postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
export NATS_URL="nats://localhost:4222"
export REDIS_ADDR="localhost:6379"
export REDIS_URL="localhost:6379"
go run cmd/main.go
exec bash
EOF

cat > .flyfleet_startup/dash.sh << 'EOF'
#!/bin/bash
echo -e "\033]0;Dashboard Web\007"
echo "Waiting for Backend to initialize... (8s)"
sleep 8
cd apps/dashboard-web
npm install >/dev/null 2>&1
export REACT_APP_API_URL="http://localhost:8084/api/v1"
PORT=3000 BROWSER=none npm start
exec bash
EOF

cat > .flyfleet_startup/land.sh << 'EOF'
#!/bin/bash
echo -e "\033]0;Landing Page\007"
echo "Waiting for Backend to initialize... (8s)"
sleep 8
cd apps/landing-page
npm install >/dev/null 2>&1
PORT=3001 npm run dev
exec bash
EOF

cat > .flyfleet_startup/utils.sh << 'EOF'
#!/bin/bash
echo -e "\033]0;FlyFleet Utils\007"
echo "Waiting for database to initialize... (4s)"
sleep 4
export DB_HOST=localhost
export DB_PORT=5432
export DB_PASSWORD=fleet_password
chmod +x migrate.sh
./migrate.sh
echo -e "\n\n🐘 Migrations Done!\n"
echo "👉 Run psql queries here or interact with the system!"
exec bash
EOF

chmod +x .flyfleet_startup/*.sh

echo "🚀 Launching native services in a single terminal window with 7 tabs..."

if command -v gnome-terminal &> /dev/null; then
  gnome-terminal --window --title="FlyFleet Services" \
    --tab -e "$PWD/.flyfleet_startup/pg.sh" \
    --tab -e "$PWD/.flyfleet_startup/redis.sh" \
    --tab -e "$PWD/.flyfleet_startup/nats.sh" \
    --tab -e "$PWD/.flyfleet_startup/go.sh" \
    --tab -e "$PWD/.flyfleet_startup/dash.sh" \
    --tab -e "$PWD/.flyfleet_startup/land.sh" \
    --tab -e "$PWD/.flyfleet_startup/utils.sh" 2>/dev/null
elif command -v konsole &> /dev/null; then
  konsole --new-tab -p tabtitle="PostgreSQL" -e "$PWD/.flyfleet_startup/pg.sh" &
  konsole --new-tab -p tabtitle="Redis" -e "$PWD/.flyfleet_startup/redis.sh" &
  konsole --new-tab -p tabtitle="NATS" -e "$PWD/.flyfleet_startup/nats.sh" &
  konsole --new-tab -p tabtitle="Go Backend" -e "$PWD/.flyfleet_startup/go.sh" &
  konsole --new-tab -p tabtitle="Dashboard Web" -e "$PWD/.flyfleet_startup/dash.sh" &
  konsole --new-tab -p tabtitle="Landing Page" -e "$PWD/.flyfleet_startup/land.sh" &
  konsole --new-tab -p tabtitle="FlyFleet Utils" -e "$PWD/.flyfleet_startup/utils.sh" &
else
  echo "Error: Neither gnome-terminal nor konsole found to support native tabs."
fi

echo "==================================================="
echo "✅ 7 GUI Terminal Tabs launched in a single window!"
echo "==================================================="
