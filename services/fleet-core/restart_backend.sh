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
echo "Building fleet-core..." > build.log
if go build -v -o fleet-core ./cmd >> build.log 2>&1; then
    echo "Build successful. Starting service..." >> build.log
    nohup ./fleet-core > fleet-manual.log 2>&1 &
    echo "Restart command initiated. Check fleet-manual.log for progress."
else
    echo "Build FAILED. Check build.log for details."
fi
