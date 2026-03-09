package integrations

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── WebhookDispatcher ────────────────────────────────────────────────────────
//
// Dispatch workflow:
//  1. EventBus → EventRouter → WebhookDispatcher
//  2. Find all active webhook subscriptions for event_type
//  3. Sign payload with HMAC-SHA256
//  4. HTTP POST to target_url
//  5. On failure: RetryManager schedules next attempt

type WebhookDispatcher struct {
	repo    Repository
	retry   *RetryManager
	metrics *IntegrationMetrics
	client  *http.Client
	logger  *zap.Logger
}

func NewWebhookDispatcher(repo Repository, retry *RetryManager, metrics *IntegrationMetrics, logger *zap.Logger) *WebhookDispatcher {
	return &WebhookDispatcher{
		repo:    repo,
		retry:   retry,
		metrics: metrics,
		client:  &http.Client{Timeout: 15 * time.Second},
		logger:  logger,
	}
}

// Dispatch sends an event to all matching webhook subscriptions.
func (d *WebhookDispatcher) Dispatch(ctx context.Context, envelope EventEnvelope) {
	subs, err := d.repo.GetWebhooksByEventType(ctx, envelope.EventType)
	if err != nil {
		d.logger.Error("Webhook: failed to query subscriptions", zap.Error(err))
		return
	}
	
	payloadBytes, err := json.Marshal(envelope)
	if err != nil {
		d.logger.Error("Webhook: failed to marshal envelope", zap.Error(err))
		return
	}

	for _, sub := range subs {
		job := &WebhookJob{
			ID:        uuid.New(),
			WebhookID: sub.ID,
			TenantID:  envelope.TenantID,
			EventID:   envelope.EventID,
			Payload:   payloadBytes,
			Status:    EventPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := d.repo.EnqueueWebhookJob(ctx, job); err != nil {
			d.logger.Error("Webhook: failed to enqueue job", zap.Error(err))
		}
	}
}

// ─── WebhookRetryWorker ───────────────────────────────────────────────────────

type WebhookRetryWorker struct {
	repo    Repository
	logger  *zap.Logger
	client  *http.Client
}

func NewWebhookRetryWorker(repo Repository, logger *zap.Logger) *WebhookRetryWorker {
	return &WebhookRetryWorker{
		repo:   repo,
		logger: logger,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (w *WebhookRetryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("WebhookRetryWorker stopping")
			return
		case <-ticker.C:
			w.processPendingJobs(ctx)
		}
	}
}

func (w *WebhookRetryWorker) processPendingJobs(ctx context.Context) {
	job, err := w.repo.AcquireWebhookJob(ctx, "worker-1")
	if err != nil || job == nil {
		return
	}

	// Fetch subscription
	subs, err := w.repo.GetWebhooksByTenant(ctx, job.TenantID)
	if err != nil {
		return
	}
	var sub *WebhookSubscription
	for _, s := range subs {
		if s.ID == job.WebhookID {
			sub = &s
			break
		}
	}
	if sub == nil {
		// Subscription deleted, mark job discarded
		_ = w.repo.UpdateWebhookJobStatus(ctx, job.ID, EventFailed, job.RetryCount, nil)
		return
	}

	start := time.Now()
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := SignWebhookPayload(sub.Secret, ts, job.Payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.TargetURL, bytes.NewReader(job.Payload))
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Signature", "sha256="+sig)
		req.Header.Set("X-Timestamp", ts)
		req.Header.Set("X-Event-ID", job.EventID.String())
	}

	var resp *http.Response
	if err == nil {
		resp, err = w.client.Do(req)
	}

	latency := time.Since(start)
	
	delivery := &WebhookDelivery{
		ID:           uuid.New(),
		WebhookID:    sub.ID,
		JobID:        job.ID,
		EventID:      job.EventID,
		Status:       "success",
		LatencyMs:    int(latency.Milliseconds()),
	}

	success := (err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300)
	
	if resp != nil {
		delivery.ResponseCode = resp.StatusCode
		resp.Body.Close()
	}
	
	if !success {
		delivery.Status = "failed"
		if err != nil {
			delivery.ErrorMessage = err.Error()
		} else {
			delivery.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		if job.RetryCount < sub.RetryPolicy.MaxRetries {
			idx := job.RetryCount
			if idx >= len(sub.RetryPolicy.BackoffSeconds) {
				idx = len(sub.RetryPolicy.BackoffSeconds) - 1
			}
			delay := time.Duration(sub.RetryPolicy.BackoffSeconds[idx]) * time.Second
			nextAtt := time.Now().Add(delay)
			_ = w.repo.UpdateWebhookJobStatus(ctx, job.ID, EventPending, job.RetryCount+1, &nextAtt)
		} else {
			_ = w.repo.UpdateWebhookJobStatus(ctx, job.ID, EventFailed, job.RetryCount, nil)
		}
	} else {
		_ = w.repo.UpdateWebhookJobStatus(ctx, job.ID, EventDelivered, job.RetryCount, nil)
	}

	_ = w.repo.RecordWebhookDelivery(ctx, delivery)
}

// SignWebhookPayload produces the HMAC-SHA256 signature for a webhook delivery.
// Format: HMAC-SHA256(secret, "timestamp.body")
func SignWebhookPayload(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature validates an incoming webhook request at the receiving end.
// Rejects requests where the timestamp is older than 5 minutes.
func VerifyWebhookSignature(secret, signature, timestamp string, body []byte) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("webhook: invalid timestamp")
	}
	age := time.Since(time.Unix(ts, 0))
	if age > 5*time.Minute {
		return fmt.Errorf("webhook: timestamp too old (%s)", age.Round(time.Second))
	}
	expected := "sha256=" + SignWebhookPayload(secret, timestamp, body)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("webhook: signature mismatch")
	}
	return nil
}

func marshalEnvelope(env EventEnvelope) ([]byte, error) {
	return json.Marshal(env)
}
