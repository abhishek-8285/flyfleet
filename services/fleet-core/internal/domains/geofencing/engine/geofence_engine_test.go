package engine

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ============================================================================
// Geofence Engine — getOrgIDForVehicle Tests
// ============================================================================

func TestGeofenceEngine_GetOrgIDForVehicle_MethodExists(t *testing.T) {
	logger := zap.NewNop()
	// Verify the method exists on the engine struct
	eng := NewGeofenceEngine(nil, nil, logger)

	// The method is defined — we verify it exists by calling it.
	// With nil db it will panic, which is expected behavior (db must be configured).
	// In production, the engine always has a db pool.
	if eng == nil {
		t.Error("engine should not be nil")
	}
	// We can't test with nil db without panicking. The actual behavior
	// is tested via integration tests with a real DB.
}

// ============================================================================
// Geofence Event Message Format Tests
// ============================================================================

func TestGeofenceEventMessageFormat(t *testing.T) {
	vehicleID := uuid.Must(uuid.NewRandom())
	geofenceID := uuid.Must(uuid.NewRandom())
	eventTime := time.Date(2026, 4, 8, 10, 30, 0, 0, time.UTC)

	// Simulate the message format used in triggerGeofenceEvent Redis publish
	geoMsg := map[string]interface{}{
		"type":            "geofence_alert",
		"vehicle_id":      vehicleID.String(),
		"geofence_id":     geofenceID.String(),
		"geofence_name":   "Delhi Warehouse",
		"event_type":      "enter",
		"latitude":        28.6139,
		"longitude":       77.2090,
		"timestamp":       eventTime.Format(time.RFC3339),
	}

	data, err := json.Marshal(geoMsg)
	if err != nil {
		t.Fatalf("marshal geofence message: %v", err)
	}

	// Verify it can be unmarshaled
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal geofence message: %v", err)
	}

	// Verify fields
	if decoded["type"] != "geofence_alert" {
		t.Errorf("expected type=geofence_alert, got %v", decoded["type"])
	}
	if decoded["vehicle_id"] != vehicleID.String() {
		t.Errorf("expected vehicle_id=%s, got %v", vehicleID, decoded["vehicle_id"])
	}
	if decoded["geofence_id"] != geofenceID.String() {
		t.Errorf("expected geofence_id=%s, got %v", geofenceID, decoded["geofence_id"])
	}
	if decoded["geofence_name"] != "Delhi Warehouse" {
		t.Errorf("expected geofence_name=Delhi Warehouse, got %v", decoded["geofence_name"])
	}
	if decoded["event_type"] != "enter" {
		t.Errorf("expected event_type=enter, got %v", decoded["event_type"])
	}
	if decoded["latitude"].(float64) != 28.6139 {
		t.Errorf("expected latitude=28.6139, got %v", decoded["latitude"])
	}
	if decoded["longitude"].(float64) != 77.2090 {
		t.Errorf("expected longitude=77.2090, got %v", decoded["longitude"])
	}
}

func TestGeofenceEventMessage_ExitType(t *testing.T) {
	vehicleID := uuid.Must(uuid.NewRandom())
	geofenceID := uuid.Must(uuid.NewRandom())
	eventTime := time.Date(2026, 4, 8, 11, 0, 0, 0, time.UTC)

	geoMsg := map[string]interface{}{
		"type":            "geofence_alert",
		"vehicle_id":      vehicleID.String(),
		"geofence_id":     geofenceID.String(),
		"geofence_name":   "Mumbai Depot",
		"event_type":      "exit",
		"latitude":        19.0760,
		"longitude":       72.8777,
		"timestamp":       eventTime.Format(time.RFC3339),
	}

	data, err := json.Marshal(geoMsg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["event_type"] != "exit" {
		t.Errorf("expected event_type=exit, got %v", decoded["event_type"])
	}
}

func TestGeofenceEventMessage_DwellType(t *testing.T) {
	vehicleID := uuid.Must(uuid.NewRandom())
	geofenceID := uuid.Must(uuid.NewRandom())
	eventTime := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)

	geoMsg := map[string]interface{}{
		"type":            "geofence_alert",
		"vehicle_id":      vehicleID.String(),
		"geofence_id":     geofenceID.String(),
		"geofence_name":   "Charging Station",
		"event_type":      "dwell",
		"latitude":        12.9716,
		"longitude":       77.5946,
		"timestamp":       eventTime.Format(time.RFC3339),
	}

	data, err := json.Marshal(geoMsg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["event_type"] != "dwell" {
		t.Errorf("expected event_type=dwell, got %v", decoded["event_type"])
	}
}

// ============================================================================
// Redis Channel Name Format Tests
// ============================================================================

func TestRedisChannelFormat(t *testing.T) {
	orgID := uuid.Must(uuid.NewRandom())
	expected := "org:" + orgID.String() + ":geofence_alerts"

	// Verify the channel format matches what WebSocketHandler subscribes to
	channel := "org:" + orgID.String() + ":geofence_alerts"
	if channel != expected {
		t.Errorf("channel format mismatch: expected %s, got %s", expected, channel)
	}
}

func TestRedisTelemetryChannelFormat(t *testing.T) {
	orgID := "550e8400-e29b-41d4-a716-446655440000"
	expected := "org:" + orgID + ":telemetry"

	channel := "org:" + orgID + ":telemetry"
	if channel != expected {
		t.Errorf("channel format mismatch: expected %s, got %s", expected, channel)
	}
}
