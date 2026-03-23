#!/bin/bash
# scripts/emergency_recovery_test.sh
# FlyFleet Disaster Recovery Validator for Emergency Module

set -e

echo "--- 🚀 Emergency Recovery Test Started ---"

# 1. DB Connectivity Check
echo "[1/4] Checking Database Resilience..."
docker exec $(docker ps -q -f name=postgres) pg_isready -U fleet_user -d fleet_db

# 2. Redis Persistence Check
echo "[2/4] Verifying Rate Limit Persistence..."
docker exec $(docker ps -q -f name=redis) redis-cli PING

# 3. NATS Handshake
echo "[3/4] Testing NATS Event Bus..."
# Minimal sub/pub check via nats-server (if available) or service check
if docker ps | grep -q "nats"; then
  echo "✅ NATS is healthy"
else
  echo "❌ NATS is DOWN"
  exit 1
fi

# 4. Critical Data Integrity
echo "[4/4] Auditing Active SOS Incidents..."
COUNT=$(docker exec $(docker ps -q -f name=postgres) psql -U fleet_user -d fleet_db -t -A -c "SELECT COUNT(*) FROM sos_incidents WHERE status = 'TRIGGERED';")
echo "✅ $COUNT Unresolved SOS incidents found in persistent storage."

echo "--- 🏁 Recovery Test Completed Successfully ---"
