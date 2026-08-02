package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the typed SDK for consuming mlplatform from any Go service.
// Import this in dispatch, pricing, fraud — never import internal/ directly.
type Client struct {
	baseURL    string
	httpClient *http.Client
	retry      *RetryPolicy
}

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 50,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		retry: &RetryPolicy{
			MaxAttempts: 3,
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    2 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// --- Events ---

type EmitRequest struct {
	EventType string            `json:"event_type"`
	Source    string            `json:"source"`
	Subject   string            `json:"subject"`
	Data      any               `json:"data"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (c *Client) Emit(ctx context.Context, req *EmitRequest) error {
	return c.post(ctx, "/v1/events", req, nil)
}

func (c *Client) EmitBatch(ctx context.Context, reqs []*EmitRequest) error {
	return c.post(ctx, "/v1/events/batch", reqs, nil)
}

// --- Scoring ---

type ScoreRequest struct {
	ModelName    string            `json:"model_name"`
	EntityType   string            `json:"entity_type"`
	EntityID     string            `json:"entity_id"`
	FeatureNames []string          `json:"feature_names"`
	Context      map[string]string `json:"context,omitempty"`
}

type ScoreResponse struct {
	Score         float64            `json:"score"`
	ModelName     string             `json:"model_name"`
	ModelVersion  string             `json:"model_version"`
	Experiment    string             `json:"experiment"`
	Variant       string             `json:"variant"`
	Features      map[string]float64 `json:"features"`
	LatencyMicros int64              `json:"latency_micros"`
}

func (c *Client) Score(ctx context.Context, req *ScoreRequest) (*ScoreResponse, error) {
	var resp ScoreResponse
	if err := c.post(ctx, "/v1/score", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ScoreBatch scores multiple entities concurrently.
// Used by dispatch ranking: "score these 50 drivers, return sorted."
func (c *Client) ScoreBatch(ctx context.Context, modelName, entityType string, entityIDs []string, featureNames []string) (map[string]*ScoreResponse, error) {
	type result struct {
		id   string
		resp *ScoreResponse
		err  error
	}

	sem := make(chan struct{}, 20)
	results := make([]result, len(entityIDs))

	var wg sync.WaitGroup
	for i, id := range entityIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, entityID string) {
			defer wg.Done()
			defer func() { <-sem }()
			resp, err := c.Score(ctx, &ScoreRequest{
				ModelName:    modelName,
				EntityType:   entityType,
				EntityID:     entityID,
				FeatureNames: featureNames,
			})
			results[idx] = result{id: entityID, resp: resp, err: err}
		}(i, id)
	}
	wg.Wait()

	out := make(map[string]*ScoreResponse, len(entityIDs))
	for _, r := range results {
		if r.err != nil {
			continue
		}
		out[r.id] = r.resp
	}
	return out, nil
}

// --- Features ---

type FeatureValue struct {
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

type FeatureVector struct {
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Features   []FeatureValue `json:"features"`
	Count      int            `json:"count"`
}

func (c *Client) GetFeatures(ctx context.Context, entityType, entityID string, featureNames []string) (*FeatureVector, error) {
	path := fmt.Sprintf("/v1/features/%s/%s?features=%s",
		entityType, entityID, joinComma(featureNames))

	var resp FeatureVector
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetFeaturesBatch(ctx context.Context, entityType string, entityIDs []string, featureNames []string) ([]*FeatureVector, error) {
	req := struct {
		EntityType   string   `json:"entity_type"`
		EntityIDs    []string `json:"entity_ids"`
		FeatureNames []string `json:"feature_names"`
	}{entityType, entityIDs, featureNames}

	var resp struct {
		Vectors []*FeatureVector `json:"vectors"`
	}
	if err := c.post(ctx, "/v1/features/batch", req, &resp); err != nil {
		return nil, err
	}
	return resp.Vectors, nil
}

// --- Models ---

type ModelInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Active  bool   `json:"active"`
}

func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	var resp struct {
		Models []ModelInfo `json:"models"`
	}
	if err := c.get(ctx, "/v1/models", &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}

// --- HTTP plumbing ---

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return c.retry.Do(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return &RetryableError{err}
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return &RetryableError{fmt.Errorf("server error: %d", resp.StatusCode)}
		}
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("client error %d: %s", resp.StatusCode, body)
		}

		if out != nil {
			return json.NewDecoder(resp.Body).Decode(out)
		}
		return nil
	})
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	return c.retry.Do(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return &RetryableError{err}
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return &RetryableError{fmt.Errorf("server error: %d", resp.StatusCode)}
		}
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("client error %d: %s", resp.StatusCode, body)
		}

		return json.NewDecoder(resp.Body).Decode(out)
	})
}

func joinComma(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out
}