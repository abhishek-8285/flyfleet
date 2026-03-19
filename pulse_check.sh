#!/bin/bash
# pulse_check.sh
# Tests backend Tenant Isolation by faking a different X-Org-ID

API_URL="http://localhost:8084/api/v1/vehicles" # generic protected endpoint
ORGA_ID="c57c4ec8-5190-4776-9267-06c3da35bc3e"
ORGB_ID="d68d5ed9-62a1-5887-a378-17d4eb46cd4f"
EMAIL="super_admin@flyfleet.com"
PASSWORD="fleet_password"

echo "Attempting to acquire JWT for Org A..."
LOGIN_PAYLOAD="{\"email\": \"${EMAIL}\", \"password\": \"${PASSWORD}\", \"org_id\": \"${ORGA_ID}\"}"
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8084/api/v1/auth/login -H "Content-Type: application/json" -d "${LOGIN_PAYLOAD}")

JWT=$(echo $LOGIN_RESPONSE | grep -oP '"access_token":"\+.*?"' | cut -d '"' -f 4)

if [ -z "$JWT" ]; then
    echo "Login failed. Could not acquire JWT."
    exit 1
fi
echo "Acquired JWT successfully."

echo -e "\nMaking Request to $API_URL with Org A JWT, but injecting Org B X-Org-ID Header..."
echo "Expected Result: 403 Forbidden"
curl -s -i -X GET $API_URL \
  -H "Authorization: Bearer $JWT" \
  -H "X-Org-ID: $ORGB_ID"
