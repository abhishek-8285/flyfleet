//go:build integration
// +build integration

package integrations

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// MockRepository is a basic mock for the Repository interface.
type MockRepository struct {
	mu              sync.Mutex
	processedEvents map[string]bool
}

func (m *MockRepository) CreateIntegration(ctx context.Context, i *Integration) error { return nil }
func (m *MockRepository) GetIntegration(ctx context.Context, id, tenantID uuid.UUID) (*Integration, error) { return nil, nil }
func (m *MockRepository) ListIntegrations(ctx context.Context, tenantID uuid.UUID) ([]Integration, error) { return nil, nil }
func (m *MockRepository) UpdateIntegrationStatus(ctx context.Context, id, tenantID uuid.UUID, status IntegrationStatus) error { return nil }
func (m *MockRepository) DeleteIntegration(ctx context.Context, id, tenantID uuid.UUID) error { return nil }
func (m *MockRepository) StoreCredential(ctx context.Context, cred *IntegrationCredential) error { return nil }
func (m *MockRepository) GetActiveCredential(ctx context.Context, integrationID uuid.UUID, credType CredentialType) (*IntegrationCredential, error) { return nil, nil }
func (m *MockRepository) GetCredentials(ctx context.Context, integrationID uuid.UUID) ([]IntegrationCredential, error) { return nil, nil }
func (m *MockRepository) DeactivateCredentials(ctx context.Context, integrationID uuid.UUID, credType CredentialType) error { return nil }
func (m *MockRepository) DeleteCredentials(ctx context.Context, integrationID uuid.UUID) error { return nil }
func (m *MockRepository) CreateWebhook(ctx context.Context, w *WebhookSubscription) error { return nil }
func (m *MockRepository) GetWebhooksByEventType(ctx context.Context, eventType string) ([]WebhookSubscription, error) { return nil, nil }
func (m *MockRepository) GetWebhooksByTenant(ctx context.Context, tenantID uuid.UUID) ([]WebhookSubscription, error) { return nil, nil }
func (m *MockRepository) DeleteWebhook(ctx context.Context, id, tenantID uuid.UUID) error { return nil }
func (m *MockRepository) EnqueueEvent(ctx context.Context, evt *IntegrationEvent) error { return nil }
func (m *MockRepository) AcquireEvent(ctx context.Context, workerID string) (*IntegrationEvent, error) { return nil, nil }
func (m *MockRepository) UpdateEventStatus(ctx context.Context, id uuid.UUID, status EventStatus, processedAt *time.Time) error { return nil }
func (m *MockRepository) MarkEventForRetry(ctx context.Context, id uuid.UUID, retryAfter time.Time) error { return nil }
func (m *MockRepository) CountPendingEvents(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockRepository) CheckAndRecordProcessedEvent(ctx context.Context, tenantID, eventID uuid.UUID, processor string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.processedEvents == nil {
		m.processedEvents = make(map[string]bool)
	}
	key := fmt.Sprintf("%s-%s", tenantID, eventID)
	if m.processedEvents[key] {
		return false, nil
	}
	m.processedEvents[key] = true
	return true, nil
}
func (m *MockRepository) EnqueueWebhookJob(ctx context.Context, job *WebhookJob) error { return nil }
func (m *MockRepository) AcquireWebhookJob(ctx context.Context, workerID string) (*WebhookJob, error) { return nil, nil }
func (m *MockRepository) UpdateWebhookJobStatus(ctx context.Context, jobID uuid.UUID, status EventStatus, retryCount int, nextAttempt *time.Time) error { return nil }
func (m *MockRepository) RecordWebhookDelivery(ctx context.Context, delivery *WebhookDelivery) error { return nil }
func (m *MockRepository) IncrementUsageQuota(ctx context.Context, integrationID, tenantID uuid.UUID, metric string, count int64) error { return nil }
func (m *MockRepository) LogAuditAction(ctx context.Context, log *IntegrationAuditLog) error { return nil }
func (m *MockRepository) SendToDLQ(ctx context.Context, entry *DLQEntry) error { return nil }
func (m *MockRepository) ListDLQ(ctx context.Context, tenantID uuid.UUID, limit int) ([]DLQEntry, error) { return nil, nil }
func (m *MockRepository) CountDLQ(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockRepository) CreateSyncJob(ctx context.Context, job *SyncJob) error { return nil }
func (m *MockRepository) AcquireSyncJob(ctx context.Context, workerID string) (*SyncJob, error) { return nil, nil }
func (m *MockRepository) UpdateSyncJobStatus(ctx context.Context, id uuid.UUID, status SyncJobStatus, errMsg string) error { return nil }
func (m *MockRepository) SetSystemFlag(ctx context.Context, key string, value string) error { return nil }
func (m *MockRepository) GetSystemFlag(ctx context.Context, key string) (string, error) { return "false", nil }


