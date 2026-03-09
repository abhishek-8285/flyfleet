package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func TestAlertingService_EnsuresJetStreamAndPublishes(t *testing.T) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, err := nats.Connect(natsURL, nats.Timeout(2*time.Second))
	if err != nil {
		t.Skipf("nats unavailable: %v", err)
	}
	defer nc.Close()

	svc := NewAlertingService(nc, zap.NewNop())
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("jetstream unavailable: %v", err)
	}

	if _, err := js.StreamInfo("FLYFLEET_ALERTS"); err != nil {
		t.Fatalf("expected FLYFLEET_ALERTS stream to exist: %v", err)
	}

	if err := svc.Publish(context.Background(), eventFuelTheftDetected, uuid.New(), map[string]interface{}{
		"vehicle_id": "test-vehicle",
		"risk_level": "HIGH",
	}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
}
