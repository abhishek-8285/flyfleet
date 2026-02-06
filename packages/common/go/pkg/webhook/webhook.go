package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type WebhookClient struct {
	httpClient *http.Client
	logger     *zap.Logger
}

type WebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	WebhookID string                 `json:"webhook_id"`
}

type WebhookDelivery struct {
	URL       string `json:"url"`
	Secret    string `json:"secret"`
	MaxRetries int   `json:"max_retries"`
	Timeout   time.Duration `json:"timeout"`
}

func NewWebhookClient(logger *zap.Logger) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (wc *WebhookClient) Deliver(ctx context.Context, delivery WebhookDelivery, payload WebhookPayload) error {
	payload.Timestamp = time.Now()
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	signature := wc.generateSignature(delivery.Secret, jsonPayload)
	
	req, err := http.NewRequestWithContext(ctx, "POST", delivery.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Timestamp", payload.Timestamp.Format(time.RFC3339))

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		wc.logger.Error("Webhook delivery failed",
			zap.String("url", delivery.URL),
			zap.String("event", payload.Event),
			zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		wc.logger.Info("Webhook delivered successfully",
			zap.String("url", delivery.URL),
			zap.String("event", payload.Event),
			zap.Int("status_code", resp.StatusCode))
		return nil
	}

	wc.logger.Warn("Webhook delivery failed with status",
		zap.String("url", delivery.URL),
		zap.String("event", payload.Event),
		zap.Int("status_code", resp.StatusCode))

	return fmt.Errorf("webhook delivery failed with status %d", resp.StatusCode)
}

func (wc *WebhookClient) generateSignature(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func (wc *WebhookClient) VerifySignature(secret, signature string, payload []byte) bool {
	expectedSignature := wc.generateSignature(secret, payload)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (wc *WebhookClient) RetryDeliver(ctx context.Context, delivery WebhookDelivery, payload WebhookPayload) error {
	var lastErr error
	
	for attempt := 1; attempt <= delivery.MaxRetries; attempt++ {
		err := wc.Deliver(ctx, delivery, payload)
		if err == nil {
			return nil
		}
		
		lastErr = err
		wc.logger.Warn("Webhook delivery attempt failed",
			zap.String("url", delivery.URL),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", delivery.MaxRetries),
			zap.Error(err))
		
		if attempt < delivery.MaxRetries {
			// Exponential backoff
			sleepDuration := time.Duration(attempt*attempt) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleepDuration):
				// Continue to next attempt
			}
		}
	}
	
	wc.logger.Error("All webhook delivery attempts failed",
		zap.String("url", delivery.URL),
		zap.Int("attempts", delivery.MaxRetries),
		zap.Error(lastErr))
	
	return fmt.Errorf("all %d delivery attempts failed: %w", delivery.MaxRetries, lastErr)
}