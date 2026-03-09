//go:build integration
// +build integration

package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestOperational_QueueSaturation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &OperationalMockRepo{}
	
	dlq := NewDLQManager(repo, nil, logger)

	// SCENARIO: 10,000 pending events (exceeds pendingThreshold=5000)
	repo.CountPendingEventsFunc = func(ctx context.Context) (int64, error) {
		return 10000, nil
	}
	repo.CountDLQFunc = func(ctx context.Context) (int64, error) {
		return 0, nil
	}

	// Trigger scan manually
	dlq.scan(context.Background())

	if !dlq.IsBackpressurePaused(context.Background()) {
		t.Error("Expected backpressure to PAUSE routing due to queue saturation")
	}

	// SCENARIO: Queue drains to 1000 (below half of threshold)
	repo.CountPendingEventsFunc = func(ctx context.Context) (int64, error) {
		return 1000, nil
	}
	dlq.scan(context.Background())

	if dlq.IsBackpressurePaused(context.Background()) {
		t.Error("Expected backpressure to RESUME routing after queue drain")
	}
}

func TestOperational_DLQOverflowAlert(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &OperationalMockRepo{}
	
	dlq := NewDLQManager(repo, nil, logger)
	alerts := dlq.Alerts()

	// SCENARIO: DLQ grows to 100 entries (exceeds dlqAlertThreshold=50)
	repo.CountDLQFunc = func(ctx context.Context) (int64, error) {
		return 100, nil
	}
	repo.CountPendingEventsFunc = func(ctx context.Context) (int64, error) {
		return 0, nil
	}

	// Trigger scan
	dlq.scan(context.Background())

	select {
	case alert := <-alerts:
		if alert.Count != 100 {
			t.Errorf("Expected alert count 100, got %d", alert.Count)
		}
		if alert.Message == "" {
			t.Error("Expected alert message, got empty")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timed out waiting for DLQ overflow alert")
	}
}

func TestOperational_ReplayIdempotency(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &OperationalMockRepo{}
	
	var callCount int32
	conn := &mockOrderingConnector{
		callback: func(entityID, eventType string) {
			atomic.AddInt32(&callCount, 1)
		},
	}
	
	registry := NewConnectorRegistry(logger)
	registry.Register("test-conn", conn)
	
	bus := NewInMemoryEventBus(logger)
	dlq := NewDLQManager(repo, nil, logger)
	dispatcher := NewWebhookDispatcher(repo, nil, nil, logger)
	router := NewEventRouter(bus, dlq, dispatcher, registry, nil, nil, repo, nil, logger)
	_ = router

	ctx := context.Background()
	tenantID := uuid.New()
	eventID := uuid.New()
	
	env := EventEnvelope{
		EventID:       eventID,
		TenantID:      tenantID,
		EntityID:      "entity-1",
		EventType:     "test.event",
		SchemaVersion: "1.0",
		Payload:       json.RawMessage(`{}`),
	}

	// 1. First processing
	_ = bus.Publish(ctx, "test.event", env)
	time.Sleep(200 * time.Millisecond)
	
	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatalf("Expected 1 call, got %d", callCount)
	}

	// 2. Replay the SAME event (check idempotency log)
	_ = bus.Publish(ctx, "test.event", env)
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Replay safety failure: expected total 1 call due to idempotency, got %d", callCount)
	}

	// 3. Replay with IsReplay=true
	env.IsReplay = true
	_ = bus.Publish(ctx, "test.event", env)
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Replay safety failure: expected total 2 calls because IsReplay=true bypassed idempotency, got %d", callCount)
	}
}

