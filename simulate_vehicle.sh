#!/bin/bash

# Configuration
VEHICLE_REG="DL01AB1234"
VEHICLE_ID="550e8400-e29b-41d4-a716-446655440001"
TENANT_ID="550e8400-e29b-41d4-a716-446655440000"
DEVICE_ID="DEV-DEVICE-001"
API_URL="http://localhost:8110/api/v1/telemetry"

# Starting state
LAT=28.6139
LNG=77.2090
PREV_LAT=$LAT
PREV_LNG=$LNG
HEADING=135
SLEEP_INTERVAL=5

echo "Starting telemetry simulation for $VEHICLE_REG ($VEHICLE_ID)..."
echo "Speed is now calculated based on coordinate changes."

while true; do
  EVENT_ID=$(cat /proc/sys/kernel/random/uuid)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  
  # Calculate HMAC Signature
  MESSAGE="${EVENT_ID}${DEVICE_ID}"
  SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "telemetry-secret" | sed 's/^.* //')

  # Simulate movement (approx 45-50 MPH at 5s intervals)
  LAT=$(echo "$LAT + 0.0003" | bc -l)
  LNG=$(echo "$LNG + 0.0004" | bc -l)
  
  # Calculate distance moved in meters (simplified Euclidean for small changes)
  DLAT=$(echo "$LAT - $PREV_LAT" | bc -l)
  DLNG=$(echo "$LNG - $PREV_LNG" | bc -l)
  
  # 1 deg lat ~ 111,139m, 1 deg lng at 28.6 deg lat ~ 97,500m
  DIST=$(echo "scale=5; sqrt(($DLAT*111139)^2 + ($DLNG*97500)^2)" | bc -l)
  
  # Speed in MPH: (meters / seconds) * 2.237
  SPEED=$(echo "scale=2; ($DIST / $SLEEP_INTERVAL) * 2.237" | bc -l)
  
  # Update Heading (rotate slightly)
  HEADING=$(( (HEADING + 2) % 360 ))

  PAYLOAD=$(cat <<EOF
{
  "event_id": "$EVENT_ID",
  "device_id": "$DEVICE_ID",
  "vehicle_id": "$VEHICLE_ID",
  "tenant_id": "$TENANT_ID",
  "lat": $LAT,
  "lng": $LNG,
  "speed": $SPEED,
  "heading": $HEADING,
  "ignition": true,
  "event_type": "GPS",
  "recorded_at": "$TIMESTAMP"
}
EOF
)

  echo "[$TIMESTAMP] Pos: $LAT, $LNG | Speed: $SPEED MPH | Hdg: $HEADING°"
  
  curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "X-HMAC-Signature: $SIGNATURE" \
    -d "$PAYLOAD" | grep -q "accepted" && echo "  - Accepted" || echo "  - Failed"

  PREV_LAT=$LAT
  PREV_LNG=$LNG
  sleep $SLEEP_INTERVAL
done


