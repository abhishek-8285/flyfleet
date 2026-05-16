#!/bin/bash

# Configuration
VEHICLE_IDS=("550e8400-e29b-41d4-a716-446655440001" "550e8400-e29b-41d4-a716-446655440002" "550e8400-e29b-41d4-a716-446655440003" "550e8400-e29b-41d4-a716-446655440004" "550e8400-e29b-41d4-a716-446655440005")
DEVICE_IDS=("DEV-DEVICE-001" "DEV-DEVICE-002" "DEV-DEVICE-003" "DEV-DEVICE-004" "DEV-DEVICE-005")
TENANT_ID="550e8400-e29b-41d4-a716-446655440000"
API_URL="http://localhost:8110/api/v1/telemetry"

echo "🚀 Launching Swarm Simulation (5 Vehicles)..."

for i in {0..4}; do
  VID=${VEHICLE_IDS[$i]}
  DID=${DEVICE_IDS[$i]}
  
  # Run a background loop for each vehicle
  (
    LAT=28.6$(($i+1))39
    LNG=77.2$(($i+1))90
    HEADING=$((135 + $i*10))
    
    while true; do
      EVENT_ID=$(cat /proc/sys/kernel/random/uuid)
      TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
      
      MESSAGE="${EVENT_ID}${DID}"
      SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "telemetry-secret" | sed 's/^.* //')
      
      LAT=$(echo "$LAT + 0.0003" | bc -l)
      LNG=$(echo "$LNG + 0.0004" | bc -l)
      
      PAYLOAD=$(cat <<EOF
{
  "event_id": "$EVENT_ID",
  "device_id": "$DID",
  "vehicle_id": "$VID",
  "tenant_id": "$TENANT_ID",
  "lat": $LAT,
  "lng": $LNG,
  "speed": 45.5,
  "heading": $HEADING,
  "ignition": true,
  "event_type": "GPS",
  "recorded_at": "$TIMESTAMP"
}
EOF
)
      
      curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -H "X-HMAC-Signature: $SIGNATURE" \
        -d "$PAYLOAD" > /dev/null
        
      sleep 5
    done
  ) &
done

echo "Swarm is running. Check air.log for ingestion logs."
echo "Press Ctrl+C to stop (may require pkill swarm.sh)."
wait
