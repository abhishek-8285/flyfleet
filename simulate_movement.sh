#!/bin/bash

# Configuration
VEHICLE_ID="8c476a51-77a7-4b9f-853e-ce1a1df3939a"
DEVICE_ID="1234567890"
TENANT_ID="550e8400-e29b-41d4-a716-446655440002"
ENDPOINT="http://localhost:8084/api/v1/telemetry"

# Starting from your real hardware coordinates (Delhi)
LAT=28.6490624
LNG=77.217792

echo "Starting Telemetry Simulation on PORT 8084 from Delhi ($LAT, $LNG)..."
echo "Press [CTRL+C] to stop."

while true; do
  # Generate a unique event ID (force lowercase for HMAC)
  EVENT_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  
  # Simulate movement (add a small random variation)
  LAT=$(echo "$LAT + 0.0005" | bc -l)
  LNG=$(echo "$LNG + 0.0003" | bc -l)
  
  # 1. Generate Signature (HMAC-SHA256 of EventID + DeviceID)
  SECRET="FF-SECRET-2026"
  MESSAGE="${EVENT_ID}${DEVICE_ID}"
  SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')

  # 2. Construct Payload
  PAYLOAD=$(cat <<EOF
{
  "event_id": "$EVENT_ID",
  "device_id": "$DEVICE_ID",
  "vehicle_id": "$VEHICLE_ID",
  "tenant_id": "$TENANT_ID",
  "lat": $LAT,
  "lng": $LNG,
  "speed": 45.0,
  "heading": 90,
  "ignition": true,
  "recorded_at": "$TIMESTAMP"
}
EOF
)

  # 3. Send Request with Signature Header
  curl -s -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -H "X-HMAC-Signature: $SIGNATURE" \
    -d "$PAYLOAD"

  echo "[$TIMESTAMP] Sent: Lat=$LAT, Lng=$LNG (Event: ${EVENT_ID:0:8}...)"
  
  sleep 5
done
