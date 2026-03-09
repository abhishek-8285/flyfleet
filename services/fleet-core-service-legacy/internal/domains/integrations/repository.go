package integrations

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Repository is the full data access layer for the integrations domain.
type Repository interface {
	// Integration CRUD
	CreateIntegration(ctx context.Context, i *Integration) error
	GetIntegration(ctx context.Context, id, tenantID uuid.UUID) (*Integration, error)
	ListIntegrations(ctx context.Context, tenantID uuid.UUID) ([]Integration, error)
	UpdateIntegrationStatus(ctx context.Context, id, tenantID uuid.UUID, status IntegrationStatus) error
	DeleteIntegration(ctx context.Context, id, tenantID uuid.UUID) error

	// Credentials
	StoreCredential(ctx context.Context, cred *IntegrationCredential) error
	GetActiveCredential(ctx context.Context, integrationID uuid.UUID, credType CredentialType) (*IntegrationCredential, error)
	GetCredentials(ctx context.Context, integrationID uuid.UUID) ([]IntegrationCredential, error)
	DeactivateCredentials(ctx context.Context, integrationID uuid.UUID, credType CredentialType) error
	DeleteCredentials(ctx context.Context, integrationID uuid.UUID) error

	// Webhook subscriptions
	CreateWebhook(ctx context.Context, w *WebhookSubscription) error
	GetWebhooksByEventType(ctx context.Context, eventType string) ([]WebhookSubscription, error)
	GetWebhooksByTenant(ctx context.Context, tenantID uuid.UUID) ([]WebhookSubscription, error)
	DeleteWebhook(ctx context.Context, id, tenantID uuid.UUID) error

	// Integration events (persistent queue)
	EnqueueEvent(ctx context.Context, evt *IntegrationEvent) error
	AcquireEvent(ctx context.Context, workerID string) (*IntegrationEvent, error)
	UpdateEventStatus(ctx context.Context, id uuid.UUID, status EventStatus, processedAt *time.Time) error
	MarkEventForRetry(ctx context.Context, id uuid.UUID, retryAfter time.Time) error
	CountPendingEvents(ctx context.Context) (int64, error)

	// Event Idempotency (Gap 1)
	CheckAndRecordProcessedEvent(ctx context.Context, tenantID, eventID uuid.UUID, processor string) (bool, error)

	// Webhook Jobs & Audit (Gaps 3 & 4)
	EnqueueWebhookJob(ctx context.Context, job *WebhookJob) error
	AcquireWebhookJob(ctx context.Context, workerID string) (*WebhookJob, error)
	UpdateWebhookJobStatus(ctx context.Context, jobID uuid.UUID, status EventStatus, retryCount int, nextAttempt *time.Time) error
	RecordWebhookDelivery(ctx context.Context, delivery *WebhookDelivery) error

	// Quotas & Audit (Gaps 7 & 10)
	IncrementUsageQuota(ctx context.Context, integrationID, tenantID uuid.UUID, metric string, count int64) error
	LogAuditAction(ctx context.Context, log *IntegrationAuditLog) error

	// DLQ
	SendToDLQ(ctx context.Context, entry *DLQEntry) error
	ListDLQ(ctx context.Context, tenantID uuid.UUID, limit int) ([]DLQEntry, error)
	CountDLQ(ctx context.Context) (int64, error)

	// Distributed State (Hardening Round 2)
	SetSystemFlag(ctx context.Context, key string, value string) error
	GetSystemFlag(ctx context.Context, key string) (string, error)

	// Sync jobs
	CreateSyncJob(ctx context.Context, job *SyncJob) error
	AcquireSyncJob(ctx context.Context, workerID string) (*SyncJob, error)
	UpdateSyncJobStatus(ctx context.Context, id uuid.UUID, status SyncJobStatus, errMsg string) error
}

type repository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) Repository {
	return &repository{db: db, logger: logger}
}

// ─── Integration CRUD ─────────────────────────────────────────────────────────

