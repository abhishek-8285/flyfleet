package integrations

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/semaphore"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── Connector Interface (SDK) ────────────────────────────────────────────────
//
// All external system integrations implement this interface.
// External teams can build connectors by implementing this interface and
// registering them with the ConnectorRegistry.

type ConnectorConfig struct {
	Provider   string
	BaseURL    string
	AuthType   string // api_key, oauth2, basic
	Headers        map[string]string
	Timeout        time.Duration
	MaxRetries     int
	MaxConcurrency int64 // Max generic workers
	Vault          CredentialVault
}

type Data struct {
	EntityType string
	EntityID   string
	Payload    interface{}
}

type Connector interface {
	Initialize(cfg ConnectorConfig) error
	SendEvent(ctx context.Context, event EventEnvelope) error
	PullData(ctx context.Context, entityType string) ([]Data, error)
	PushData(ctx context.Context, data []Data) error
	TransformPayload(input interface{}, mapping map[string]string) (interface{}, error)
	HealthCheck(ctx context.Context) error
	Shutdown() error
}

// ─── Base Connector ───────────────────────────────────────────────────────────

// BaseConnector provides shared HTTP, retry, and auth functionality.
// Embed in every connector implementation.
type BaseConnector struct {
	cfg    ConnectorConfig
	client *http.Client
	sem    *semaphore.Weighted
	logger *zap.Logger
}

func (b *BaseConnector) Initialize(cfg ConnectorConfig) error {
	b.cfg = cfg
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	b.client = &http.Client{Timeout: timeout}
	
	concurrency := cfg.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 100 // default loose limit
	}
	b.sem = semaphore.NewWeighted(concurrency)
	
	return nil
}

func (b *BaseConnector) TransformPayload(input interface{}, mapping map[string]string) (interface{}, error) {
	// Delegate to the shared PayloadTransformer
	t := &JSONPathTransformer{}
	return t.Transform(input, mapping)
}

func (b *BaseConnector) Shutdown() error { return nil }

// ─── Connector Registry ───────────────────────────────────────────────────────

type ConnectorRegistry struct {
	connectors map[string]Connector
	logger     *zap.Logger
}

func NewConnectorRegistry(logger *zap.Logger) *ConnectorRegistry {
	return &ConnectorRegistry{connectors: make(map[string]Connector), logger: logger}
}

func (r *ConnectorRegistry) Register(provider string, c Connector) {
	r.connectors[provider] = c
	r.logger.Info("Connector registered", zap.String("provider", provider))
}

func (r *ConnectorRegistry) Get(provider string) (Connector, error) {
	c, ok := r.connectors[provider]
	if !ok {
		return nil, fmt.Errorf("connector %q not registered", provider)
	}
	return c, nil
}

func (r *ConnectorRegistry) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)
	for name, c := range r.connectors {
		results[name] = c.HealthCheck(ctx)
	}
	return results
}

// ─── Connector Health Monitor ─────────────────────────────────────────────────

type ConnectorHealthMonitor struct {
	registry *ConnectorRegistry
	logger   *zap.Logger
}

func NewConnectorHealthMonitor(reg *ConnectorRegistry, logger *zap.Logger) *ConnectorHealthMonitor {
	return &ConnectorHealthMonitor{registry: reg, logger: logger}
}

func (m *ConnectorHealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("ConnectorHealthMonitor stopping")
			return
		case <-ticker.C:
			res := m.registry.HealthCheckAll(ctx)
			for p, err := range res {
				if err != nil {
					m.logger.Warn("Connector health check failed", zap.String("provider", p), zap.Error(err))
				}
			}
		}
	}
}

// ─── SAP Connector ────────────────────────────────────────────────────────────

type SAPConnector struct{ BaseConnector }

func NewSAPConnector(logger *zap.Logger) Connector {
	return &SAPConnector{BaseConnector{logger: logger}}
}

func (c *SAPConnector) SendEvent(ctx context.Context, event EventEnvelope) error {
	c.logger.Info("SAP: SendEvent (stub)", zap.String("event_type", event.EventType),
		zap.String("tenant_id", event.TenantID.String()))
	// Production: POST to SAP REST API or RFC connector
	return nil
}

func (c *SAPConnector) PullData(ctx context.Context, entityType string) ([]Data, error) {
	c.logger.Info("SAP: PullData (stub)", zap.String("entity_type", entityType))
	return nil, nil
}

func (c *SAPConnector) PushData(ctx context.Context, data []Data) error {
	c.logger.Info("SAP: PushData (stub)", zap.Int("records", len(data)))
	return nil
}

