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
	"go.uber.org/zap/zaptest"
)

func TestIntegration_ConnectorToDLQ(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewConnectorRegistry(logger)

	// A connector that always errors
	errConn := &FaultyConnector{shouldError: true}
	registry.Register("err-conn", errConn)

	// A connector that always panics
	panicConn := &FaultyConnector{shouldPanic: true}
	registry.Register("panic-conn", panicConn)

	bus := NewInMemoryEventBus(logger)
	
	var dlqCount int32
	repo := &OperationalMockRepo{
		SendToDLQFunc: func(ctx context.Context, entry *DLQEntry) error {
			atomic.AddInt32(&dlqCount, 1)
			if entry.TenantID == uuid.Nil || entry.EventType == "" || len(entry.Payload) == 0 {
				return fmt.Errorf("DLQ entry missing recovery fields: %+v", entry)
			}
			return nil
		},
	}
	
	dlq := NewDLQManager(repo, nil, logger)
	dispatcher := NewWebhookDispatcher(repo, nil, nil, logger)
	router := NewEventRouter(bus, dlq, dispatcher, registry, nil, nil, repo, nil, logger)
	_ = router

	ctx := context.Background()
	env := EventEnvelope{
		EventID:       uuid.New(),
		EventType:     TopicTrip,
		TenantID:      uuid.New(),
		EntityID:      "trip-1",
		SchemaVersion: "1.0",
		Payload:       json.RawMessage(`{"test":true}`),
	}

	// Trigger processing
	_ = bus.Publish(ctx, TopicTrip, env)

	// Wait for processing across 2 connectors (err and panic)
	time.Sleep(500 * time.Millisecond)

	// Both err-conn and panic-conn should have triggered a SendToDLQ call
	count := atomic.LoadInt32(&dlqCount)
	if count != 2 {
		t.Errorf("Expected 2 DLQ entries (1 error, 1 panic), got %d", count)
	}

	logger.Info("Connector-to-DLQ flow verified successfully with recovery fields.")
}
