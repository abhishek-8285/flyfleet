#!/bin/bash
# defense_test.sh - Verify Global Kill-Switch

API_URL="http://localhost:8084"
STATUS_KEY="flyfleet:system:status"
REDIS_PORT="6379" # Standard local redis for tests

echo "[SRE Defense] Activating MAINTENANCE mode via Redis..."
redis-cli -p $REDIS_PORT SET $STATUS_KEY "MAINTENANCE" > /dev/null

# Wait for the Go local memory cache (1 sec) to expire
echo "Waiting 1.5 seconds for SystemGuard local cache expiry..."
sleep 1.5

echo "[SRE Defense] Testing GET request to /health..."
GET_BODY=$(curl -s -w "\nHTTP_CODE:%{http_code}\n" -X GET "$API_URL/health")
GET_STATUS=$(echo "$GET_BODY" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$GET_STATUS" = "503" ]; then
    echo "❌ FAILURE: GET request was blocked! (Status: $GET_STATUS)"
    GET_PASS=0
else
    echo "✅ SUCCESS: GET request was allowed during MAINTENANCE. (Status: $GET_STATUS)"
    GET_PASS=1
fi

echo "[SRE Defense] Testing POST request to /api/v1/auth/login (Should return 503)..."
POST_BODY=$(curl -s -w "\nHTTP_CODE:%{http_code}\n" -X POST "$API_URL/api/v1/auth/login" -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"test"}')
POST_STATUS=$(echo "$POST_BODY" | grep "HTTP_CODE" | cut -d':' -f2)
POST_DATA=$(echo "$POST_BODY" | sed '$d')

if [ "$POST_STATUS" = "503" ]; then
    echo "✅ SUCCESS: POST mutation was blocked perfectly! (Status: 503)"
    echo "Response: $POST_DATA"
    POST_PASS=1
else
    echo "❌ FAILURE: POST mutation bypassed the Kill-Switch! (Status: $POST_STATUS)"
    echo "Response: $POST_DATA"
    POST_PASS=0
fi

echo "[SRE Defense] Cleaning up Redis..."
redis-cli -p $REDIS_PORT DEL $STATUS_KEY > /dev/null

if [ "$GET_PASS" -eq 1 ] && [ "$POST_PASS" -eq 1 ]; then
    echo "🎉 [Result] Automated SRE Defense functionality is VERIFIED."
    exit 0
else
    echo "💀 [Result] Automated SRE Defense functionality FAILED."
    exit 1
fi
