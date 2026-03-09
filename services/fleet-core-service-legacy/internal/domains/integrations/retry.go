package integrations

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── Retry Manager ────────────────────────────────────────────────────────────
//
// Backoff schedule (±20% jitter applied to each delay):
//   Attempt 1 → 1 min
//   Attempt 2 → 5 min
//   Attempt 3 → 15 min
//   Attempt 4 → 1 hour
//   Attempt 5 → 6 hours

var backoffSchedule = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
	1 * time.Hour,
	6 * time.Hour,
}

type RetryManager struct {
	repo   Repository
	dlq    *DLQManager
	logger *zap.Logger
}

func NewRetryManager(repo Repository, dlq *DLQManager, logger *zap.Logger) *RetryManager {
	return &RetryManager{repo: repo, dlq: dlq, logger: logger}
}

// ScheduleRetry calculates the next attempt time and updates the event record.
// On max retries exceeded, the event is forwarded to the DLQ.
func (m *RetryManager) ScheduleRetry(ctx context.Context, evt *IntegrationEvent, errMsg string) {
	if evt.RetryCount >= 5 {
		m.logger.Error("RetryManager: max retries exceeded, sending to DLQ",
			zap.String("event_id", evt.ID.String()),
			zap.Int("retries", evt.RetryCount))
		entry := &DLQEntry{
			ID:            uuid.New(),
			IntegrationID: evt.IntegrationID,
			EventID:       &evt.ID,
			ErrorMessage:  errMsg,
			RetryCount:    evt.RetryCount,
		}
		if err := m.repo.SendToDLQ(ctx, entry); err != nil {
			m.logger.Error("RetryManager: DLQ write failed", zap.Error(err))
		}
		m.dlq.NotifyNewEntry(entry)

		now := time.Now()
		m.repo.UpdateEventStatus(ctx, evt.ID, EventFailed, &now)
		return
	}

	delay := backoffWithJitter(evt.RetryCount)
	retryAt := time.Now().Add(delay)
	m.logger.Info("RetryManager: scheduling retry",
		zap.String("event_id", evt.ID.String()),
		zap.Int("attempt", evt.RetryCount+1),
		zap.Duration("delay", delay),
		zap.Time("retry_at", retryAt))
	m.repo.MarkEventForRetry(ctx, evt.ID, retryAt)
}

// backoffWithJitter applies ±20% random jitter to the base delay.
func backoffWithJitter(attempt int) time.Duration {
	if attempt >= len(backoffSchedule) {
		attempt = len(backoffSchedule) - 1
	}
	base := backoffSchedule[attempt]
	// Deterministic jitter using current nanoseconds (avoids importing math/rand)
	jitterFraction := 0.8 + (float64(time.Now().Nanosecond()%100)/100)*0.4 // 0.80–1.20
	return time.Duration(float64(base) * jitterFraction)
}

// ─── Retry Worker ─────────────────────────────────────────────────────────────
// Picks up pending+eligible events and re-dispatches them.

type RetryWorker struct {
	repo       Repository
	dispatcher *WebhookDispatcher
	bus        EventBus
	logger     *zap.Logger
	stopCh     chan struct{}
}

func NewRetryWorker(repo Repository, dispatcher *WebhookDispatcher, bus EventBus, logger *zap.Logger) *RetryWorker {
	return &RetryWorker{repo: repo, dispatcher: dispatcher, bus: bus, logger: logger, stopCh: make(chan struct{})}
}

func (w *RetryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	w.logger.Info("RetryWorker: started")
	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *RetryWorker) Stop() { close(w.stopCh) }

func (w *RetryWorker) processNext(ctx context.Context) {
	evt, err := w.repo.AcquireEvent(ctx, "retry-worker")
	if err != nil {
		return // no eligible events
	}
	w.logger.Info("RetryWorker: re-dispatching event",
		zap.String("event_id", evt.ID.String()),
		zap.Int("retry_count", evt.RetryCount))
	envelope := EventEnvelope{
		EventID:       evt.ID,
		EventType:     evt.EventType,
		TenantID:      evt.TenantID,
		Timestamp:     evt.CreatedAt,
		SchemaVersion: evt.SchemaVersion,
		TraceID:       evt.TraceID,
		Payload:       evt.Payload,
	}
	w.dispatcher.Dispatch(ctx, envelope)
}

