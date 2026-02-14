package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type VehicleClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewVehicleClient(baseURL string) *VehicleClient {
	return &VehicleClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Vehicle struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	VIN       string    `json:"vin"`
	Make      string    `json:"make"`
	Model     string    `json:"model"`
	Year      int       `json:"year"`
	Plate     string    `json:"license_plate"`
	Status    string    `json:"status"`
}

type VehicleListResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

func (c *VehicleClient) ListVehicles(ctx context.Context, page, limit int) (*VehicleListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/vehicles?page=%d&limit=%d", c.BaseURL, page, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vehicle service returned status %d", resp.StatusCode)
	}

	var result VehicleListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

type VehicleStats struct {
	TotalVehicles  int `json:"total_vehicles"`
	ActiveVehicles int `json:"active_vehicles"`
	TotalDistance  float64 `json:"total_distance"`
	TotalFuel      float64 `json:"total_fuel"`
}

func (c *VehicleClient) GetStats(ctx context.Context, token string) (*VehicleStats, error) {
	url := fmt.Sprintf("%s/api/v1/vehicles/stats", c.BaseURL)
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
		// Fallback for now if endpoint doesn't exist or error
		return &VehicleStats{
			TotalVehicles: 0,
			ActiveVehicles: 0,
		}, fmt.Errorf("vehicle service returned status %d", resp.StatusCode)
	}

	var result VehicleStats
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
