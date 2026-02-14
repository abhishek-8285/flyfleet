package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type NotificationClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type NotificationLog struct {
	ID            uuid.UUID `json:"id"`
	CorrelationID string    `json:"correlation_id"`
	UserID        uuid.UUID `json:"user_id"`
	UserEmail     string    `json:"user_email"`
	Channel       string    `json:"channel"`
	Type          string    `json:"type"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (c *NotificationClient) ListLogs(ctx context.Context, page, limit int) ([]NotificationLog, int, error) {
	url := fmt.Sprintf("%s/api/v1/logs?page=%d&limit=%d", c.BaseURL, page, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	var result struct {
		Data  []NotificationLog `json:"data"`
		Total int               `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	return result.Data, result.Total, nil
}
