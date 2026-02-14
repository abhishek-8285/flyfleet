package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AuthClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Organization struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Domain       string    `json:"domain"`
	DefaultTheme string    `json:"default_theme"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (c *AuthClient) GetOrganization(ctx context.Context, id, token string) (*Organization, error) {
	url := fmt.Sprintf("%s/api/v1/admin/organizations/%s", c.BaseURL, id)
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
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result struct {
		Data Organization `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *AuthClient) ListOrganizations(ctx context.Context, page, limit int, token string) ([]Organization, int, error) {
	url := fmt.Sprintf("%s/api/v1/admin/organizations?page=%d&size=%d", c.BaseURL, page, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result struct {
		Data  []Organization `json:"data"`
		Meta  struct {
			Total int `json:"total"`
		} `json:"meta"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	return result.Data, result.Meta.Total, nil
}

type UserStats struct {
	TotalUsers    int `json:"total_users"`
	ActiveUsers   int `json:"active_users"`
	TotalDrivers  int `json:"total_drivers"`
	TotalOrgs     int `json:"total_orgs"`
	ActiveOrgs    int `json:"active_orgs"`
	PendingOrgs   int `json:"pending_orgs"`
	InactiveOrgs  int `json:"inactive_orgs"`
}

func (c *AuthClient) GetStats(ctx context.Context, token string) (*UserStats, error) {
	url := fmt.Sprintf("%s/api/v1/admin/users/stats", c.BaseURL)
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
		return &UserStats{TotalUsers: 0, ActiveUsers: 0, TotalDrivers: 0}, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result struct {
		Data UserStats `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *AuthClient) GetSecurityTelemetry(ctx context.Context, token string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/v1/security/telemetry", c.BaseURL)
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
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result struct {
		Data map[string]any `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

type AuditLog struct {
	ID        uuid.UUID `json:"id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	ActorID   uuid.UUID `json:"actor_id"`
	CreatedAt time.Time `json:"created_at"`
	Metadata  map[string]any `json:"metadata"`
}

func (c *AuthClient) GetAuditLogs(ctx context.Context, page, size int, token string) ([]AuditLog, error) {
	url := fmt.Sprintf("%s/api/v1/audit/logs?page=%d&size=%d", c.BaseURL, page, size)
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
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result struct {
		Data []AuditLog `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
