package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewNotificationClient(baseURL string, logger *zap.Logger) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

type SendNotificationRequest struct {
	Recipient string                 `json:"recipient"`
	Channel   string                 `json:"channel"`
	Subject   string                 `json:"subject"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	Priority  string                 `json:"priority,omitempty"`
}

func (c *NotificationClient) SendEmail(ctx context.Context, recipient, subject, body string, metadata map[string]interface{}) error {
	reqBody := SendNotificationRequest{
		Recipient: recipient,
		Channel:   "EMAIL", // Hardcoded for now
		Subject:   subject,
		Message:   body,
		Metadata:  metadata,
		Priority:  "STANDARD",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/internal/notifications", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.logger.Error("Failed to send notification", zap.Int("status", resp.StatusCode), zap.String("url", url))
		return fmt.Errorf("notification service returned status: %d", resp.StatusCode)
	}

	return nil
}
