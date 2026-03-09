package integrations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── Service Interface ────────────────────────────────────────────────────────

type Service interface {
	// Integration lifecycle
	CreateIntegration(ctx context.Context, tenantID uuid.UUID, req CreateIntegrationRequest) (*Integration, error)
	GetIntegration(ctx context.Context, id, tenantID uuid.UUID) (*Integration, error)
	ListIntegrations(ctx context.Context, tenantID uuid.UUID) ([]Integration, error)
	EnableIntegration(ctx context.Context, id, tenantID uuid.UUID) error
	DisableIntegration(ctx context.Context, id, tenantID uuid.UUID) error
	DeleteIntegration(ctx context.Context, id, tenantID uuid.UUID) error

	// Credentials
	AddCredential(ctx context.Context, integrationID uuid.UUID, req AddCredentialRequest) error
	RotateCredential(ctx context.Context, integrationID uuid.UUID, credType CredentialType, newSecret string) error

	// Webhooks
	CreateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, req CreateWebhookRequest) (*WebhookSubscription, error)
	ListWebhookSubscriptions(ctx context.Context, tenantID uuid.UUID) ([]WebhookSubscription, error)
	DeleteWebhookSubscription(ctx context.Context, id, tenantID uuid.UUID) error

	// Events
	PublishEvent(ctx context.Context, tenantID uuid.UUID, entityID, eventType string, traceID string, payload interface{}) error
	ReplayEvent(ctx context.Context, eventID uuid.UUID) error
	ListDLQ(ctx context.Context, tenantID uuid.UUID) ([]DLQEntry, error)

	// Observability
	ConnectorHealthStatus(ctx context.Context) []ConnectorHealthStatus
}

// ─── Service Implementation ───────────────────────────────────────────────────

type service struct {
	repo       Repository
	vault      CredentialVault
	bus        EventBus
	dispatcher *WebhookDispatcher
	retry      *RetryManager
	registry   *ConnectorRegistry
	cbRegistry *CircuitBreakerRegistry
	rlRegistry *RateLimiterRegistry
	metrics    *IntegrationMetrics
	logger     *zap.Logger
}

// NewService wires all subsystems together.
func NewService(repo Repository, vault CredentialVault, bus EventBus, dispatcher *WebhookDispatcher,
	retry *RetryManager, registry *ConnectorRegistry, cbRegistry *CircuitBreakerRegistry,
	rlRegistry *RateLimiterRegistry, metrics *IntegrationMetrics, logger *zap.Logger) Service {
	return &service{
		repo:       repo,
		vault:      vault,
		bus:        bus,
		dispatcher: dispatcher,
		retry:      retry,
		registry:   registry,
		cbRegistry: cbRegistry,
		rlRegistry: rlRegistry,
		metrics:    metrics,
		logger:     logger,
	}
}

// ─── Integration Lifecycle ────────────────────────────────────────────────────

func (s *service) CreateIntegration(ctx context.Context, tenantID uuid.UUID, req CreateIntegrationRequest) (*Integration, error) {
	i := &Integration{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Name:            req.Name,
		Provider:        req.Provider,
		IntegrationType: req.IntegrationType,
		Status:          StatusCreated,
		ConfigJSON:      req.Config,
	}
	if err := s.repo.CreateIntegration(ctx, i); err != nil {
		return nil, fmt.Errorf("create integration: %w", err)
	}
	s.logger.Info("Integration created",
		zap.String("id", i.ID.String()),
		zap.String("provider", i.Provider))

	_ = s.repo.LogAuditAction(ctx, &IntegrationAuditLog{
		ID:            uuid.New(),
		IntegrationID: &i.ID,
		TenantID:      tenantID,
		ActionType:    "integration.created",
		Details:       []byte(fmt.Sprintf(`{"provider":"%s"}`, req.Provider)),
	})
	
	return i, nil
}

func (s *service) GetIntegration(ctx context.Context, id, tenantID uuid.UUID) (*Integration, error) {
	return s.repo.GetIntegration(ctx, id, tenantID)
}

func (s *service) ListIntegrations(ctx context.Context, tenantID uuid.UUID) ([]Integration, error) {
	return s.repo.ListIntegrations(ctx, tenantID)
}