func (c *SAPConnector) HealthCheck(ctx context.Context) error {
	c.logger.Info("SAP: HealthCheck (stub)")
	return nil // Production: GET SAP ping endpoint
}

// ─── Salesforce Connector ─────────────────────────────────────────────────────

type SalesforceConnector struct{ BaseConnector }

func NewSalesforceConnector(logger *zap.Logger) Connector {
	return &SalesforceConnector{BaseConnector{logger: logger}}
}

func (c *SalesforceConnector) SendEvent(ctx context.Context, event EventEnvelope) error {
	c.logger.Info("Salesforce: SendEvent (stub)", zap.String("event_type", event.EventType))
	// Production: Platform Events API or Streaming API
	return nil
}

func (c *SalesforceConnector) PullData(ctx context.Context, entityType string) ([]Data, error) {
	c.logger.Info("Salesforce: PullData (stub)", zap.String("entity_type", entityType))
	return nil, nil
}

func (c *SalesforceConnector) PushData(ctx context.Context, data []Data) error {
	c.logger.Info("Salesforce: PushData (stub)", zap.Int("records", len(data)))
	return nil
}

func (c *SalesforceConnector) HealthCheck(ctx context.Context) error { return nil }

// ─── Stripe Connector ─────────────────────────────────────────────────────────

type StripeConnector struct{ BaseConnector }

func NewStripeConnector(logger *zap.Logger) Connector {
	return &StripeConnector{BaseConnector{logger: logger}}
}

func (c *StripeConnector) SendEvent(ctx context.Context, event EventEnvelope) error {
	c.logger.Info("Stripe: SendEvent (stub)", zap.String("event_type", event.EventType))
	// Production: Stripe webhooks API / charges / invoices
	return nil
}

func (c *StripeConnector) PullData(ctx context.Context, entityType string) ([]Data, error) {
	return nil, nil
}

func (c *StripeConnector) PushData(ctx context.Context, data []Data) error { return nil }
func (c *StripeConnector) HealthCheck(ctx context.Context) error            { return nil }

// ─── Insurance Connector ──────────────────────────────────────────────────────

type InsuranceConnector struct{ BaseConnector }

func NewInsuranceConnector(logger *zap.Logger) Connector {
	return &InsuranceConnector{BaseConnector{logger: logger}}
}

func (c *InsuranceConnector) SendEvent(ctx context.Context, event EventEnvelope) error {
	c.logger.Info("Insurance: SendEvent (stub)", zap.String("event_type", event.EventType))
	// Production: telematics data push, incident report, usage-based premium
	return nil
}

func (c *InsuranceConnector) PullData(ctx context.Context, entityType string) ([]Data, error) {
	return nil, nil
}

func (c *InsuranceConnector) PushData(ctx context.Context, data []Data) error { return nil }
func (c *InsuranceConnector) HealthCheck(ctx context.Context) error            { return nil }

// ─── Government Compliance Connector ─────────────────────────────────────────

type ComplianceConnector struct{ BaseConnector }

func NewComplianceConnector(logger *zap.Logger) Connector {
	return &ComplianceConnector{BaseConnector{logger: logger}}
}

func (c *ComplianceConnector) SendEvent(ctx context.Context, event EventEnvelope) error {
	c.logger.Info("Compliance: SendEvent (stub)", zap.String("event_type", event.EventType))
	// Production: VAHAN, SARATHI, RTO API or MCA endpoints
	return nil
}

func (c *ComplianceConnector) PullData(ctx context.Context, entityType string) ([]Data, error) {
	return nil, nil
}

func (c *ComplianceConnector) PushData(ctx context.Context, data []Data) error { return nil }
func (c *ComplianceConnector) HealthCheck(ctx context.Context) error            { return nil }

// ConnectorHealthStatus is returned by the health endpoint
type ConnectorHealthStatus struct {
	Provider  string    `json:"provider"`
	Healthy   bool      `json:"healthy"`
	Error     string    `json:"error,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

func BuildDefaultRegistry(logger *zap.Logger) *ConnectorRegistry {
	reg := NewConnectorRegistry(logger)
	cfg := ConnectorConfig{Timeout: 30 * time.Second, MaxRetries: 3}

	connectors := map[string]Connector{
		"sap":        NewSAPConnector(logger),
		"salesforce": NewSalesforceConnector(logger),
		"stripe":     NewStripeConnector(logger),
		"insurance":  NewInsuranceConnector(logger),
		"compliance": NewComplianceConnector(logger),
	}
	for name, c := range connectors {
		c.Initialize(cfg)
		reg.Register(name, c)
	}

	_ = uuid.Nil // imported for potential future use
	return reg
}
