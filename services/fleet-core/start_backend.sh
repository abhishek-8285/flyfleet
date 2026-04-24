#!/bin/bash
cd /home/abhishek/Desktop/flyfleet/services/fleet-core
go run ./cmd/main.go > fleet-startup.log 2>&1 &
echo $! > backend.pid
