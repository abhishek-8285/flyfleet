package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gonotify/gonotify/internal/queue"
)

func TestSubjectFormatting(t *testing.T) {
	tenantID := "00000000-0000-0000-0000-000000000001"
	userID := "usr_test_123"

	got := Subject(tenantID, userID)
	expected := "gn.inbox.00000000-0000-0000-0000-000000000001.usr_test_123"

	if got != expected {
		t.Errorf("Subject() = %q; want %q", got, expected)
	}
}

func TestNilPublisherPublish(t *testing.T) {
	var p *Publisher
	ctx := context.Background()

	err := p.Publish(ctx, "tenant_1", "usr_1", Event{
		Type:       "inbox.item_created",
		OccurredAt: time.Now(),
	})

	if err != nil {
		t.Errorf("Nil publisher Publish() returned unexpected error: %v", err)
	}
}

func TestNewPublisher(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	pub := NewPublisher(&queue.NATSBus{}, logger)

	if pub == nil {
		t.Fatal("NewPublisher(&queue.NATSBus{}, logger) returned nil")
	}

	// Unmarshalable event payload error test
	err := pub.Publish(context.Background(), "tenant_1", "usr_1", Event{
		Type: "inbox.item_created",
		Data: map[string]any{
			"unsupported": make(chan int),
		},
	})
	if err == nil {
		t.Error("Publish expected JSON marshal error for bad Data, got nil")
	}
}

func TestEventSerialization(t *testing.T) {
	now := time.Now().UTC()
	evt := Event{
		Type:       "inbox.item_created",
		TenantID:   "tenant_123",
		UserID:     "usr_456",
		ItemID:     "item_789",
		Data:       map[string]any{"key": "value"},
		OccurredAt: now,
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("json.Marshal(Event) unexpected error: %v", err)
	}

	var parsed Event
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(Event) unexpected error: %v", err)
	}

	if parsed.Type != evt.Type || parsed.TenantID != evt.TenantID || parsed.UserID != evt.UserID {
		t.Errorf("parsed = %v; want %v", parsed, evt)
	}
}
