#!/bin/bash

# ============================================================================
# FLYFLEET VEHICLE SIMULATOR (DUAL-MODE)
# ============================================================================

# Color definitions
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Configuration
VEHICLE_REG="DL01AB1234"
VEHICLE_ID="550e8400-e29b-41d4-a716-446655440001"
TENANT_ID="550e8400-e29b-41d4-a716-446655440000"
DEVICE_ID="DEV-DEVICE-001"
API_URL="http://localhost:8110/api/v1/telemetry"
INTERVAL="5s"

echo -e "${BOLD}${CYAN}================================================================${NC}"
echo -e "${BOLD}${CYAN}          FLYFLEET VEHICLE SIMULATOR - DEPLOYMENT CONTROL        ${NC}"
echo -e "${BOLD}${CYAN}================================================================${NC}"
echo -e "Press [1] for Protobuf Ingress Mode (Recommended - Edge Telemetry)"
echo -e "Press [2] for Legacy JSON Ingress Mode"
echo -e "----------------------------------------------------------------"
read -p "Select Mode (1 or 2): " mode_choice

if [ "$mode_choice" = "2" ]; then
  MODE="json"
else
  MODE="protobuf"
fi

# Detect Go and compile the premium Go simulator
if command -v go &> /dev/null; then
  echo -e "\n${GREEN}✔ Go Compiler Detected.${NC} Launching premium ${BOLD}${MODE}${NC} simulator..."
  echo -e "${GRAY}Compiling services/fleet-core/cmd/simulator/main.go...${NC}"
  
  # Run the premium Go simulator
  go run services/fleet-core/cmd/simulator/main.go \
    -mode "$MODE" \
    -url "$API_URL" \
    -vehicle-id "$VEHICLE_ID" \
    -tenant-id "$TENANT_ID" \
    -device-id "$DEVICE_ID" \
    -reg "$VEHICLE_REG" \
    -interval "$INTERVAL"
    
  exit 0
fi

# ============================================================================
# FALLBACK: Bash-based Legacy JSON Loop (If Go is not installed)
# ============================================================================
echo -e "\n${YELLOW}⚠ Go Compiler Not Detected. Using Legacy Bash fallback (JSON only)...${NC}"

LAT=28.6139
LNG=77.2090
PREV_LAT=$LAT
PREV_LNG=$LNG
HEADING=135
SLEEP_INTERVAL=5

while true; do
  EVENT_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || uuidgen 2>/dev/null || echo "550e8400-e29b-41d4-a716-446655440099")
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  
  # Calculate HMAC Signature
  MESSAGE="${EVENT_ID}${DEVICE_ID}"
  SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "telemetry-secret" | sed 's/^.* //')

  # Simulate movement
  LAT=$(echo "$LAT + 0.0003" | bc -l)
  LNG=$(echo "$LNG + 0.0004" | bc -l)
  
  DLAT=$(echo "$LAT - $PREV_LAT" | bc -l)
  DLNG=$(echo "$LNG - $PREV_LNG" | bc -l)
  DIST=$(echo "scale=5; sqrt(($DLAT*111139)^2 + ($DLNG*97500)^2)" | bc -l)
  SPEED=$(echo "scale=2; ($DIST / $SLEEP_INTERVAL) * 2.237" | bc -l)
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

  echo -e "[${GRAY}$TIMESTAMP${NC}] Pos: $LAT, $LNG | Speed: $SPEED MPH | Hdg: $HEADING°"
  
  curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "X-HMAC-Signature: $SIGNATURE" \
    -d "$PAYLOAD" | grep -q "accepted" && echo -e "  - ${GREEN}Accepted (JSON)${NC}" || echo -e "  - ${RED}Failed${NC}"

  PREV_LAT=$LAT
  PREV_LNG=$LNG
  sleep $SLEEP_INTERVAL
done
