#!/bin/bash
# zombie_test.sh
# Tests Zombie Killer Ping Pre-Flight Check

API_URL="http://localhost:8084/api/v1/device/telemetry/gps"
AUTH_URL="http://localhost:8084/api/v1/auth/login"
ORGA_ID="c57c4ec8-5190-4776-9267-06c3da35bc3e"
EMAIL="super_admin@flyfleet.com"
PASSWORD="fleet_password"
DEVICE_ID="TEST-DEV-001"
TIMESTAMP=$(date +%s)

echo "Attempting to acquire JWT for Zombie Test..."
LOGIN_PAYLOAD="{\"email\": \"${EMAIL}\", \"password\": \"${PASSWORD}\", \"org_id\": \"${ORGA_ID}\"}"
LOGIN_RESPONSE=$(curl -s -X POST $AUTH_URL -H "Content-Type: application/json" -d "${LOGIN_PAYLOAD}")

JWT=$(echo $LOGIN_RESPONSE | grep -oP '"access_token":"\+.*?"' | cut -d '"' -f 4)
REFRESH=$(echo $LOGIN_RESPONSE | grep -oP '"refresh_token":"\+.*?"' | cut -d '"' -f 4)
# In our architecture RevokeToken revokes refresh tokens directly, but RevokeSession targets device sessions.
# Let's hit the manual session revoke or emergency global revoke via the refresh token.

if [ -z "$JWT" ]; then
    echo "Login failed. Could not acquire JWT."
    exit 1
fi
echo "Acquired JWT successfully."

echo -e "\n1. Sending Valid GPS Ping..."
PING_PAYLOAD="{\"device_id\": \"${DEVICE_ID}\", \"timestamp\": ${TIMESTAMP}, \"latitude\": 40.7128, \"longitude\": -74.0060}"
curl -s -o /dev/null -w "%{http_code}\n" -X POST $API_URL \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d "${PING_PAYLOAD}"

echo -e "\n2. Revoking Session/Token..."
# Attempting global token revocation or refresh token revocation for speed
REVOKE_PAYLOAD="{\"refresh_token\": \"${REFRESH}\"}"
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8084/api/v1/auth/revoke \
  -H "Content-Type: application/json" \
  -d "${REVOKE_PAYLOAD}"

echo -e "\n3. Sending Zombie GPS Ping with Revoked Token..."
# Bump timestamp to avoid dedupe
NEW_TIMESTAMP=$(($TIMESTAMP + 10))
NEW_PING_PAYLOAD="{\"device_id\": \"${DEVICE_ID}\", \"timestamp\": ${NEW_TIMESTAMP}, \"latitude\": 40.7128, \"longitude\": -74.0060}"

echo "Expected Result: 401 Unauthorized"
curl -s -i -X POST $API_URL \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d "${NEW_PING_PAYLOAD}"