func (r *repository) CreateIntegration(ctx context.Context, i *Integration) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO integrations (id,tenant_id,name,provider,integration_type,status,config_json,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())`,
		i.ID, i.TenantID, i.Name, i.Provider, string(i.IntegrationType),
		string(i.Status), i.ConfigJSON)
	return err
}

func (r *repository) GetIntegration(ctx context.Context, id, tenantID uuid.UUID) (*Integration, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id,tenant_id,name,provider,integration_type,status,config_json,created_at,updated_at
		 FROM integrations WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var i Integration
	var iType, status string
	var cfg []byte
	if err := row.Scan(&i.ID, &i.TenantID, &i.Name, &i.Provider, &iType, &status, &cfg, &i.CreatedAt, &i.UpdatedAt); err != nil {
		return nil, err
	}
	i.IntegrationType = IntegrationType(iType)
	i.Status = IntegrationStatus(status)
	i.ConfigJSON = json.RawMessage(cfg)
	return &i, nil
}

func (r *repository) ListIntegrations(ctx context.Context, tenantID uuid.UUID) ([]Integration, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,tenant_id,name,provider,integration_type,status,config_json,created_at,updated_at
		 FROM integrations WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Integration
	for rows.Next() {
		var i Integration
		var iType, status string
		var cfg []byte
		if err := rows.Scan(&i.ID, &i.TenantID, &i.Name, &i.Provider, &iType, &status, &cfg, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		i.IntegrationType = IntegrationType(iType)
		i.Status = IntegrationStatus(status)
		i.ConfigJSON = json.RawMessage(cfg)
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *repository) UpdateIntegrationStatus(ctx context.Context, id, tenantID uuid.UUID, status IntegrationStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE integrations SET status=$1,updated_at=NOW() WHERE id=$2 AND tenant_id=$3`,
		string(status), id, tenantID)
	return err
}

func (r *repository) DeleteIntegration(ctx context.Context, id, tenantID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE integrations SET status='disabled',updated_at=NOW() WHERE id=$1 AND tenant_id=$2`,
		id, tenantID)
	return err
}

// ─── Credentials ──────────────────────────────────────────────────────────────

func (r *repository) StoreCredential(ctx context.Context, c *IntegrationCredential) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO integration_credentials (id,integration_id,credential_type,encrypted_secret,key_version,credential_version,active_version,expires_at,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())`,
		c.ID, c.IntegrationID, string(c.CredentialType), c.EncryptedSecret, c.KeyVersion, c.Version, c.ActiveVersion, c.ExpiresAt)
	return err
}

func (r *repository) GetCredentials(ctx context.Context, integrationID uuid.UUID) ([]IntegrationCredential, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,integration_id,credential_type,encrypted_secret,key_version,expires_at,created_at
		 FROM integration_credentials WHERE integration_id=$1`, integrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []IntegrationCredential
	for rows.Next() {
		var c IntegrationCredential
		var ct string
		if err := rows.Scan(&c.ID, &c.IntegrationID, &ct, &c.EncryptedSecret, &c.KeyVersion, &c.ExpiresAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CredentialType = CredentialType(ct)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *repository) DeleteCredentials(ctx context.Context, integrationID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM integration_credentials WHERE integration_id=$1`, integrationID)
	return err
}

// ─── Webhooks ─────────────────────────────────────────────────────────────────

func (r *repository) CreateWebhook(ctx context.Context, w *WebhookSubscription) error {
	policy, _ := json.Marshal(w.RetryPolicy)
	_, err := r.db.Exec(ctx,
		`INSERT INTO webhook_subscriptions (id,tenant_id,event_type,target_url,secret,status,retry_policy,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())`,
		w.ID, w.TenantID, w.EventType, w.TargetURL, w.Secret, w.Status, policy)
	return err
}

func (r *repository) GetWebhooksByEventType(ctx context.Context, eventType string) ([]WebhookSubscription, error) {
	return r.queryWebhooks(ctx,
		`SELECT id,tenant_id,event_type,target_url,secret,status,retry_policy,created_at
		 FROM webhook_subscriptions WHERE event_type=$1 AND status='active'`, eventType)
}

