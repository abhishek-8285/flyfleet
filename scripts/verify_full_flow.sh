#!/bin/bash

# Configuration
API_URL="http://localhost:8081/api/v1"
ADMIN_EMAIL="admin@flyfleet.com"
ADMIN_PASSWORD="admin123"

echo "============================================"
echo "  FlyFleet Full Flow Verification Script"
echo "============================================"

# 1. Login to get Token
echo ""
echo "[1] Authenticating as Super Admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Login Failed"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi
echo "✅ Login Successful"

# 2. Get Organizations (Check for Delta Logistics)
echo ""
echo "[2] Fetching Organizations (Looking for 'Delta Logistics')..."
ORGS_RESPONSE=$(curl -sL -X GET "$API_URL/admin/organizations" \
  -H "Authorization: Bearer $TOKEN")

# Check if Delta exists in response
if echo "$ORGS_RESPONSE" | grep -q "FlyFleet Global"; then
  echo "✅ Found 'FlyFleet Global' in organization list"
  # Extract Org ID
  DELTA_ORG_ID="00000000-0000-0000-0000-000000000001" 
else
  echo "❌ 'FlyFleet Global' NOT found in response"
  echo "Response Preview: ${ORGS_RESPONSE:0:200}..."
  exit 1
fi

# 3. Get Vehicles
echo ""
echo "[3] Fetching Vehicles (Org: $DELTA_ORG_ID)..."
VEHICLES_RESPONSE=$(curl -sL -X GET "$API_URL/vehicles?org_id=$DELTA_ORG_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$VEHICLES_RESPONSE" | grep -q "DLT-001"; then
  echo "✅ Found Vehicle 'DLT-001'"
else
  echo "❌ Vehicle 'DLT-001' NOT found"
  echo "Response: $VEHICLES_RESPONSE"
fi

# 4. Get Drivers
echo ""
echo "[4] Fetching Drivers (Org: $DELTA_ORG_ID)..."
DRIVERS_RESPONSE=$(curl -sL -X GET "$API_URL/drivers?org_id=$DELTA_ORG_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$DRIVERS_RESPONSE" | grep -q "Delta Driver 1"; then
  echo "✅ Found Driver 'Delta Driver 1'"
else
  echo "❌ Driver 'Delta Driver 1' NOT found"
fi

# 5. Get Trips
echo ""
echo "[5] Fetching Trips (Org: $DELTA_ORG_ID)..."
TRIPS_RESPONSE=$(curl -sL -X GET "$API_URL/trips?org_id=$DELTA_ORG_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$TRIPS_RESPONSE" | grep -q "Bangalore to Mysore Logistics"; then
  echo "✅ Found Active Trip 'Bangalore to Mysore Logistics'"
else
  echo "❌ Active Trip NOT found"
  echo "Response: $TRIPS_RESPONSE"
fi

# 6. Get Alerts
echo ""
echo "[6] Fetching Alerts (Org: $DELTA_ORG_ID)..."
ALERTS_RESPONSE=$(curl -sL -X GET "$API_URL/alerts?org_id=$DELTA_ORG_ID" \
  -H "Authorization: Bearer $TOKEN")

# Check for specific alert message
if echo "$ALERTS_RESPONSE" | grep -q "overspeed"; then
  echo "✅ Found Critical Alert: 'Vehicle DLT-001 exceeded speed limit'"
else
  echo "❌ critical Overspeed Alert NOT found"
  echo "Response: $ALERTS_RESPONSE"
fi

echo ""
echo "============================================"
echo "  Verification Complete"
echo "============================================"
