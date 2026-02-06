#!/bin/bash

# Trip Service Test Script
# Tests the trip service endpoints

echo "🚀 Testing Trip Service"

# Service configuration
SERVICE_URL="http://localhost:8082"
HEALTH_ENDPOINT="$SERVICE_URL/health"

# Wait for service to be ready
echo "⏳ Waiting for trip service to start..."
for i in {1..30}; do
    if curl -s "$HEALTH_ENDPOINT" >/dev/null 2>&1; then
        echo "✅ Trip service is ready"
        break
    fi
    echo "⏳ Waiting... ($i/30)"
    sleep 2
done

# Test health endpoint
echo "🏥 Testing health endpoint..."
curl -s "$HEALTH_ENDPOINT" | jq '.'
echo ""

# Test trip creation
echo "🚗 Testing trip creation..."
CREATE_RESPONSE=$(curl -s -X POST "$SERVICE_URL/api/v1/trips" \
    -H "Content-Type: application/json" \
    -d '{
        "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
        "driver_id": "550e8400-e29b-41d4-a716-446655440001",
        "title": "Test Trip to Mumbai",
        "description": "Delivery trip to Mumbai warehouse",
        "planned_route": [
            {
                "id": "wp1",
                "name": "Warehouse",
                "type": "start",
                "address": "Bangalore Warehouse",
                "location": {
                    "latitude": 12.9716,
                    "longitude": 77.5946
                },
                "sequence": 1
            },
            {
                "id": "wp2", 
                "name": "Customer Site",
                "type": "end",
                "address": "Mumbai Customer",
                "location": {
                    "latitude": 19.0760,
                    "longitude": 72.8777
                },
                "sequence": 2
            }
        ],
        "waypoints": [
            {
                "id": "wp1",
                "name": "Warehouse",
                "type": "pickup",
                "address": "Bangalore Warehouse",
                "location": {
                    "latitude": 12.9716,
                    "longitude": 77.5946
                },
                "sequence": 1
            },
            {
                "id": "wp2",
                "name": "Customer Site", 
                "type": "dropoff",
                "address": "Mumbai Customer",
                "location": {
                    "latitude": 19.0760,
                    "longitude": 72.8777
                },
                "sequence": 2
            }
        ],
        "estimated_distance_km": 950.5,
        "estimated_duration_min": 720,
        "cargo_details": {
            "type": "electronics",
            "weight_kg": 500.0,
            "volume_m3": 15.0,
            "value_inr": 250000.0,
            "description": "Electronics shipment",
            "fragile": true,
            "temperature": "ambient"
        },
        "scheduled_start": "2024-01-15T08:00:00Z",
        "customer_name": "ABC Corporation",
        "customer_phone": "+919876543210",
        "customer_email": "contact@abccorp.com"
    }')

echo "📝 Create response:"
echo "$CREATE_RESPONSE" | jq '.'
echo ""

# Extract trip ID for subsequent tests
TRIP_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
if [ "$TRIP_ID" != "null" ] && [ -n "$TRIP_ID" ]; then
    echo "📋 Created trip ID: $TRIP_ID"
    
    # Test get trip
    echo "🔍 Testing get trip..."
    curl -s "$SERVICE_URL/api/v1/trips/$TRIP_ID" | jq '.'
    echo ""
    
    # Test list trips
    echo "📋 Testing list trips..."
    curl -s "$SERVICE_URL/api/v1/trips?page=1&page_size=10" | jq '.'
    echo ""
    
    # Test start trip
    echo "🏁 Testing start trip..."
    START_RESPONSE=$(curl -s -X POST "$SERVICE_URL/api/v1/trips/$TRIP_ID/start" \
        -H "Content-Type: application/json" \
        -d '{
            "odometer_reading": 15000.5,
            "fuel_level": 85.5,
            "notes": "Starting delivery trip"
        }')
    echo "$START_RESPONSE" | jq '.'
    echo ""
    
    # Test add event
    echo "🔔 Testing add event..."
    curl -s -X POST "$SERVICE_URL/api/v1/trips/$TRIP_ID/events" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "departure",
            "location": {
                "latitude": 12.9716,
                "longitude": 77.5946,
                "address": "Bangalore Warehouse"
            },
            "description": "Departed from warehouse",
            "photos": []
        }' -o /dev/null -w "Status: %{http_code}\n"
    echo ""
    
    # Test complete trip
    echo "✅ Testing complete trip..."
    COMPLETE_RESPONSE=$(curl -s -X POST "$SERVICE_URL/api/v1/trips/$TRIP_ID/complete" \
        -H "Content-Type: application/json" \
        -d '{
            "odometer_reading": 15950.5,
            "fuel_level": 25.0,
            "actual_distance_km": 950.0,
            "fuel_consumed_liters": 85.0,
            "idling_time_minutes": 45,
            "pod_documents": [
                {
                    "type": "signature",
                    "url": "https://storage.example.com/signature.jpg",
                    "recipient": {
                        "name": "John Smith",
                        "phone": "+919876543210",
                        "email": "john@customer.com"
                    },
                    "otp_verified": true,
                    "signature": "base64_signature_data"
                }
            ],
            "notes": "Successful delivery completed"
        }')
    echo "$COMPLETE_RESPONSE" | jq '.'
    echo ""
    
    # Test optimize route
    echo "🛣️ Testing route optimization..."
    curl -s "$SERVICE_URL/api/v1/trips/$TRIP_ID/optimize-route" | jq '.'
    echo ""
    
else
    echo "❌ Failed to create trip, skipping dependent tests"
fi

# Test trip summary
echo "📊 Testing trip summary..."
curl -s "$SERVICE_URL/api/v1/trips/summary?start_date=2024-01-01&end_date=2024-12-31" | jq '.'
echo ""

echo "🎉 Trip service tests completed!"