func (r *repository) GetWebhooksByTenant(ctx context.Context, tenantID uuid.UUID) ([]WebhookSubscription, error) {
	return r.queryWebhooks(ctx,
		`SELECT id,tenant_id,event_type,target_url,secret,status,retry_policy,created_at
		 FROM webhook_subscriptions WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
}

func (r *repository) queryWebhooks(ctx context.Context, sql string, args ...interface{}) ([]WebhookSubscription, error) {
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WebhookSubscription
	for rows.Next() {
		var w WebhookSubscription
		var policyJSON []byte
		if err := rows.Scan(&w.ID, &w.TenantID, &w.EventType, &w.TargetURL, &w.Secret, &w.Status, &policyJSON, &w.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(policyJSON, &w.RetryPolicy)
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *repository) DeleteWebhook(ctx context.Context, id, tenantID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE webhook_subscriptions SET status='inactive' WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return err
}

// ─── Event Queue ──────────────────────────────────────────────────────────────

func (r *repository) EnqueueEvent(ctx context.Context, evt *IntegrationEvent) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO integration_events
		 (id,integration_id,tenant_id,event_type,schema_version,trace_id,payload,status,retry_count,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,'pending',0,NOW())`,
		evt.ID, evt.IntegrationID, evt.TenantID, evt.EventType,
		evt.SchemaVersion, evt.TraceID, evt.Payload)
	return err
}

func (r *repository) AcquireEvent(ctx context.Context, workerID string) (*IntegrationEvent, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE integration_events SET status='processing', updated_at=NOW()
		 WHERE id = (
		   SELECT id FROM integration_events
		   WHERE status='pending' AND (retry_after IS NULL OR retry_after <= NOW())
		   ORDER BY created_at ASC LIMIT 1 FOR UPDATE SKIP LOCKED
		 )
		 RETURNING id,integration_id,tenant_id,event_type,schema_version,trace_id,payload,status,retry_count,created_at`)
	var evt IntegrationEvent
	var status string
	var payload []byte
	if err := row.Scan(&evt.ID, &evt.IntegrationID, &evt.TenantID, &evt.EventType,
		&evt.SchemaVersion, &evt.TraceID, &payload, &status, &evt.RetryCount, &evt.CreatedAt); err != nil {
		return nil, err
	}
	evt.Status = EventStatus(status)
	evt.Payload = json.RawMessage(payload)
	return &evt, nil
}

func (r *repository) UpdateEventStatus(ctx context.Context, id uuid.UUID, status EventStatus, processedAt *time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE integration_events SET status=$1,processed_at=$2,updated_at=NOW() WHERE id=$3`,
		string(status), processedAt, id)
	return err
}

func (r *repository) MarkEventForRetry(ctx context.Context, id uuid.UUID, retryAfter time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE integration_events SET status='pending', retry_count=retry_count+1, retry_after=$1, updated_at=NOW()
		 WHERE id=$2`, retryAfter, id)
	return err
}

func (r *repository) CountPendingEvents(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM integration_events WHERE status='pending'`).Scan(&n)
	return n, err
}

// ─── DLQ ─────────────────────────────────────────────────────────────────────

func (r *repository) SendToDLQ(ctx context.Context, entry *DLQEntry) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO integration_dlq (id,integration_id,event_id,tenant_id,event_type,payload,error_message,retry_count,created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())`,
		entry.ID, entry.IntegrationID, entry.EventID, entry.TenantID, entry.EventType, entry.Payload, entry.ErrorMessage, entry.RetryCount)
	return err
}

func (r *repository) ListDLQ(ctx context.Context, tenantID uuid.UUID, limit int) ([]DLQEntry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,integration_id,event_id,tenant_id,event_type,payload,error_message,retry_count,created_at
		 FROM integration_dlq
		 WHERE tenant_id=$1 OR integration_id IN (SELECT id FROM integrations WHERE tenant_id=$1)
		 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DLQEntry
	for rows.Next() {
		var e DLQEntry
		if err := rows.Scan(&e.ID, &e.IntegrationID, &e.EventID, &e.TenantID, &e.EventType, &e.Payload, &e.ErrorMessage, &e.RetryCount, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *repository) CountDLQ(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM integration_dlq`).Scan(&n)
	return n, err
}