// OperationalMockRepo allows overriding specific methods for testing.
type OperationalMockRepo struct {
	MockRepository
	CountDLQFunc            func(ctx context.Context) (int64, error)
	CountPendingEventsFunc  func(ctx context.Context) (int64, error)
	GetActiveCredentialFunc func(ctx context.Context, integrationID uuid.UUID, credType CredentialType) (*IntegrationCredential, error)
	SendToDLQFunc           func(ctx context.Context, entry *DLQEntry) error
	processedEventsMap      map[string]int32
	mu                      sync.Mutex
}

func (m *OperationalMockRepo) CountDLQ(ctx context.Context) (int64, error) {
	if m.CountDLQFunc != nil {
		return m.CountDLQFunc(ctx)
	}
	return 0, nil
}

func (m *OperationalMockRepo) CountPendingEvents(ctx context.Context) (int64, error) {
	if m.CountPendingEventsFunc != nil {
		return m.CountPendingEventsFunc(ctx)
	}
	return 0, nil
}

func (m *OperationalMockRepo) GetActiveCredential(ctx context.Context, integrationID uuid.UUID, credType CredentialType) (*IntegrationCredential, error) {
	if m.GetActiveCredentialFunc != nil {
		return m.GetActiveCredentialFunc(ctx, integrationID, credType)
	}
	return nil, nil
}

func (m *OperationalMockRepo) CheckAndRecordProcessedEvent(ctx context.Context, tenantID, eventID uuid.UUID, processor string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.processedEventsMap == nil {
		m.processedEventsMap = make(map[string]int32)
	}
	key := fmt.Sprintf("%s-%s-%s", tenantID, eventID, processor)
	if m.processedEventsMap[key] > 0 {
		return false, nil
	}
	m.processedEventsMap[key] = 1
	return true, nil
}

func (m *OperationalMockRepo) SendToDLQ(ctx context.Context, entry *DLQEntry) error {
	if m.SendToDLQFunc != nil {
		return m.SendToDLQFunc(ctx, entry)
	}
	return nil
}

func (m *OperationalMockRepo) SetSystemFlag(ctx context.Context, key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.processedEventsMap == nil {
		m.processedEventsMap = make(map[string]int32)
	}
	// Overloading processedEventsMap to store flags for simplicity in this flat mock
	if value == "true" {
		m.processedEventsMap["flag_"+key] = 1
	} else {
		m.processedEventsMap["flag_"+key] = 0
	}
	return nil
}

func (m *OperationalMockRepo) GetSystemFlag(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.processedEventsMap != nil && m.processedEventsMap["flag_"+key] == 1 {
		return "true", nil
	}
	return "false", nil
}


// FaultyConnector simulates various failure modes for chaos testing.
type FaultyConnector struct {
	BaseConnector
	shouldPanic bool
	shouldHang  bool
	shouldError bool
	errToReturn error
	processed   int32
	initializeHook func()
}

func (c *FaultyConnector) Initialize(cfg ConnectorConfig) error {
	if c.initializeHook != nil {
		c.initializeHook()
	}
	atomic.AddInt32(&c.processed, 100) // indicator of initialization
	return nil
}

func (c *FaultyConnector) SendEvent(ctx context.Context, event EventEnvelope) error {
	atomic.AddInt32(&c.processed, 1)
	if c.shouldPanic {
		panic("simulated connector panic")
	}
	if c.errToReturn != nil {
		return c.errToReturn
	}
	if c.shouldError {
		return errors.New("simulated connector error")
	}
	if c.shouldHang {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Hour):
			return nil
		}
	}
	return nil
}

func (c *FaultyConnector) HealthCheck(ctx context.Context) error { return nil }
func (c *FaultyConnector) PullData(ctx context.Context, entityType string) ([]Data, error) { return nil, nil }
func (c *FaultyConnector) PushData(ctx context.Context, data []Data) error { return nil }

type mockOrderingConnector struct {
	BaseConnector
	callback func(entityID, eventType string)
}

func (c *mockOrderingConnector) SendEvent(ctx context.Context, event EventEnvelope) error {
	c.callback(event.EntityID, event.EventType)
	return nil
}
func (c *mockOrderingConnector) HealthCheck(ctx context.Context) error { return nil }
func (c *mockOrderingConnector) PullData(ctx context.Context, entityType string) ([]Data, error) { return nil, nil }
func (c *mockOrderingConnector) PushData(ctx context.Context, data []Data) error { return nil }
