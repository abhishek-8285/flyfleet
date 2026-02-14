#!/bin/bash

# FlyFleet Integrated Startup Script
# This script handles port cleanup, service builds, migrations, and frontend startup.

# 1. Pull latest changes (Optional, uncomment if using Git)
# git pull

# 2. Ensure scripts are executable
chmod +x migrate.sh

# Define service ports to clean up (Backend + Frontend + Infrastructure)
PORTS="8080 8081 8082 8083 8084 8085 8086 8087 8088 8089 8090 8091 8092 3000 3001 4222 8222 6380"

echo "Stopping any existing services on ports $PORTS..."

for port in $PORTS; do
  pid=$(lsof -t -i:$port)
  if [ -n "$pid" ]; then
    echo "Killing process on port $port (PID: $pid)..."
    kill -9 $pid
  fi
done

echo "Stopping Docker containers..."
docker-compose down

echo "Building and starting Backend services..."
docker-compose up --build -d

echo "Waiting for database to be ready..."
sleep 10

echo "Running database migrations..."
chmod +x migrate.sh
./migrate.sh

echo "Starting Frontend services..."

# Start Dashboard Web (Port 3000)
echo "Starting Dashboard Web on port 3000..."
cd apps/dashboard-web
# Install dependencies to ensure latest changes are picked up
npm install
# Start in background, verify 
PORT=3000 npm start > ../../dashboard.log 2>&1 &
cd ../..

# Start Landing Page (Port 3001)
echo "Starting Landing Page on port 3001..."
cd apps/landing-page
# Install dependencies
npm install
# Start in background
PORT=3001 npm run dev > ../../landing.log 2>&1 &
cd ../..

echo "---------------------------------------------------"
echo "All Services Started Successfully!"
echo "---------------------------------------------------"
echo "Backend API Gateway: http://localhost:8080"
echo "Dashboard Web App:   http://localhost:3000"
echo "Landing Page:        http://localhost:3001"
echo "---------------------------------------------------"
echo "Frontend logs are being written to dashboard.log and landing.log"
echo "Use 'tail -f dashboard.log' to monitor Dashboard startup."
echo "---------------------------------------------------"

docker-compose ps
