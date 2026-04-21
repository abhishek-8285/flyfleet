#!/bin/bash
# Force kill existing processes
pkill -f air || true
pkill -f "go-build" || true
pkill -f "cmd/main.go" || true
pkill -f "fleet-core" || true

# Wait for ports to clear
sleep 2

# Start the service in the background
cd /home/abhishek/Desktop/flyfleet/services/fleet-core
go build -o fleet-core ./cmd
nohup ./fleet-core > fleet-manual.log 2>&1 &

echo "Restart command initiated. Check fleet-manual.log for progress."
