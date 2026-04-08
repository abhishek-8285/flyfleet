package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

// ============================================================================
// WebSocket Handler Tests (with httptest + gorilla websocket)
// ============================================================================

func TestNewWebSocketHandler(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	logger := zap.NewNop()

	h := NewWebSocketHandler(nil, "test-secret", vr, lc, logger)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.vehicleRepo != vr {
		t.Error("vehicleRepo not set")
	}
	if h.liveCache != lc {
		t.Error("liveCache not set")
	}
}

func TestWebSocketHandler_RegisterUnregisterClient(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	h := NewWebSocketHandler(nil, "test-secret", vr, lc, zap.NewNop())

	client := &WSClient{
		conn:       nil, // not needed for register/unregister
		orgID:      "test-org",
		send:       make(chan []byte, 1),
		subscribed: make(map[string]bool),
		lastSent:   make(map[string]*models.LiveLocation),
		lastPong:   time.Now(),
	}

	h.registerClient("test-org", client)

	h.mu.RLock()
	count := len(h.clients["test-org"])
	h.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 client, got %d", count)
	}

	h.unregisterClient("test-org", client)

	h.mu.RLock()
	count = len(h.clients["test-org"])
	h.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 clients after unregister, got %d", count)
	}
}

func TestWebSocketHandler_UnregisterNonExistent(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	h := NewWebSocketHandler(nil, "test-secret", vr, lc, zap.NewNop())

	client := &WSClient{
		orgID: "test-org",
		send:  make(chan []byte, 1),
	}

	// Should not panic
	h.unregisterClient("test-org", client)
}

func TestWebSocketHandler_ConnectNoToken(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	h := NewWebSocketHandler(nil, "test-secret", vr, lc, zap.NewNop())

	srv := httptest.NewServer(h)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/telemetry"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		conn.Close()
		t.Error("expected auth error without token")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestWebSocketHandler_ConnectInvalidToken(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	h := NewWebSocketHandler(nil, "test-secret", vr, lc, zap.NewNop())

	srv := httptest.NewServer(h)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/telemetry?token=invalid"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		conn.Close()
		t.Error("expected auth error with invalid token")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestWebSocketHandler_NonWebSocketRequest(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	h := NewWebSocketHandler(nil, "test-secret", vr, lc, zap.NewNop())

	srv := httptest.NewServer(h)
	defer srv.Close()

	// Regular HTTP GET — gets rejected by auth first (401), then by upgrade (400)
	resp, err := http.Get(srv.URL + "/ws/telemetry")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Without auth token, gets 401
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 401 or 400, got %d", resp.StatusCode)
	}
}

func TestWSClient_WritePump_EchoTest(t *testing.T) {
	// Create a simple echo WebSocket server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Send ping, expect it back
	pingMsg := models.WSMessage{Type: models.WSMsgPing}
	if err := conn.WriteJSON(pingMsg); err != nil {
		t.Fatalf("write: %v", err)
	}

	var resp models.WSMessage
	conn.SetReadDeadline(time.Now().Add(time.Second))
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.Type != models.WSMsgPing {
		t.Errorf("expected ping, got %s", resp.Type)
	}
}

func TestHasSignificantChange_Boundary(t *testing.T) {
	base := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}

	// Exactly at threshold (0.0001) — should NOT trigger (comparison is >, not >=)
	atBoundary := models.LiveLocation{Lat: 28.6140, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	if hasSignificantChange(base, atBoundary) {
		t.Error("expected no change at exact boundary (uses > not >=)")
	}

	// Just above boundary
	above := models.LiveLocation{Lat: 28.61401, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	if !hasSignificantChange(base, above) {
		t.Error("expected change above boundary")
	}

	// Well below boundary
	below := models.LiveLocation{Lat: 28.61399, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	if hasSignificantChange(base, below) {
		t.Error("expected no change below boundary")
	}

	// Speed exactly at threshold (2.0) — should NOT trigger
	speedAt := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 42.0, Status: models.StatusMoving}
	if hasSignificantChange(base, speedAt) {
		t.Error("expected no change at exact speed boundary")
	}

	// Speed above threshold
	speedAbove := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 42.1, Status: models.StatusMoving}
	if !hasSignificantChange(base, speedAbove) {
		t.Error("expected change above speed boundary")
	}
}

func TestWSMessage_AllTypesRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  models.WSMessage
	}{
		{"connection_ack", models.WSMessage{Type: models.WSMsgConnectionAck, Data: map[string]interface{}{"status": "connected"}}},
		{"position_update", models.WSMessage{Type: models.WSMsgPositionUpdate, Data: models.PositionUpdate{VehicleID: "v1", Latitude: 28.6, Longitude: 77.2, SpeedKmh: 45, Status: "moving"}}},
		{"geofence_alert", models.WSMessage{Type: models.WSMsgGeofenceAlert, Data: models.GeofenceWSMessage{VehicleID: "v1", Registration: "DL01AB", GeofenceName: "Warehouse", EventType: "enter"}}},
		{"subscribe_vehicles", models.WSMessage{Type: models.WSMsgSubscribe, Data: models.SubscribeRequest{VehicleIDs: []string{"v1", "v2"}}}},
		{"ping", models.WSMessage{Type: models.WSMsgPing}},
		{"pong", models.WSMessage{Type: models.WSMsgPong, Data: map[string]string{"timestamp": "2026-04-08T10:30:00Z"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded models.WSMessage
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded.Type != tt.msg.Type {
				t.Errorf("expected type=%s, got %s", tt.msg.Type, decoded.Type)
			}
		})
	}
}

// ============================================================================
// Ingest Handler Constructor Test
// ============================================================================

func TestNewIngestHandler(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	svc := &simpleIngestSvc{telemRepo: tr, liveCache: lc}

	h := NewIngestHandler(svc, zap.NewNop())
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.ingestion == nil {
		t.Error("ingestion is nil")
	}
}

// ============================================================================
// Geofence Handler Additional Tests
// ============================================================================

func TestGeofenceHandler_GetWithGeofences(t *testing.T) {
	gID := uuid.Must(uuid.NewV4())
	orgID := uuid.Must(uuid.NewV4())
	radius := 500.0
	repo := repository.NewMemGeofenceRepo()
	repo.AddGeofence(models.Geofence{ID: gID, OrgID: orgID, Name: "Warehouse A", Type: "circle", RadiusM: &radius})
	h := NewGeofenceHandler(repo, zap.NewNop())

	ctx := ctxkeys.WithOrgID(context.Background(), orgID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geofences", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("expected 1 geofence, got %d", len(data))
	}
}
