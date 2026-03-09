package integrations

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ─── Integration Status State Machine ─────────────────────────────────────────

type IntegrationStatus string

const (
	StatusCreated    IntegrationStatus = "created"
	StatusConfigured IntegrationStatus = "configured"
	StatusActive     IntegrationStatus = "active"
	StatusPaused     IntegrationStatus = "paused"
	StatusFailed     IntegrationStatus = "failed"
	StatusDisabled   IntegrationStatus = "disabled"
)

// ValidTransition enforces the lifecycle state machine.
//
//	CREATED → CONFIGURED → ACTIVE ⇄ PAUSED
//	ACTIVE/PAUSED/FAILED → DISABLED
func ValidTransition(from, to IntegrationStatus) bool {
	allowed := map[IntegrationStatus][]IntegrationStatus{
		StatusCreated:    {StatusConfigured, StatusDisabled},
		StatusConfigured: {StatusActive, StatusDisabled},
		StatusActive:     {StatusPaused, StatusFailed, StatusDisabled},
		StatusPaused:     {StatusActive, StatusDisabled},
		StatusFailed:     {StatusActive, StatusDisabled},
		StatusDisabled:   {},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}

// ─── Integration Types ────────────────────────────────────────────────────────

type IntegrationType string

const (
	TypeERP        IntegrationType = "erp"
	TypeInsurance  IntegrationType = "insurance"
	TypeTelematics IntegrationType = "telematics"
	TypePayment    IntegrationType = "payment"
	TypeCompliance IntegrationType = "compliance"
	TypeAnalytics  IntegrationType = "analytics"
	TypePartner    IntegrationType = "partner"
)

// ─── Core Domain Models ───────────────────────────────────────────────────────

type Integration struct {
	ID              uuid.UUID         `json:"id"`
	TenantID        uuid.UUID         `json:"tenant_id"`
	Name            string            `json:"name"`
	Provider        string            `json:"provider"` // sap, salesforce, stripe, …
	IntegrationType IntegrationType   `json:"integration_type"`
	Status          IntegrationStatus `json:"status"`
	ConfigJSON      json.RawMessage   `json:"config,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type IntegrationConfig struct {
	BaseURL         string           `json:"base_url,omitempty"`
	AuthType        string           `json:"auth_type"` // api_key, oauth2, basic
	RateLimitConfig *RateLimitConfig `json:"rate_limit,omitempty"`
	RetryPolicy     *RetryPolicy     `json:"retry_policy,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
}

// ─── Credentials ──────────────────────────────────────────────────────────────

type CredentialType string

const (
	CredAPIKey CredentialType = "api_key"
	CredOAuth  CredentialType = "oauth_token"
	CredBasic  CredentialType = "basic_auth"
	CredMTLS   CredentialType = "mtls_cert"
)

type IntegrationCredential struct {
	ID              uuid.UUID      `json:"id"`
	IntegrationID   uuid.UUID      `json:"integration_id"`
	CredentialType  CredentialType `json:"credential_type"`
	EncryptedSecret []byte         `json:"-"` // never expose raw bytes via JSON
	KeyVersion      string         `json:"key_version"`
	Version         int            `json:"credential_version"`
	ActiveVersion   int            `json:"active_version"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

// ─── Webhook Subscription ────────────────────────────────────────────────────

type WebhookSubscription struct {
	ID          uuid.UUID   `json:"id"`
	TenantID    uuid.UUID   `json:"tenant_id"`
	EventType   string      `json:"event_type"`
	TargetURL   string      `json:"target_url"`
	Secret      string      `json:"secret"`
	Status      string      `json:"status"`
	RetryPolicy RetryPolicy `json:"retry_policy"`
	CreatedAt   time.Time   `json:"created_at"`
}

type WebhookJob struct {
	ID            uuid.UUID       `json:"id"`
	WebhookID     uuid.UUID       `json:"webhook_id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	EventID       uuid.UUID       `json:"event_id"`
	Payload       json.RawMessage `json:"payload"`
	Status        EventStatus     `json:"status"`
	RetryCount    int             `json:"retry_count"`
	NextAttemptAt *time.Time      `json:"next_attempt_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type WebhookDelivery struct {
	ID           uuid.UUID  `json:"id"`
	WebhookID    uuid.UUID  `json:"webhook_id"`
	JobID        uuid.UUID  `json:"job_id"`
	EventID      uuid.UUID  `json:"event_id"`
	Status       string     `json:"status"`
	ResponseCode int        `json:"response_code,omitempty"`
	LatencyMs    int        `json:"latency_ms,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
}

type RetryPolicy struct {
	MaxRetries     int   `json:"max_retries"`
	BackoffSeconds []int `json:"backoff_seconds"`
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:     5,
		BackoffSeconds: []int{60, 300, 900, 3600, 21600},
	}
}

type RateLimitConfig struct {
	Capacity   int     `json:"capacity"`    // token bucket size
	RefillRate float64 `json:"refill_rate"` // tokens/sec
}

// ─── Event Envelope  (canonical schema for all event bus messages) ────────────