// ─── Sync Jobs ────────────────────────────────────────────────────────────────

func (r *repository) CreateSyncJob(ctx context.Context, job *SyncJob) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO integration_sync_jobs (id,integration_id,job_type,status,scheduled_at,created_at)
		 VALUES ($1,$2,$3,'scheduled',$4,NOW())`,
		job.ID, job.IntegrationID, string(job.JobType), job.ScheduledAt)
	return err
}

func (r *repository) AcquireSyncJob(ctx context.Context, workerID string) (*SyncJob, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE integration_sync_jobs SET status='running', started_at=NOW()
		 WHERE id = (
		   SELECT id FROM integration_sync_jobs
		   WHERE status='scheduled' AND scheduled_at <= NOW()
		   ORDER BY scheduled_at ASC LIMIT 1 FOR UPDATE SKIP LOCKED
		 )
		 RETURNING id,integration_id,job_type,status,scheduled_at,started_at,created_at`)
	var job SyncJob
	var jt, st string
	if err := row.Scan(&job.ID, &job.IntegrationID, &jt, &st,
		&job.ScheduledAt, &job.StartedAt, &job.CreatedAt); err != nil {
		return nil, err
	}
	job.JobType = SyncJobType(jt)
	job.Status = SyncJobStatus(st)
	return &job, nil
}

func (r *repository) UpdateSyncJobStatus(ctx context.Context, id uuid.UUID, status SyncJobStatus, errMsg string) error {
	var completedAt *time.Time
	if status == SyncDone || status == SyncJobFailed {
		t := time.Now()
		completedAt = &t
	}
	_, err := r.db.Exec(ctx,
		`UPDATE integration_sync_jobs SET status=$1,completed_at=$2,error_message=$3 WHERE id=$4`,
		string(status), completedAt, errMsg, id)
	return err
}

// ─── Extra Credential Methods (Gap 6) ─────────────────────────────────────────

func (r *repository) GetActiveCredential(ctx context.Context, integrationID uuid.UUID, credType CredentialType) (*IntegrationCredential, error) {
	var c IntegrationCredential
	err := r.db.QueryRow(ctx,
		`SELECT id, integration_id, credential_type, encrypted_secret, key_version, credential_version, active_version, expires_at, created_at
		 FROM integration_credentials 
		 WHERE integration_id = $1 AND credential_type = $2 AND active_version > 0 
		 ORDER BY created_at DESC LIMIT 1`,
		integrationID, credType).Scan(
		&c.ID, &c.IntegrationID, &c.CredentialType, &c.EncryptedSecret, &c.KeyVersion, &c.Version, &c.ActiveVersion, &c.ExpiresAt, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repository) DeactivateCredentials(ctx context.Context, integrationID uuid.UUID, credType CredentialType) error {
	_, err := r.db.Exec(ctx, `UPDATE integration_credentials SET active_version = 0 WHERE integration_id = $1 AND credential_type = $2`, integrationID, credType)
	return err
}

// ─── Event Idempotency (Gap 1) ────────────────────────────────────────────────

func (r *repository) CheckAndRecordProcessedEvent(ctx context.Context, tenantID, eventID uuid.UUID, processor string) (bool, error) {
	q := `INSERT INTO integration_processed_events (tenant_id, event_id, processor, processed_at)
		  VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING RETURNING 1`

	var val int
	err := r.db.QueryRow(ctx, q, tenantID, eventID, processor).Scan(&val)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil // Conflict occurred; event already processed by this processor
		}
		return false, err
	}
	return true, nil // successfully inserted
}

// ─── Webhook Jobs & Audit (Gaps 3 & 4) ────────────────────────────────────────

