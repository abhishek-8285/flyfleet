#!/bin/bash

# Driver Service Integration Tests
# Tests all driver service endpoints and functionality

set -e

echo "🚀 Starting Driver Service Integration Tests"
echo "============================================"

# Configuration
DRIVER_SERVICE_URL="http://localhost:8083"
HEALTH_ENDPOINT="$DRIVER_SERVICE_URL/health"
DRIVERS_ENDPOINT="$DRIVER_SERVICE_URL/api/v1/drivers"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TEST_COUNT=0
PASS_COUNT=0
FAIL_COUNT=0

# Helper functions
print_header() {
    echo -e "${BLUE}$1${NC}"
    echo "----------------------------------------"
}

print_test() {
    TEST_COUNT=$((TEST_COUNT + 1))
    echo -n "Test $TEST_COUNT: $1... "
}

print_pass() {
    PASS_COUNT=$((PASS_COUNT + 1))
    echo -e "${GREEN}PASS${NC}"
}

print_fail() {
    FAIL_COUNT=$((FAIL_COUNT + 1))
    echo -e "${RED}FAIL${NC}"
    if [ -n "$1" ]; then
        echo -e "${RED}Error: $1${NC}"
    fi
}

# Wait for service to be ready
wait_for_service() {
    print_header "Waiting for Driver Service to start..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$HEALTH_ENDPOINT" >/dev/null 2>&1; then
            echo -e "${GREEN}Driver Service is ready!${NC}"
            return 0
        fi
        
        echo "Attempt $attempt/$max_attempts: Service not ready, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}Driver Service failed to start within timeout${NC}"
    return 1
}

# Test health endpoint
test_health_endpoint() {
    print_test "Health endpoint"
    
    response=$(curl -s -w "%{http_code}" "$HEALTH_ENDPOINT")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test create driver
test_create_driver() {
    print_test "Create driver"
    
    driver_data='{
        "first_name": "Test",
        "last_name": "Driver",
        "email": "test.driver@example.com",
        "phone": "+919999999999",
        "license_number": "DLTEST001",
        "license_type": "LMV",
        "license_expiry": "2027-12-31",
        "license_authority": "Transport Department, Test State",
        "date_of_birth": "1990-01-01",
        "address": {
            "street": "123 Test Street",
            "city": "Test City",
            "state": "Test State",
            "pincode": "123456",
            "country": "India"
        },
        "emergency_contact": {
            "name": "Emergency Contact",
            "phone": "+918888888888",
            "relation": "Friend"
        },
        "experience_years": 5
    }'
    
    response=$(curl -s -w "%{http_code}" \
        -H "Content-Type: application/json" \
        -H "X-Org-ID: test-org" \
        -d "$driver_data" \
        "$DRIVERS_ENDPOINT")
    
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "201" ]; then
        DRIVER_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$DRIVER_ID" ]; then
            print_pass
        else
            print_fail "Could not extract driver ID from response"
        fi
    else
        print_fail "Expected 201, got $http_code. Response: $body"
    fi
}

