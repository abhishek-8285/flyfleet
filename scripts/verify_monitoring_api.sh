#!/bin/bash

# Configuration
API_URL="http://localhost:8081/api/v1/admin/monitoring"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZG1pbi0xMjMiLCJyb2xlIjoic3VwZXJfYWRtaW4iLCJwZXJtaXNzaW9ucyI6WyIqIl0sImV4cCI6MTc3MDY2NDExM30.FkjH-WdtdKzQKtU_MCgI1cS_XtnkuiY457QqkHGLVb8"
HEADER="Authorization: Bearer $JWT_TOKEN"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "Starting Monitoring API E2E Verification..."
echo "Target: $API_URL"
echo "----------------------------------------"

check_endpoint() {
    local endpoint=$1
    local name=$2
    
    echo -n "Checking $name... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" -H "$HEADER" "$API_URL/$endpoint")
    
    if [ "$response" == "200" ]; then
        echo -e "${GREEN}PASSED${NC} (200 OK)"
        # Optional: Print first few bytes of body to confirm it's JSON
        # curl -s -H "$HEADER" "$API_URL/$endpoint" | head -c 50
        # echo "..."
    else
        echo -e "${RED}FAILED${NC} (Status: $response)"
        # Print error body for debugging
        curl -s -H "$HEADER" "$API_URL/$endpoint"
        echo ""
    fi
}

# 1. Health
check_endpoint "health" "System Health"

# 2. Metrics
check_endpoint "metrics" "Performance Metrics"

# 3. Logs
check_endpoint "logs" "Error Logs"

# 4. Queues
check_endpoint "queues" "Job Queues"

# 5. SLA
check_endpoint "sla" "SLA Reports"

echo "----------------------------------------"
echo "Verification Complete."
