package queue

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestMessageSerialization(t *testing.T) {
	now := time.Now().UTC()
	msg := Message{
		ID:         "msg_123",
		TenantID:   "tenant_456",
		Type:       "order.created",
		OccurredAt: now,
		Payload:    json.RawMessage(`{"order_id":"123"}`),
		Metadata:   map[string]string{"env": "test"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal(Message) unexpected error: %v", err)
	}

	var parsed Message
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(Message) unexpected error: %v", err)
	}

	if parsed.ID != msg.ID {
		t.Errorf("parsed.ID = %q; want %q", parsed.ID, msg.ID)
	}
	if parsed.TenantID != msg.TenantID {
		t.Errorf("parsed.TenantID = %q; want %q", parsed.TenantID, msg.TenantID)
	}
	if parsed.Type != msg.Type {
		t.Errorf("parsed.Type = %q; want %q", parsed.Type, msg.Type)
	}
}

func TestNilNATSBus(t *testing.T) {
	var bus *NATSBus

	err := bus.PublishTransient("subject.test", []byte("hello"))
	if err == nil {
		t.Error("PublishTransient on nil NATSBus expected error, got nil")
	}

	_, err = bus.SubscribeTransient("subject.test", func(data []byte) {})
	if err == nil {
		t.Error("SubscribeTransient on nil NATSBus expected error, got nil")
	}
}

func TestNewNATSBusConnectError(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	_, err := NewNATSBus("nats://invalid-host-65530:4222", logger)
	if err == nil {
		t.Error("NewNATSBus() expected error for unreachable host, got nil")
	}
}
