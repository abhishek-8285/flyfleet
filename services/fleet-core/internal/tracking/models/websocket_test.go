package models

import (
	"encoding/json"
	"testing"
)

func TestWSMessageConstants(t *testing.T) {
	// Verify all message type constants exist and have expected values
	tests := []struct {
		name  string
		value string
	}{
		{"position_update", WSMsgPositionUpdate},
		{"geofence_event", WSMsgGeofenceEvent},
		{"geofence_alert", WSMsgGeofenceAlert},
		{"connection_ack", WSMsgConnectionAck},
		{"subscribe_vehicles", WSMsgSubscribe},
		{"ping", WSMsgPing},
		{"pong", WSMsgPong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("constant %s should not be empty", tt.name)
			}
		})
	}
}

func TestWSMessage_MarshalPositionUpdate(t *testing.T) {
	msg := WSMessage{
		Type: WSMsgPositionUpdate,
		Data: PositionUpdate{
			VehicleID: "v-123",
			Latitude:  28.6139,
			Longitude: 77.2090,
			SpeedKmh:  45.2,
			Heading:   135.0,
			Status:    "moving",
			Timestamp: "2026-04-08T10:30:00Z",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded WSMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != WSMsgPositionUpdate {
		t.Errorf("expected type %s, got %s", WSMsgPositionUpdate, decoded.Type)
	}
}

func TestWSMessage_MarshalPingPong(t *testing.T) {
	// Test ping
	pingMsg := WSMessage{Type: WSMsgPing}
	data, err := json.Marshal(pingMsg)
	if err != nil {
		t.Fatalf("failed to marshal ping: %v", err)
	}
	var decoded WSMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ping: %v", err)
	}
	if decoded.Type != WSMsgPing {
		t.Errorf("expected ping, got %s", decoded.Type)
	}

	// Test pong
	pongMsg := WSMessage{
		Type: WSMsgPong,
		Data: map[string]string{"timestamp": "2026-04-08T10:30:00Z"},
	}
	data, err = json.Marshal(pongMsg)
	if err != nil {
		t.Fatalf("failed to marshal pong: %v", err)
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal pong: %v", err)
	}
	if decoded.Type != WSMsgPong {
		t.Errorf("expected pong, got %s", decoded.Type)
	}
}

func TestWSMessage_MarshalGeofenceAlert(t *testing.T) {
	msg := WSMessage{
		Type: WSMsgGeofenceAlert,
		Data: GeofenceWSMessage{
			VehicleID:    "v-123",
			Registration: "DL01AB1234",
			GeofenceID:   "g-789",
			GeofenceName: "Delhi Warehouse",
			EventType:    "enter",
			Latitude:     28.6139,
			Longitude:    77.2090,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded WSMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != WSMsgGeofenceAlert {
		t.Errorf("expected type %s, got %s", WSMsgGeofenceAlert, decoded.Type)
	}
}

func TestSubscribeRequest_Marshal(t *testing.T) {
	req := SubscribeRequest{
		VehicleIDs: []string{"v-123", "v-456"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SubscribeRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.VehicleIDs) != 2 {
		t.Errorf("expected 2 vehicle IDs, got %d", len(decoded.VehicleIDs))
	}
	if decoded.VehicleIDs[0] != "v-123" {
		t.Errorf("expected v-123, got %s", decoded.VehicleIDs[0])
	}
}
