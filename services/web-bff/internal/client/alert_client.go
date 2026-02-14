package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AlertClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAlertClient(baseURL string) *AlertClient {
	return &AlertClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type AlertStats struct {
	TotalAlerts int `json:"total_alerts"`
	NewAlerts   int `json:"new_alerts"`
}

func (c *AlertClient) GetStats(ctx context.Context, token string) (*AlertStats, error) {
	url := fmt.Sprintf("%s/api/v1/alerts/stats/new", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &AlertStats{TotalAlerts: 0, NewAlerts: 0}, fmt.Errorf("alert service returned status %d", resp.StatusCode)
	}

	var result AlertStats
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