func (s *service) EnableIntegration(ctx context.Context, id, tenantID uuid.UUID) error {
	i, err := s.repo.GetIntegration(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if !ValidTransition(i.Status, StatusActive) {
		return fmt.Errorf("invalid transition: %s → active", i.Status)
	}
	err = s.repo.UpdateIntegrationStatus(ctx, id, tenantID, StatusActive)
	if err == nil {
		_ = s.repo.LogAuditAction(ctx, &IntegrationAuditLog{
			ID:            uuid.New(),
			IntegrationID: &id,
			TenantID:      tenantID,
			ActionType:    "integration.enabled",
		})
	}
	return err
}

func (s *service) DisableIntegration(ctx context.Context, id, tenantID uuid.UUID) error {
	i, err := s.repo.GetIntegration(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if !ValidTransition(i.Status, StatusDisabled) {
		return fmt.Errorf("cannot disable: status is %s", i.Status)
	}
	err = s.repo.UpdateIntegrationStatus(ctx, id, tenantID, StatusDisabled)
	if err == nil {
		_ = s.repo.LogAuditAction(ctx, &IntegrationAuditLog{
			ID:            uuid.New(),
			IntegrationID: &id,
			TenantID:      tenantID,
			ActionType:    "integration.disabled",
		})
	}
	return err
}

func (s *service) DeleteIntegration(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteIntegration(ctx, id, tenantID)
}

// ─── Credentials ──────────────────────────────────────────────────────────────

func (s *service) AddCredential(ctx context.Context, integrationID uuid.UUID, req AddCredentialRequest) error {
	encrypted, err := s.vault.Encrypt(req.Secret)
	if err != nil {
		return fmt.Errorf("credential encryption failed: %w", err)
	}
	cred := &IntegrationCredential{
		ID:              uuid.New(),
		IntegrationID:   integrationID,
		CredentialType:  req.CredentialType,
		EncryptedSecret: encrypted,
		KeyVersion:      "v1",
		Version:         1,
		ActiveVersion:   1,
		ExpiresAt:       req.ExpiresAt,
	}
	err = s.repo.StoreCredential(ctx, cred)
	if err == nil {
		_ = s.repo.LogAuditAction(ctx, &IntegrationAuditLog{
			ID:            uuid.New(),
			IntegrationID: &integrationID,
			TenantID:      uuid.Nil, // assuming admin for now
			ActionType:    "credential.added",
		})
	}
	return err
}

func (s *service) RotateCredential(ctx context.Context, integrationID uuid.UUID, credType CredentialType, newSecret string) error {
	activeCred, err := s.repo.GetActiveCredential(ctx, integrationID, credType)
	if err != nil {
		return fmt.Errorf("failed to get active credential to rotate: %w", err)
	}
	
	newVersion := 1
	if activeCred != nil {
		newVersion = activeCred.Version + 1
		// Deactivate the old credential to enforce singular active credential per type
		if err := s.repo.DeactivateCredentials(ctx, integrationID, credType); err != nil {
			return fmt.Errorf("failed to deactivate old credential: %w", err)
		}
	}

	encrypted, err := s.vault.Encrypt(newSecret)
	if err != nil {
		return fmt.Errorf("credential encryption failed logic: %w", err)
	}

	cred := &IntegrationCredential{
		ID:              uuid.New(),
		IntegrationID:   integrationID,
		CredentialType:  credType,
		EncryptedSecret: encrypted,
		KeyVersion:      "v1",
		Version:         newVersion,
		ActiveVersion:   1,
	}

	err = s.repo.StoreCredential(ctx, cred)
	if err == nil {
		_ = s.repo.LogAuditAction(ctx, &IntegrationAuditLog{
			ID:            uuid.New(),
			IntegrationID: &integrationID,
			TenantID:      uuid.Nil,
			ActionType:    "credential.rotated",
			Details:       []byte(fmt.Sprintf(`{"version":%d}`, newVersion)),
		})
	}
	return err
}

// ─── Webhooks ─────────────────────────────────────────────────────────────────

func (s *service) CreateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, req CreateWebhookRequest) (*WebhookSubscription, error) {
	policy := DefaultRetryPolicy()
	if req.RetryPolicy != nil {
		policy = *req.RetryPolicy
	}
	w := &WebhookSubscription{
		ID:          uuid.New(),
		TenantID:    tenantID,
		EventType:   req.EventType,
		TargetURL:   req.TargetURL,
		Secret:      "whsec_" + uuid.New().String(), // secure random per spec
		Status:      "active",
		RetryPolicy: policy,
	}
	if err := s.repo.CreateWebhook(ctx, w); err != nil {
		return nil, err
	}

	_ = s.repo.LogAuditAction(ctx, &IntegrationAuditLog{
		ID:         uuid.New(),
		TenantID:   tenantID,
		ActionType: "webhook.created",
	})
	
	return w, nil
}

func (s *service) ListWebhookSubscriptions(ctx context.Context, tenantID uuid.UUID) ([]WebhookSubscription, error) {
	return s.repo.GetWebhooksByTenant(ctx, tenantID)
}

func (s *service) DeleteWebhookSubscription(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteWebhook(ctx, id, tenantID)
}

// ─── Event Publishing ─────────────────────────────────────────────────────────

func (s *service) PublishEvent(ctx context.Context, tenantID uuid.UUID, entityID, eventType, traceID string, payload interface{}) error {
	envelope, err := NewEnvelope(eventType, tenantID, entityID, traceID, payload)
	if err != nil {
		return fmt.Errorf("publish event: marshal payload: %w", err)
	}
	topic := TopicForEventType(eventType)

	if s.metrics != nil {
		s.metrics.IncEventsLabeled(eventType)
	}

	if err := s.bus.Publish(ctx, topic, envelope); err != nil {
		s.logger.Error("PublishEvent: bus error", zap.Error(err))
		return err
	}
	return nil
}

func (s *service) ReplayEvent(ctx context.Context, eventID uuid.UUID) error {
	s.logger.Info("Service: replay event", zap.String("event_id", eventID.String()))
	// Reset status to pending so RetryWorker picks it up
	err := s.repo.UpdateEventStatus(ctx, eventID, EventPending, nil)
	if err == nil {
		_ = s.repo.LogAuditAction(ctx, &IntegrationAuditLog{
			ID:         uuid.New(),
			TenantID:   uuid.Nil, // recorded at system/admin level
			ActionType: "event.replayed",
			Details:    []byte(fmt.Sprintf(`{"event_id":"%s"}`, eventID.String())),
		})
	}
	return err
}

func (s *service) ListDLQ(ctx context.Context, tenantID uuid.UUID) ([]DLQEntry, error) {
	return s.repo.ListDLQ(ctx, tenantID, 100)
}

// ─── Observability ────────────────────────────────────────────────────────────

func (s *service) ConnectorHealthStatus(ctx context.Context) []ConnectorHealthStatus {
	results := s.registry.HealthCheckAll(ctx)
	out := make([]ConnectorHealthStatus, 0, len(results))
	for provider, err := range results {
		status := ConnectorHealthStatus{
			Provider:  provider,
			Healthy:   err == nil,
			CheckedAt: time.Now(),
		}
		if err != nil {
			status.Error = err.Error()
		}
		out = append(out, status)
	}
	return out
}