// ─── Event Router ─────────────────────────────────────────────────────────────
//
// EventRouter subscribes to all event bus topics and fans out to:
//  1. WebhookDispatcher (HTTP push to external systems)
//  2. ConnectorRegistry (direct API push via connectors)

type EventRouter struct {
	bus        EventBus
	dlq        *DLQManager
	dispatcher *WebhookDispatcher
	registry   *ConnectorRegistry
	vault      CredentialVault
	schema     SchemaRegistry
	repo       Repository
	metrics    *IntegrationMetrics
	logger     *zap.Logger
	mu         sync.RWMutex
	pools      map[string]*connectorWorkerPool
}

type connectorWorkerPool struct {
	connector  Connector
	name       string
	partitions []chan EventEnvelope
	dlq        *DLQManager
	registry   *ConnectorRegistry // for re-init
	vault      CredentialVault    // for re-init
	logger     *zap.Logger
	refreshMu  sync.Mutex // Guard against refresh storms (Hardening Round 2)
	lastRefresh time.Time // Timestamp of last success
}

func newConnectorWorkerPool(name string, connector Connector, registry *ConnectorRegistry, vault CredentialVault, numPartitions int, dlq *DLQManager, logger *zap.Logger) *connectorWorkerPool {
	p := &connectorWorkerPool{
		name:       name,
		connector:  connector,
		registry:   registry,
		vault:      vault,
		partitions: make([]chan EventEnvelope, numPartitions),
		dlq:        dlq,
		logger:     logger,
	}
	for i := 0; i < numPartitions; i++ {
		p.partitions[i] = make(chan EventEnvelope, 100)
		go p.startWorker(i)
	}
	return p
}

func (p *connectorWorkerPool) startWorker(id int) {
	for env := range p.partitions[id] {
		p.dispatchWithRecovery(env)
	}
}

func (p *connectorWorkerPool) dispatchWithRecovery(env EventEnvelope) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("Connector worker panic recovered",
				zap.Any("panic", r),
				zap.String("event_id", env.EventID.String()))
			
			if p.dlq != nil {
				_ = p.dlq.SendToDLQ(context.Background(), env, fmt.Sprintf("panic: %v", r))
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := p.connector.SendEvent(ctx, env); err != nil {
		p.logger.Warn("Connector SendEvent failed",
			zap.String("event_id", env.EventID.String()),
			zap.Error(err))
		
		// Automatic Re-initialization on 401 (Enterprise Readiness)
		if errors.Is(err, ErrUnauthorized) && p.registry != nil {
			p.logger.Info("Connector unauthorized, attempting re-initialization", 
				zap.String("connector", p.name),
				zap.String("tenant_id", env.TenantID.String()),
				zap.String("event_id", env.EventID.String()))
			
			// Refresh Storm Protection: only one worker should refresh at a time
			p.refreshMu.Lock()
			// If a refresh happened very recently, skip and just retry
			if time.Since(p.lastRefresh) < 10*time.Second {
				p.refreshMu.Unlock()
				p.logger.Debug("Skipping re-initialization, already refreshed recently", 
					zap.String("connector", p.name),
					zap.Duration("since_last", time.Since(p.lastRefresh)))
				err = p.connector.SendEvent(ctx, env)
			} else {
				initErr := p.connector.Initialize(ConnectorConfig{Vault: p.vault})
				if initErr == nil {
					p.lastRefresh = time.Now()
					p.refreshMu.Unlock()
					p.logger.Info("Re-initialization successful, retrying event once", 
						zap.String("event_id", env.EventID.String()),
						zap.String("tenant_id", env.TenantID.String()))
					err = p.connector.SendEvent(ctx, env) // Retry once after init
				} else {
					p.refreshMu.Unlock()
					p.logger.Error("Re-initialization failed", zap.Error(initErr))
				}
			}
		}

		if err != nil && p.dlq != nil {
			_ = p.dlq.SendToDLQ(ctx, env, err.Error())
		}
	}
}

func NewEventRouter(bus EventBus, dlq *DLQManager, dispatcher *WebhookDispatcher, registry *ConnectorRegistry, vault CredentialVault, schema SchemaRegistry, repo Repository, metrics *IntegrationMetrics, logger *zap.Logger) *EventRouter {
	r := &EventRouter{
		bus:        bus,
		dlq:        dlq,
		dispatcher: dispatcher,
		registry:   registry,
		vault:      vault,
		schema:     schema,
		repo:       repo,
		metrics:    metrics,
		logger:     logger,
		pools:      make(map[string]*connectorWorkerPool),
	}
	r.initPools()
	r.subscribeAll()
	return r
}

