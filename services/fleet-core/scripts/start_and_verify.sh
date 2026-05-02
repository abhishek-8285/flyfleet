#!/bin/bash
cd /home/abhishek/Desktop/flyfleet/services/fleet-core
./fleet-core > startup_final.log 2>&1 &
PID=$!
echo "Started fleet-core with PID $PID"
sleep 15
if ps -p $PID > /dev/null; then
    echo "Service is still running"
else
    echo "Service failed to start"
    cat startup_final.log
fi
