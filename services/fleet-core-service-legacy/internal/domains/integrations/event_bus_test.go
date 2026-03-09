//go:build integration
// +build integration

package integrations

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"
)

func TestInMemoryEventBus_PartitionOrdering(t *testing.T) {
	logger := zaptest.NewLogger(t)
	bus := NewInMemoryEventBus(logger)
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var mu sync.Mutex
	var processed []string

	bus.Subscribe(TopicTrip, func(ctx context.Context, env EventEnvelope) {
		mu.Lock()
		processed = append(processed, env.EventType)
		mu.Unlock()
		// Simulate small delay to test concurrency bounds
		time.Sleep(10 * time.Millisecond)
	})

	tenantID := uuid.New()
	entityID := "trip_1001"

	// Publish 3 events for the exact same entity in rapid succession
	_ = bus.Publish(ctx, TopicTrip, EventEnvelope{EventID: uuid.New(), EventType: "trip.created", TenantID: tenantID, EntityID: entityID})
	_ = bus.Publish(ctx, TopicTrip, EventEnvelope{EventID: uuid.New(), EventType: "trip.started", TenantID: tenantID, EntityID: entityID})
	_ = bus.Publish(ctx, TopicTrip, EventEnvelope{EventID: uuid.New(), EventType: "trip.completed", TenantID: tenantID, EntityID: entityID})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(processed) != 3 {
		t.Fatalf("expected 3 events to be processed, got %d", len(processed))
	}

	// Verify STRICT ordering. Because they share an EntityID, they must execute sequentially on the same worker.
	if processed[0] != "trip.created" || processed[1] != "trip.started" || processed[2] != "trip.completed" {
		t.Errorf("events were processed out of order: %v", processed)
	}
}
