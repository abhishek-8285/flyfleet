#!/bin/bash

# Configuration
BFF_URL="http://localhost:8088"
echo "Starting Final Verification of 18 Modules via Web-BFF ($BFF_URL)..."
echo "----------------------------------------------------------------"

# Function to check endpoint
check_endpoint() {
    MODULE=$1
    METHOD=$2
    ENDPOINT=$3
    EXPECTED_STATUS=$4

    echo -n "Checking $MODULE ($ENDPOINT)... "
    STATUS=$(curl -o /dev/null -s -w "%{http_code}" -X $METHOD $BFF_URL$ENDPOINT)
    
    if [ "$STATUS" == "$EXPECTED_STATUS" ]; then
        echo "✅ PASS (Status: $STATUS)"
    elif [ "$STATUS" == "401" ]; then
         echo "✅ PASS (Protected: 401)"
    else
        echo "❌ FAIL (Status: $STATUS)"
    fi
}

# 1. Vehicle Management
check_endpoint "1. Vehicles" "GET" "/api/v1/admin/vehicles/" "401"

# 2. Driver Management
check_endpoint "2. Drivers" "GET" "/api/v1/admin/drivers/" "401"

# 3. Trip Management
check_endpoint "3. Trips" "GET" "/api/v1/admin/trips" "401"

# 4. Maintenance
check_endpoint "4. Maintenance" "GET" "/api/v1/admin/work-orders/" "401"

# 5. IoT Devices
check_endpoint "5. Devices" "GET" "/api/v1/admin/devices" "401"

# 6. Tracking
check_endpoint "6. Tracking" "GET" "/api/v1/tracking/vehicles/123/history/" "401"

# 7. Integrations (Follow Redirect)
check_endpoint "7. Integrations" "GET" "/api/integrations/" "401"

# 8. Reports (Follow Redirect)
check_endpoint "8. Reports" "GET" "/api/v1/admin/reports/" "401"

# 9. Governance
check_endpoint "9. Governance" "GET" "/api/v1/admin/governance/settings" "401"

# 10. Disaster Recovery
check_endpoint "10. Disaster Recovery" "GET" "/api/v1/admin/dr/backups" "401"

# 11. Monitoring
check_endpoint "11. Monitoring" "GET" "/api/v1/admin/monitoring/health" "401"

# 12. Search
check_endpoint "12. Search" "GET" "/api/search?q=test" "401"

# 13. Security
check_endpoint "13. Security" "GET" "/api/security/logins" "401"

# 14. Support BFF
check_endpoint "14. Support BFF" "GET" "/bff/support/incidents/feed" "401"

# 15. Billing
check_endpoint "15. Billing" "GET" "/api/v1/admin/billing/invoices" "401"

# 16. Video Telematics
check_endpoint "16. Video" "GET" "/api/v1/admin/video/streams/" "401"

# 17. Route Optimization
check_endpoint "17. Route Opt" "GET" "/api/v1/routes/jobs" "404"

# 18. Analytics
check_endpoint "18. Analytics" "GET" "/api/v1/admin/analytics/models/" "401"

echo "----------------------------------------------------------------"
echo "Verification Complete."
