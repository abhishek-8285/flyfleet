package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Emit(t *testing.T) {
	var received struct {
		EventType string `json:"event_type"`
		Source    string `json:"source"`
		Subject   string `json:"subject"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/events", r.URL.Path)
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := New(srv.URL)
	err := c.Emit(context.Background(), &EmitRequest{
		EventType: "test.event",
		Source:    "unit-test",
		Subject:   "entity:1",
		Data:      map[string]any{"key": "value"},
	})
	require.NoError(t, err)
	assert.Equal(t, "test.event", received.EventType)
	assert.Equal(t, "unit-test", received.Source)
	assert.Equal(t, "entity:1", received.Subject)
}

func TestClient_Score(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/score", r.URL.Path)
		json.NewEncoder(w).Encode(ScoreResponse{
			Score:        0.85,
			ModelName:    "dispatch_ranker",
			ModelVersion: "v1",
			Variant:      "control",
			LatencyMicros: 1200,
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	resp, err := c.Score(context.Background(), &ScoreRequest{
		ModelName:  "dispatch_ranker",
		EntityType: "driver",
		EntityID:   "42",
		FeatureNames: []string{"acceptance_rate"},
	})
	require.NoError(t, err)
	assert.Equal(t, 0.85, resp.Score)
	assert.Equal(t, "control", resp.Variant)
}

func TestClient_RetryOn500(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ScoreResponse{Score: 0.5})
	}))
	defer srv.Close()

	c := New(srv.URL,
		WithRetry(5, 10*time.Millisecond, 50*time.Millisecond),
	)
	resp, err := c.Score(context.Background(), &ScoreRequest{
		ModelName:  "test",
		EntityType: "driver",
		EntityID:   "1",
	})
	require.NoError(t, err)
	assert.Equal(t, 0.5, resp.Score)
	assert.Equal(t, 3, attempts)
}

func TestClient_NoRetryOn400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := New(srv.URL,
		WithRetry(3, 10*time.Millisecond, 50*time.Millisecond),
	)
	_, err := c.Score(context.Background(), &ScoreRequest{
		ModelName:  "test",
		EntityType: "driver",
		EntityID:   "1",
	})
	require.Error(t, err)
}

func TestClient_GetFeatures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/v1/features/driver/42")
		json.NewEncoder(w).Encode(FeatureVector{
			EntityType: "driver",
			EntityID:   "42",
			Features: []FeatureValue{
				{Name: "acceptance_rate", Value: 0.75, Timestamp: time.Now()},
			},
			Count: 1,
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	fv, err := c.GetFeatures(context.Background(), "driver", "42", []string{"acceptance_rate"})
	require.NoError(t, err)
	assert.Equal(t, "driver", fv.EntityType)
	assert.Equal(t, "42", fv.EntityID)
	assert.Len(t, fv.Features, 1)
}

func TestClient_ListModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/models", r.URL.Path)
		json.NewEncoder(w).Encode(struct {
			Models []ModelInfo `json:"models"`
		}{
			Models: []ModelInfo{
				{Name: "dispatch_ranker", Version: "v1", Active: true},
				{Name: "fraud_detector", Version: "v2", Active: false},
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	models, err := c.ListModels(context.Background())
	require.NoError(t, err)
	assert.Len(t, models, 2)
	assert.True(t, models[0].Active)
	assert.False(t, models[1].Active)
}