#!/bin/bash
# test_billing_idempotency.sh

ORG_ID="org-finshield-test"
AUTH_TOKEN="Bearer mock-token-for-testing"
API_URL="http://localhost:8080/api/v1/india/fastag/transactions"

echo "[Financial Shield] Generating X-Request-ID..."
REQUEST_ID=$(uuidgen)
echo "Using ID: $REQUEST_ID"

PAYLOAD='{"vehicle_id": "truck-001", "toll_plaza": "Test Plaza", "amount": 500.0, "status": "SUCCESS"}'

echo "[Financial Shield] Firing 5 concurrent requests..."

# Run 5 requests in parallel and capture HTTP codes
> /tmp/results.txt
for i in {1..5}; do
    curl -s -o /dev/null -w "%{http_code}\n" -X POST "$API_URL" \
         -H "Content-Type: application/json" \
         -H "Authorization: $AUTH_TOKEN" \
         -H "X-Org-ID: $ORG_ID" \
         -H "X-Request-ID: $REQUEST_ID" \
         -d "$PAYLOAD" >> /tmp/results.txt &
done
wait

RESULTS=$(cat /tmp/results.txt)
echo -e "\n[Results Analysis]"
SUCCESS_COUNT=$(echo "$RESULTS" | grep -c "201")
CONFLICT_COUNT=$(echo "$RESULTS" | grep -c "409")

echo "Created (201): $SUCCESS_COUNT times"
echo "Conflicts (409): $CONFLICT_COUNT times"

if [ "$SUCCESS_COUNT" -eq 1 ] && [ "$CONFLICT_COUNT" -eq 4 ]; then
    echo "✅ SUCCESS: Idempotency Financial Shield works perfectly!"
    exit 0
else
    echo "❌ FAILURE: Unexpected results."
    exit 1
fi
