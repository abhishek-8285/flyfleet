package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BillingClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewBillingClient(baseURL string) *BillingClient {
	return &BillingClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type BillingMetrics struct {
	TotalOutstanding float64 `json:"total_outstanding"`
	OverdueAmount    float64 `json:"overdue_amount"`
	MonthlyRevenue   float64 `json:"monthly_revenue"`
	RiskScore        float64 `json:"risk_score"`
}

func (c *BillingClient) GetMetrics(ctx context.Context, token string) (*BillingMetrics, error) {
	url := fmt.Sprintf("%s/api/v1/admin/billing-metrics", c.BaseURL)
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
		return &BillingMetrics{}, fmt.Errorf("billing service returned status %d", resp.StatusCode)
	}

	var result struct {
		Data BillingMetrics `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
