package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// This shows how a dispatch service consumes the ML platform.

type ScoreRequest struct {
	ModelName    string   `json:"model_name"`
	EntityType   string   `json:"entity_type"`
	EntityID     string   `json:"entity_id"`
	FeatureNames []string `json:"feature_names"`
}

type ScoreResponse struct {
	Score   float64 `json:"score"`
	Variant string  `json:"variant"`
	Latency int64   `json:"latency_micros"`
}

func main() {
	ctx := context.Background()

	// 1. Emit an event (driver accepted a ride)
	emitEvent(ctx, "driver.accepted", "dispatch-service", "driver:42", map[string]any{
		"driver_id": "42",
		"action":    "accepted",
		"ride_id":   "ride-9981",
		"zone":      "downtown",
	})

	// 2. Score a driver for dispatch ranking
	score, err := scoreDriver(ctx, "42")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Driver 42 score: %.4f (variant=%s, latency=%dµs)\n",
		score.Score, score.Variant, score.Latency)

	// 3. Batch score for ranking top-N drivers
	drivers := []string{"42", "17", "88", "5", "123"}
	scores := batchScore(ctx, "dispatch_ranker", drivers)
	for id, s := range scores {
		fmt.Printf("  driver %s → %.4f\n", id, s)
	}
}

func emitEvent(ctx context.Context, eventType, source, subject string, data any) {
	payload, _ := json.Marshal(map[string]any{
		"event_type": eventType,
		"source":     source,
		"subject":    subject,
		"data":       data,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST",
		"http://localhost:8080/v1/events",
		strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}

func scoreDriver(ctx context.Context, driverID string) (*ScoreResponse, error) {
	payload, _ := json.Marshal(ScoreRequest{
		ModelName:    "dispatch_ranker",
		EntityType:   "driver",
		EntityID:     driverID,
		FeatureNames: []string{"acceptance_rate", "avg_rating", "total_rides", "distance_to_rider", "surge_multiplier", "time_since_last_ride_min"},
	})

	resp, err := http.Post("http://localhost:8080/v1/score",
		"application/json",
		strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sr ScoreResponse
	json.NewDecoder(resp.Body).Decode(&sr)
	return &sr, nil
}

func batchScore(ctx context.Context, model string, driverIDs []string) map[string]float64 {
	results := make(map[string]float64)
	for _, id := range driverIDs {
		sr, err := scoreDriver(ctx, id)
		if err == nil {
			results[id] = sr.Score
		}
	}
	return results
}