type EventEnvelope struct {
	EventID       uuid.UUID       `json:"event_id"`
	EventType     string          `json:"event_type"` // e.g. "trip.completed"
	TenantID      uuid.UUID       `json:"tenant_id"`
	EntityID      string          `json:"entity_id,omitempty"` // Partition key
	Timestamp     time.Time       `json:"timestamp"`
	SchemaVersion string          `json:"schema_version"`
	TraceID       string          `json:"trace_id,omitempty"`
	Payload       json.RawMessage `json:"payload"`
	IsReplay      bool            `json:"is_replay,omitempty"`
}

// Topic constants — partition key within each topic is tenant_id.
const (
	TopicTrip    = "events.trip.v1"
	TopicDriver  = "events.driver.v1"
	TopicVideo   = "events.video.v1"
	TopicInvoice = "events.invoice.v1"
	TopicAlert   = "events.alert.v1"
	TopicVehicle = "events.vehicle.v1"
	TopicFinance = "events.finance.v1"
	TopicFuel    = "events.fuel.v1"
)

// ─── Integration Event (persisted queue) ─────────────────────────────────────

type EventStatus string

const (
	EventPending    EventStatus = "pending"
	EventProcessing EventStatus = "processing"
	EventDelivered  EventStatus = "delivered"
	EventFailed     EventStatus = "failed"
)

type IntegrationEvent struct {
	ID            uuid.UUID       `json:"id"`
	IntegrationID *uuid.UUID      `json:"integration_id,omitempty"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	EventType     string          `json:"event_type"`
	SchemaVersion string          `json:"schema_version"`
	TraceID       string          `json:"trace_id,omitempty"`
	Payload       json.RawMessage `json:"payload"`
	Status        EventStatus     `json:"status"`
	RetryCount    int             `json:"retry_count"`
	RetryAfter    *time.Time      `json:"retry_after,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty"`
}

var ErrUnauthorized = errors.New("unauthorized: token expired or invalid")

type ProcessedEvent struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	EventID     uuid.UUID `json:"event_id"`
	Processor   string    `json:"processor"`
	ProcessedAt time.Time `json:"processed_at"`
}

// ─── Usage Quotas & Audit ───────────────────────────────────────────────────

type IntegrationUsage struct {
	IntegrationID uuid.UUID `json:"integration_id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	PeriodStart   time.Time `json:"period_start"`
	EventsCount   int64     `json:"events_count"`
	APICallsCount int64     `json:"api_calls_count"`
	WebhooksCount int64     `json:"webhooks_count"`
}

type IntegrationAuditLog struct {
	ID            uuid.UUID       `json:"id"`
	IntegrationID *uuid.UUID      `json:"integration_id,omitempty"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	ActionType    string          `json:"action_type"`
	ActorID       *uuid.UUID      `json:"actor_id,omitempty"`
	Details       json.RawMessage `json:"details"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ─── DLQ ─────────────────────────────────────────────────────────────────────

type DLQEntry struct {
	ID            uuid.UUID       `json:"id"`
	IntegrationID *uuid.UUID      `json:"integration_id,omitempty"`
	EventID       *uuid.UUID      `json:"event_id,omitempty"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	ErrorMessage  string          `json:"error_message"`
	RetryCount    int             `json:"retry_count"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ─── Sync Job ─────────────────────────────────────────────────────────────────

type SyncJobType   string
type SyncJobStatus string

const (
	SyncJobERPDaily         SyncJobType = "erp_daily"
	SyncJobVehicleHourly    SyncJobType = "vehicle_hourly"
	SyncJobComplianceWeekly SyncJobType = "compliance_weekly"
)

const (
	SyncScheduled SyncJobStatus = "scheduled"
	SyncRunning   SyncJobStatus = "running"
	SyncDone      SyncJobStatus = "done"
	SyncJobFailed SyncJobStatus = "failed"
)

type SyncJob struct {
	ID            uuid.UUID     `json:"id"`
	IntegrationID uuid.UUID     `json:"integration_id"`
	JobType       SyncJobType   `json:"job_type"`
	Status        SyncJobStatus `json:"status"`
	ScheduledAt   time.Time     `json:"scheduled_at"`
	StartedAt     *time.Time    `json:"started_at,omitempty"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// ─── API Request / Response Types ────────────────────────────────────────────

type CreateIntegrationRequest struct {
	Name            string          `json:"name"`
	Provider        string          `json:"provider"`
	IntegrationType IntegrationType `json:"integration_type"`
	Config          json.RawMessage `json:"config,omitempty"`
}

type UpdateStatusRequest struct {
	Status IntegrationStatus `json:"status"`
}

type CreateWebhookRequest struct {
	EventType   string       `json:"event_type"`
	TargetURL   string       `json:"target_url"`
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`
}

type AddCredentialRequest struct {
	CredentialType CredentialType `json:"credential_type"`
	Secret         string         `json:"secret"` // plaintext, encrypted at service layer
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
}

type ReplayEventRequest struct {
	EventID uuid.UUID `json:"event_id"`
}
