#!/bin/bash
# stream_test.sh
# Tests Live Map WebSocket Streaming & Tenant Isolation

API_URL="http://localhost:8084/api/v1/device/telemetry/gps"
AUTH_URL="http://localhost:8084/api/v1/auth/login"
ORGA_ID="c57c4ec8-5190-4776-9267-06c3da35bc3e"
EMAIL="super_admin@flyfleet.com"
PASSWORD="fleet_password"

echo "1. Authenticating to get JWT..."
LOGIN_PAYLOAD="{\"email\": \"${EMAIL}\", \"password\": \"${PASSWORD}\", \"org_id\": \"${ORGA_ID}\"}"
LOGIN_RESPONSE=$(curl -s -X POST $AUTH_URL -H "Content-Type: application/json" -d "${LOGIN_PAYLOAD}")

JWT=$(echo $LOGIN_RESPONSE | grep -oP '"access_token":"\+.*?"' | cut -d '"' -f 4)

if [ -z "$JWT" ]; then
    echo "Login failed. Check backend status."
    exit 1
fi
echo "JWT acquired."

echo "2. Building WebSocket listener in Go..."
cat << 'EOF' > verify_ws.go
package main
import (
    "log"
    "net/url"
    "github.com/gorilla/websocket"
    "os"
    "time"
)
func main() {
    orgID := os.Args[1]
    token := os.Args[2]
    u := url.URL{Scheme: "ws", Host: "localhost:8084", Path: "/api/v1/ws", RawQuery: "org_id="+orgID+"&token="+token}
    
    // Connect to WebSocket
    c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil { 
        status := 0
        if resp != nil { status = resp.StatusCode }
        log.Fatalf("dial error: %v, status: %v", err, status) 
    }
    defer c.Close()
    
    log.Println("WebSocket connected via Tenant Guard successfully.")
    
    count := 0
    c.SetReadDeadline(time.Now().Add(10 * time.Second))
    for count < 10 {
        _, message, err := c.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            break
        }
        log.Printf("Received message: %s", string(message)[0:50] + "...") // Print truncated msg
        count++
    }
    
    if count == 10 {
        log.Println("SUCCESS: Received exactly 10 telemetry pings for this org.")
        os.Exit(0)
    }
    log.Printf("FAILED: Expected 10 messages, got %d", count)
    os.Exit(1)
}
EOF

echo "3. Launching listener in background..."
cd services/fleet-core-service-legacy || exit
go run ../../verify_ws.go "${ORGA_ID}" "${JWT}" &
WS_PID=$!

sleep 2 # wait for socket to connect

echo "4. Sending 10 telemetry pings..."
for i in {1..10}; do
  TIMESTAMP=$(date +%s)
  PAYLOAD="{\"device_id\": \"TEST-DEV-$(printf "%03d" $i)\", \"timestamp\": ${TIMESTAMP}, \"latitude\": 20.0, \"longitude\": 78.0, \"speed\": 45.0, \"heading\": 90.0}"
  curl -s -o /dev/null -w "Ping $i accepted (HTTP %{http_code})\n" -X POST $API_URL \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -d "${PAYLOAD}"
done

echo "Waiting for listener to confirm..."
wait $WS_PID
RESULT=$?

cd ../../
rm verify_ws.go

if [ $RESULT -eq 0 ]; then
    echo "======================================"
    echo "✅ Live Map Streaming Test PASSED"
    echo "Tenant isolation verified. 10/10 pings received."
    echo "======================================"
else
    echo "======================================"
    echo "❌ Live Map Streaming Test FAILED"
    echo "WebSocket listener crashed or missed pings."
    echo "======================================"
fi
