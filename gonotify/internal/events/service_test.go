package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
)

func TestPublishValidation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	svc := NewService(nil, logger)

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}

	ctx := context.Background()

	// Missing event type
	_, err := svc.Publish(ctx, PublishRequest{
		TenantID: "tenant_1",
	})
	if err == nil {
		t.Error("Publish() expected error for missing event type, got nil")
	}

	// Missing tenant ID
	_, err = svc.Publish(ctx, PublishRequest{
		EventType: "order.created",
	})
	if err == nil {
		t.Error("Publish() expected error for missing tenant ID, got nil")
	}
}

func TestPublishResponseSerialization(t *testing.T) {
	resp := PublishResponse{
		EventID: "evt_123456",
		Status:  "accepted",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal(PublishResponse) unexpected error: %v", err)
	}

	var parsed PublishResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(PublishResponse) unexpected error: %v", err)
	}

	if parsed.EventID != resp.EventID || parsed.Status != resp.Status {
		t.Errorf("parsed = %v; want %v", parsed, resp)
	}
}