func (r *EventRouter) initPools() {
	for name, connector := range r.registry.connectors {
		r.pools[name] = newConnectorWorkerPool(name, connector, r.registry, r.vault, 5, r.dlq, r.logger)
	}
}

func (r *EventRouter) subscribeAll() {
	topics := []string{TopicTrip, TopicDriver, TopicVideo, TopicInvoice, TopicAlert, TopicVehicle}
	for _, topic := range topics {
		t := topic // capture
		r.bus.Subscribe(t, func(ctx context.Context, env EventEnvelope) {
			r.route(ctx, env)
		})
	}
}

func (r *EventRouter) route(ctx context.Context, env EventEnvelope) {
	r.logger.Debug("EventRouter: routing event",
		zap.String("event_type", env.EventType),
		zap.String("tenant_id", env.TenantID.String()))

	// Schema Validation (Gap 8)
	if env.SchemaVersion == "" {
		r.logger.Warn("EventRouter: event rejected due to missing schema version", zap.String("event_id", env.EventID.String()))
		return
	}
	if r.schema != nil {
		if err := r.schema.ValidatePayload(env.EventType, env.SchemaVersion, []byte(env.Payload)); err != nil {
			r.logger.Warn("EventRouter: event payload schema validation failed", 
				zap.String("event_id", env.EventID.String()),
				zap.Error(err))
			_ = r.repo.IncrementUsageQuota(ctx, uuid.Nil, env.TenantID, "invalid_schema_events", 1)
			return // Discard invalid events to protect downstream connectors
		}
	}

	// Backpressure Check (Gap 9 - Distributed in Round 2)
	if r.dlq != nil && r.dlq.IsBackpressurePaused(ctx) {
		// Only allow high priority events during backpressure
		if env.EventType != TopicAlert {
			r.logger.Warn("EventRouter: event dropped due to active backpressure", zap.String("event_id", env.EventID.String()))
			_ = r.repo.IncrementUsageQuota(ctx, uuid.Nil, env.TenantID, "dropped_events", 1)
			return
		}
	}

	// Idempotency Check (Gap 1)
	if !env.IsReplay {
		inserted, err := r.repo.CheckAndRecordProcessedEvent(ctx, env.TenantID, env.EventID, "event_router")
		if err != nil {
			r.logger.Error("EventRouter: failed to check idempotency", zap.Error(err))
			return
		}
		if !inserted {
			r.logger.Warn("EventRouter: duplicate event dropped", zap.String("event_id", env.EventID.String()))
			return
		}
	} else {
		r.logger.Info("EventRouter: processing replayed event, bypassing idempotency check", zap.String("event_id", env.EventID.String()))
	}

	// 1. Webhook fan-out
	if r.dispatcher != nil {
		r.dispatcher.Dispatch(ctx, env)
	}

	// 2. Connector fan-out via worker pools
	r.dispatchToConnectors(env)
	
	// Quota tracking: Increment webhooks created
	_ = r.repo.IncrementUsageQuota(ctx, uuid.Nil, env.TenantID, "events", 1)

	// 2. Persist event for audit + replay
	evt := &IntegrationEvent{
		ID:            env.EventID,
		TenantID:      env.TenantID,
		EventType:     env.EventType,
		SchemaVersion: env.SchemaVersion,
		TraceID:       env.TraceID,
		Payload:       env.Payload,
		Status:        EventDelivered,
	}
	if err := r.repo.EnqueueEvent(ctx, evt); err != nil {
		r.logger.Warn("EventRouter: failed to persist audit event", zap.Error(err))
	}

	if r.metrics != nil {
		r.metrics.IncEvents()
	}
}

func (r *EventRouter) dispatchToConnectors(env EventEnvelope) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, pool := range r.pools {
		partition := hashEntityForPartition(env.EntityID, len(pool.partitions))
		select {
		case pool.partitions[partition] <- env:
		default:
			r.logger.Warn("Connector pool partition full, dropping event",
				zap.String("event_id", env.EventID.String()))
		}
	}
}

func hashEntityForPartition(entityID string, numPartitions int) int {
	if entityID == "" || numPartitions <= 0 {
		return 0
	}
	h := 0
	for i := 0; i < len(entityID); i++ {
		h = 31*h + int(entityID[i])
	}
	if h < 0 {
		h = -h
	}
	return h % numPartitions
}