func TestHardening_RefreshStorm(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &OperationalMockRepo{}
	
	initCount := int32(0)
	conn := &FaultyConnector{
		initializeHook: func() {
			atomic.AddInt32(&initCount, 1)
			time.Sleep(100 * time.Millisecond) // Simulate slow init
		},
	}
	conn.errToReturn = ErrUnauthorized
	
	registry := NewConnectorRegistry(logger)
	registry.Register("storm-conn", conn)
	
	bus := NewInMemoryEventBus(logger)
	router := NewEventRouter(bus, nil, nil, registry, nil, nil, repo, nil, logger)
	_ = router

	ctx := context.Background()
	
	// Burst of 10 concurrent events for the same connector
	numEvents := 10
	for i := 0; i < numEvents; i++ {
		env := EventEnvelope{
			EventID:       uuid.New(),
			TenantID:      uuid.New(),
			EntityID:      "entity-1",
			EventType:     TopicTrip,
			SchemaVersion: "1.0",
			Payload:       json.RawMessage(`{}`),
		}
		_ = bus.Publish(ctx, TopicTrip, env)
	}

	time.Sleep(500 * time.Millisecond)

	// Due to the refreshMu lock, initCount should be 1 even if 10 events hit 401 simultaneously
	finalInitCount := atomic.LoadInt32(&initCount)
	if finalInitCount != 1 {
		t.Errorf("Refresh storm protection failure: expected 1 Initialize call, got %d", finalInitCount)
	}
}

func TestHardening_MixedLatencyIsolation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &OperationalMockRepo{}
	
	// Fast connector
	fastProc := int32(0)
	fastConn := &FaultyConnector{
		initializeHook: func() { atomic.AddInt32(&fastProc, 1) }, // just to tracking
	}
	
	// Slow connector
	slowConn := &FaultyConnector{
		shouldHang: true, // will take 100ms per event in FaultyConnector implementation
	}
	
	registry := NewConnectorRegistry(logger)
	registry.Register("fast-conn", fastConn)
	registry.Register("slow-conn", slowConn)
	
	bus := NewInMemoryEventBus(logger)
	// We need a dispatcher but can mock it
	router := NewEventRouter(bus, nil, nil, registry, nil, nil, repo, nil, logger)
	_ = router

	ctx := context.Background()
	numEvents := 10000 // Scaled to 10k for stress verification
	
	start := time.Now()
	for i := 0; i < numEvents; i++ {
		env := EventEnvelope{
			EventID:       uuid.New(),
			TenantID:      uuid.New(),
			EntityID:      fmt.Sprintf("entity-%d", i%50), // Spread across 50 partition groups
			EventType:     TopicTrip,
			SchemaVersion: "1.0",
			Payload:       json.RawMessage(`{}`),
		}
		_ = bus.Publish(ctx, TopicTrip, env)
	}

	// Wait for processing. Fast connector should handle 10k nearly instantly 
	// (worker pools are concurrent). Slow connector hangs for 100ms.
	time.Sleep(2 * time.Second)
	
	fastCount := atomic.LoadInt32(&fastConn.processed) 
	slowCount := atomic.LoadInt32(&slowConn.processed)
	
	// Verification
	if fastCount < int32(numEvents) {
		t.Errorf("Isolation failure at scale: Fast connector only processed %d/%d events in 2s", fastCount, numEvents)
	}
	
	if slowCount >= int32(numEvents) {
		t.Errorf("Test error: Slow connector processed all events too quickly? %d", slowCount)
	}

	logger.Info("Mixed latency isolation verified", 
		zap.Int32("fast_processed", fastCount), 
		zap.Int32("slow_processed", slowCount),
		zap.Duration("duration", time.Since(start)))
}



func TestOperational_CredentialRetryLifecycle(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &OperationalMockRepo{}
	
	conn := &FaultyConnector{}
	
	registry := NewConnectorRegistry(logger)
	registry.Register("retry-conn", conn)
	
	bus := NewInMemoryEventBus(logger)
	dlq := NewDLQManager(repo, nil, logger)
	router := NewEventRouter(bus, dlq, nil, registry, nil, nil, repo, nil, logger)
	_ = router

	ctx := context.Background()
	env := EventEnvelope{
		EventID:       uuid.New(),
		TenantID:      uuid.New(),
		EntityID:      "retry-1",
		EventType:     TopicTrip,
		SchemaVersion: "1.0",
		Payload:       json.RawMessage(`{}`),
	}

	// 1. Dispatch event - connector will error with unauthorized
	conn.errToReturn = ErrUnauthorized 
	
	_ = bus.Publish(ctx, TopicTrip, env)
	time.Sleep(200 * time.Millisecond)

	// Verify it attempted re-initialization (Initialize adds 100 to processed)
	if atomic.LoadInt32(&conn.processed) < 100 {
		t.Errorf("Expected re-initialization attempt, processed count indicates it didn't happen: %d", conn.processed)
	}
}
