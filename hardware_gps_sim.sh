#!/bin/bash

# Configuration
VEHICLE_ID="01886585-6189-47b2-a97e-843cba9b8ddf"
DEVICE_ID="1234567890"
TENANT_ID="550e8400-e29b-41d4-a716-446655440002"
ENDPOINT="http://localhost:8084/api/v1/telemetry"
SECRET="FF-SECRET-2026"

echo "Accessing Laptop Hardware GPS module..."

# Check if gpsd is available
if ! command -v gpspipe &> /dev/null; then
    echo "Error: 'gpsd-clients' not installed. Please run: sudo apt install gpsd-clients"
    exit 1
fi

while true; do
  # Extract Latitude and Longitude from gpsd JSON output
  GPS_DATA=$(gpspipe -w -n 10 | grep -m 1 "TPV")
  LAT=$(echo "$GPS_DATA" | jq -r '.lat // empty')
  LNG=$(echo "$GPS_DATA" | jq -r '.lon // empty')
  SPEED=$(echo "$GPS_DATA" | jq -r '.speed // 0')
  
  if [ -z "$LAT" ] || [ "$LAT" == "null" ]; then
    echo "Waiting for GPS Fix (Satellite Lock)..."
    sleep 2
    continue
  fi

  EVENT_ID=$(uuidgen)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  
  # Generate HMAC Signature
  MESSAGE="${EVENT_ID}${DEVICE_ID}"
  SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')

  # Construct Payload
  PAYLOAD=$(cat <<EOF
{
  "event_id": "$EVENT_ID",
  "device_id": "$DEVICE_ID",
  "vehicle_id": "$VEHICLE_ID",
  "tenant_id": "$TENANT_ID",
  "lat": $LAT,
  "lng": $LNG,
  "speed": $SPEED,
  "heading": 0,
  "ignition": true,
  "event_type": "GPS",
  "recorded_at": "$TIMESTAMP"
}
EOF
)

  # Send to Backend
  curl -s -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -H "X-HMAC-Signature: $SIGNATURE" \
    -d "$PAYLOAD" > /dev/null

  echo "[$TIMESTAMP] Hardware GPS Update: Lat=$LAT, Lng=$LNG, Speed=$SPEED"
  
  sleep 5
done
