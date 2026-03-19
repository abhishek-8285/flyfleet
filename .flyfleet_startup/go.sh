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