# Test get driver by ID
test_get_driver() {
    print_test "Get driver by ID"
    
    if [ -z "$DRIVER_ID" ]; then
        print_fail "No driver ID available"
        return
    fi
    
    response=$(curl -s -w "%{http_code}" "$DRIVERS_ENDPOINT/$DRIVER_ID")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test list drivers
test_list_drivers() {
    print_test "List drivers"
    
    response=$(curl -s -w "%{http_code}" "$DRIVERS_ENDPOINT?page=1&page_size=10")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test update driver
test_update_driver() {
    print_test "Update driver"
    
    if [ -z "$DRIVER_ID" ]; then
        print_fail "No driver ID available"
        return
    fi
    
    update_data='{
        "first_name": "Updated",
        "last_name": "Driver",
        "phone": "+919999999998"
    }'
    
    response=$(curl -s -w "%{http_code}" \
        -X PUT \
        -H "Content-Type: application/json" \
        -d "$update_data" \
        "$DRIVERS_ENDPOINT/$DRIVER_ID")
    
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test update driver status
test_update_driver_status() {
    print_test "Update driver status"
    
    if [ -z "$DRIVER_ID" ]; then
        print_fail "No driver ID available"
        return
    fi
    
    status_data='{
        "status": "inactive"
    }'
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$status_data" \
        "$DRIVERS_ENDPOINT/$DRIVER_ID/update-status")
    
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test get behavior score
test_get_behavior_score() {
    print_test "Get driver behavior score"
    
    if [ -z "$DRIVER_ID" ]; then
        print_fail "No driver ID available"
        return
    fi
    
    response=$(curl -s -w "%{http_code}" "$DRIVERS_ENDPOINT/$DRIVER_ID/behavior-score")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test update behavior score
test_update_behavior_score() {
    print_test "Update driver behavior score"
    
    if [ -z "$DRIVER_ID" ]; then
        print_fail "No driver ID available"
        return
    fi
    
    behavior_data='{
        "score_change": -5,
        "reason": "Minor traffic violation"
    }'
    
    response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$behavior_data" \
        "$DRIVERS_ENDPOINT/$DRIVER_ID/update-behavior")
    
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test get driver analytics
test_get_driver_analytics() {
    print_test "Get driver analytics"
    
    if [ -z "$DRIVER_ID" ]; then
        print_fail "No driver ID available"
        return
    fi
    
    response=$(curl -s -w "%{http_code}" "$DRIVERS_ENDPOINT/$DRIVER_ID/analytics")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_pass
    else
        print_fail "Expected 200, got $http_code"
    fi
}

# Test invalid driver ID
test_invalid_driver_id() {
    print_test "Invalid driver ID handling"
    
    response=$(curl -s -w "%{http_code}" "$DRIVERS_ENDPOINT/invalid-id")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "400" ]; then
        print_pass
    else
        print_fail "Expected 400, got $http_code"
    fi
}

# Test create driver with invalid data
test_create_invalid_driver() {
    print_test "Create driver with invalid data"
    
    invalid_data='{
        "first_name": "",
        "last_name": "",
        "email": "invalid-email",
        "phone": "123"
    }'
    
    response=$(curl -s -w "%{http_code}" \
        -H "Content-Type: application/json" \
        -H "X-Org-ID: test-org" \
        -d "$invalid_data" \
        "$DRIVERS_ENDPOINT")
    
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "400" ]; then
        print_pass
    else
        print_fail "Expected 400, got $http_code"
    fi
}

# Test driver not found
test_driver_not_found() {
    print_test "Driver not found handling"
    
    fake_id="drv_nonexistent123"
    response=$(curl -s -w "%{http_code}" "$DRIVERS_ENDPOINT/$fake_id")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "404" ]; then
        print_pass
    else
        print_fail "Expected 404, got $http_code"
    fi
}

# Run all tests
run_tests() {
    print_header "Running Driver Service Tests"
    
    test_health_endpoint
    test_create_driver
    test_get_driver
    test_list_drivers
    test_update_driver
    test_update_driver_status
    test_get_behavior_score
    test_update_behavior_score
    test_get_driver_analytics
    test_invalid_driver_id
    test_create_invalid_driver
    test_driver_not_found
    
    # Summary
    echo ""
    print_header "Test Results Summary"
    echo "Total Tests: $TEST_COUNT"
    echo -e "Passed: ${GREEN}$PASS_COUNT${NC}"
    echo -e "Failed: ${RED}$FAIL_COUNT${NC}"
    
    if [ $FAIL_COUNT -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}❌ Some tests failed${NC}"
        return 1
    fi
}

# Main execution
main() {
    # Check if service is running
    if ! wait_for_service; then
        echo -e "${RED}Cannot run tests - service not available${NC}"
        exit 1
    fi
    
    # Run tests
    if run_tests; then
        exit 0
    else
        exit 1
    fi
}

# Execute main function
main