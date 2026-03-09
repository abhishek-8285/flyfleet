package integrations

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── EventBus Interface ───────────────────────────────────────────────────────
//
// EventBus decouples event producers (trips, vehicles, drivers) from consumers
// (webhook dispatcher, connectors). The production implementation is Kafka, NATS,
// or Redis Streams — swap via NewKafkaEventBus / NewNATSEventBus.

type EventHandler func(ctx context.Context, envelope EventEnvelope)

type EventBus interface {
	// Publish sends an event to the topic for the given event type.
	// Partition key is tenant_id for per-tenant ordering.
	Publish(ctx context.Context, topic string, envelope EventEnvelope) error
	// Subscribe registers a handler for all events on the given topic.
	Subscribe(topic string, handler EventHandler)
	// Close shuts down the bus cleanly.
	Close() error
}

// ─── InMemoryEventBus (dev / unit tests) ─────────────────────────────────────
//
// Implements partitioned event dispatching (Gap 2). Events with the same
// EntityID are routed to the same worker goroutine via modulo hashing,
// guaranteeing strict ordering within that partition.

const numPartitions = 8

type partitionQueue struct {
	ch chan EventEnvelope
}

type InMemoryEventBus struct {
	mu         sync.RWMutex
	handlers   map[string][]EventHandler
	partitions []partitionQueue
	logger     *zap.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewInMemoryEventBus(logger *zap.Logger) EventBus {
	ctx, cancel := context.WithCancel(context.Background())
	b := &InMemoryEventBus{
		handlers:   make(map[string][]EventHandler),
		partitions: make([]partitionQueue, numPartitions),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}

	for i := 0; i < numPartitions; i++ {
		b.partitions[i] = partitionQueue{ch: make(chan EventEnvelope, 1000)}
		b.wg.Add(1)
		go b.worker(i)
	}

	return b
}

// worker sequentially processes all events routed to this partition.
func (b *InMemoryEventBus) worker(partitionID int) {
	defer b.wg.Done()
	q := b.partitions[partitionID].ch
	for {
		select {
		case <-b.ctx.Done():
			return
		case env := <-q:
			b.mu.RLock()
			handlers := b.handlers[TopicForEventType(env.EventType)]
			b.mu.RUnlock()

			for _, h := range handlers {
				// Use context.Background() or a specific worker context to avoid tying to parent
				h(context.Background(), env)
			}
		}
	}
}

// hashEntity returns a simple sum of the EntityID string bytes modulo the number of partitions.
func hashEntity(entityID string) int {
	if entityID == "" {
		return 0 // Fallback partition for events without a precise entity
	}
	var sum int
	for i := 0; i < len(entityID); i++ {
		sum += int(entityID[i])
	}
	return sum % numPartitions
}

func (b *InMemoryEventBus) Publish(ctx context.Context, topic string, envelope EventEnvelope) error {
	b.logger.Debug("EventBus: publish",
		zap.String("topic", topic),
		zap.String("event_type", envelope.EventType),
		zap.String("tenant_id", envelope.TenantID.String()),
		zap.String("entity_id", envelope.EntityID),
		zap.String("trace_id", envelope.TraceID))

	partitionIdx := hashEntity(envelope.EntityID)
	
	select {
	case b.partitions[partitionIdx].ch <- envelope:
		return nil
	default:
		// Queue full
		b.logger.Warn("EventBus: partition queue full, blocking", zap.Int("partition", partitionIdx))
		select {
		case b.partitions[partitionIdx].ch <- envelope:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *InMemoryEventBus) Subscribe(topic string, handler EventHandler) {
	b.mu.Lock()
	b.handlers[topic] = append(b.handlers[topic], handler)
	b.mu.Unlock()
}

func (b *InMemoryEventBus) Close() error { 
	b.cancel()
	b.wg.Wait()
	return nil 
}

// ─── KafkaEventBus (production stub, wire when kafka client is imported) ──────

// KafkaEventBus is the production event bus using Apache Kafka.
// Topics partition by tenant_id for per-tenant ordering guarantees.
//
// To enable:
//  1. go get github.com/segmentio/kafka-go
//  2. Replace InMemoryEventBus with NewKafkaEventBus(brokers, logger)
type KafkaEventBus struct {
	brokers []string
	logger  *zap.Logger
	// writer  *kafka.Writer  // uncomment after importing kafka-go
}

func NewKafkaEventBus(brokers []string, logger *zap.Logger) EventBus {
	logger.Info("KafkaEventBus stub — wire segmentio/kafka-go to enable",
		zap.Strings("brokers", brokers))
	return NewInMemoryEventBus(logger) // fallback to in-memory until wired
}

// ─── Event Builder ────────────────────────────────────────────────────────────

// NewEnvelope constructs a well-formed EventEnvelope for publishing.
func NewEnvelope(eventType string, tenantID uuid.UUID, entityID string, traceID string, payload interface{}) (EventEnvelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return EventEnvelope{}, err
	}
	return EventEnvelope{
		EventID:       uuid.New(),
		EventType:     eventType,
		TenantID:      tenantID,
		EntityID:      entityID,
		Timestamp:     time.Now().UTC(),
		SchemaVersion: "1.0",
		TraceID:       traceID,
		Payload:       json.RawMessage(raw),
	}, nil
}

// TopicForEventType maps event types to their canonical topic.
func TopicForEventType(eventType string) string {
	switch {
	case len(eventType) >= 4 && eventType[:4] == "trip":
		return TopicTrip
	case len(eventType) >= 6 && eventType[:6] == "driver":
		return TopicDriver
	case len(eventType) >= 5 && eventType[:5] == "video":
		return TopicVideo
	case len(eventType) >= 7 && eventType[:7] == "invoice":
		return TopicInvoice
	case len(eventType) >= 5 && eventType[:5] == "alert":
		return TopicAlert
	case len(eventType) >= 7 && eventType[:7] == "vehicle":
		return TopicVehicle
	case len(eventType) >= 7 && eventType[:7] == "finance":
		return TopicFinance
	case len(eventType) >= 4 && eventType[:4] == "fuel":
		return TopicFuel
	default:
		return TopicAlert // fallback
	}
}
