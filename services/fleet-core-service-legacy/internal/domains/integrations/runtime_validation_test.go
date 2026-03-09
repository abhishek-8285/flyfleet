//go:build integration
// +build integration

package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"
)

func TestEventRouter_ProductionHardening(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewConnectorRegistry(logger)

	// 1. Chaos Test: Panic Recovery
	panicConn := &FaultyConnector{shouldPanic: true}
	registry.Register("panic-conn", panicConn)

	// 2. Load Test: Slow Connector Isolation
	hangConn := &FaultyConnector{shouldHang: true}
	registry.Register("hang-conn", hangConn)

	// 3. Success Connector
	successConn := &FaultyConnector{}
	registry.Register("success-conn", successConn)

	bus := NewInMemoryEventBus(logger)
	repo := &MockRepository{}
	dispatcher := NewWebhookDispatcher(repo, nil, nil, logger)
	router := NewEventRouter(bus, nil, dispatcher, registry, nil, nil, repo, nil, logger)
	_ = router // keep alive

	ctx := context.Background()
	tenantID := uuid.New()

	// Dispatch a batch of events
	numEvents := 100
	for i := 0; i < numEvents; i++ {
		env := EventEnvelope{
			EventID:       uuid.New(),
			EventType:     TopicTrip,
			TenantID:      tenantID,
			EntityID:      fmt.Sprintf("entity-%d", i%10), // 10 distinct entities
			SchemaVersion: "1.0",
			Payload:       json.RawMessage(`{}`),
		}
		_ = bus.Publish(ctx, TopicTrip, env)
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Verify Isolation: success-conn should have processed all 100 events
	// despite panic-conn crashing and hang-conn being stuck.
	if atomic.LoadInt32(&successConn.processed) != int32(numEvents) {
		t.Errorf("Isolation failure: expected success-conn to process %d, got %d", numEvents, successConn.processed)
	}

	logger.Info("Production hardening verified: Success connector unaffected by failures.")
}

func TestEventRouter_OrderingVerification(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewConnectorRegistry(logger)

	processedEvents := make(map[string][]string)
	var mu sync.Mutex

	orderingConn := &mockOrderingConnector{
		callback: func(entityID, eventType string) {
			mu.Lock()
			processedEvents[entityID] = append(processedEvents[entityID], eventType)
			mu.Unlock()
		},
	}
	registry.Register("ordering-conn", orderingConn)

	bus := NewInMemoryEventBus(logger)
	repo := &MockRepository{}
	dispatcher := NewWebhookDispatcher(repo, nil, nil, logger)
	_ = NewEventRouter(bus, nil, dispatcher, registry, nil, nil, repo, nil, logger)

	ctx := context.Background()
	tenantID := uuid.New()
	entityID := "trip-123"

	// Sequence of events for the same entity
	events := []string{"created", "started", "completed"}
	for _, et := range events {
		bus.Publish(ctx, TopicTrip, EventEnvelope{
			EventID:       uuid.New(),
			EventType:     et,
			TenantID:      tenantID,
			EntityID:      entityID,
			SchemaVersion: "1.0",
			Payload:       json.RawMessage(`{}`),
		})
	}

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	seq := processedEvents[entityID]
	if len(seq) != 3 {
		t.Fatalf("Expected 3 events for %s, got %d", entityID, len(seq))
	}

	for i, et := range events {
		if seq[i] != et {
			t.Errorf("Out of order delivery for %s: expected %s at index %d, got %s", entityID, et, i, seq[i])
		}
	}
}