func (r *repository) EnqueueWebhookJob(ctx context.Context, job *WebhookJob) error {
	q := `INSERT INTO webhook_jobs (id, webhook_id, tenant_id, event_id, payload, status, retry_count, next_attempt_at, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, q, job.ID, job.WebhookID, job.TenantID, job.EventID, job.Payload, job.Status, job.RetryCount, job.NextAttemptAt, job.CreatedAt, job.UpdatedAt)
	return err
}

func (r *repository) AcquireWebhookJob(ctx context.Context, workerID string) (*WebhookJob, error) {
	q := `UPDATE webhook_jobs 
		  SET status = 'processing', updated_at = NOW()
		  WHERE id = (
			  SELECT id FROM webhook_jobs 
			  WHERE status = 'pending' AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
			  ORDER BY created_at ASC
			  FOR UPDATE SKIP LOCKED LIMIT 1
		  )
		  RETURNING id, webhook_id, tenant_id, event_id, payload, status, retry_count, next_attempt_at, created_at, updated_at`

	var j WebhookJob
	var payload []byte
	var status string
	err := r.db.QueryRow(ctx, q).Scan(
		&j.ID, &j.WebhookID, &j.TenantID, &j.EventID, &payload, &status, &j.RetryCount, &j.NextAttemptAt, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	j.Payload = json.RawMessage(payload)
	j.Status = EventStatus(status)
	return &j, nil
}

func (r *repository) UpdateWebhookJobStatus(ctx context.Context, jobID uuid.UUID, status EventStatus, retryCount int, nextAttempt *time.Time) error {
	q := `UPDATE webhook_jobs SET status = $1, retry_count = $2, next_attempt_at = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.Exec(ctx, q, string(status), retryCount, nextAttempt, jobID)
	return err
}

func (r *repository) RecordWebhookDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	q := `INSERT INTO webhook_deliveries (id, webhook_id, job_id, event_id, status, response_code, latency_ms, error_message, timestamp)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`
	_, err := r.db.Exec(ctx, q, delivery.ID, delivery.WebhookID, delivery.JobID, delivery.EventID, delivery.Status, delivery.ResponseCode, delivery.LatencyMs, delivery.ErrorMessage)
	return err
}

// ─── Quotas & Audit (Gaps 7 & 10) ─────────────────────────────────────────────

func (r *repository) IncrementUsageQuota(ctx context.Context, integrationID, tenantID uuid.UUID, metric string, count int64) error {
	periodStart := time.Now().Truncate(time.Hour)
	col := ""
	switch metric {
	case "events":
		col = "events_count"
	case "api":
		col = "api_calls_count"
	case "webhooks":
		col = "webhooks_count"
	default:
		return nil
	}

	q := "INSERT INTO integration_usage_counters (integration_id, tenant_id, period_start, " + col + ") " +
		"VALUES ($1, $2, $3, $4) " +
		"ON CONFLICT (integration_id, period_start) " +
		"DO UPDATE SET " + col + " = integration_usage_counters." + col + " + EXCLUDED." + col

	_, err := r.db.Exec(ctx, q, integrationID, tenantID, periodStart, count)
	return err
}

func (r *repository) LogAuditAction(ctx context.Context, log *IntegrationAuditLog) error {
	q := `INSERT INTO integration_audit_logs (id, integration_id, tenant_id, action_type, actor_id, details, created_at)
		  VALUES ($1, $2, $3, $4, $5, $6, NOW())`
	_, err := r.db.Exec(ctx, q, log.ID, log.IntegrationID, log.TenantID, log.ActionType, log.ActorID, log.Details)
	return err
}
func (r *repository) SetSystemFlag(ctx context.Context, key string, value string) error {
	query := `
		INSERT INTO integration_system_flags (flag_key, flag_value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (flag_key) DO UPDATE SET flag_value = EXCLUDED.flag_value, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, key, value)
	return err
}

func (r *repository) GetSystemFlag(ctx context.Context, key string) (string, error) {
	query := `SELECT flag_value FROM integration_system_flags WHERE flag_key = $1`
	var value string
	err := r.db.QueryRow(ctx, query, key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

