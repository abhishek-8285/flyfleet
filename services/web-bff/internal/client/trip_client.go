package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TripClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewTripClient(baseURL string) *TripClient {
	return &TripClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type TripStats struct {
	ActiveTrips int `json:"active_trips"`
}

func (c *TripClient) GetStats(ctx context.Context, token string) (*TripStats, error) {
	url := fmt.Sprintf("%s/api/v1/stats/summary", c.BaseURL)
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
		return &TripStats{ActiveTrips: 0}, fmt.Errorf("trip service returned status %d", resp.StatusCode)
	}

	var result TripStats
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
