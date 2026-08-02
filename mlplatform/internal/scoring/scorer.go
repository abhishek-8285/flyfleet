package scoring

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/flyfleet/mlplatform/internal/eventbus"
	"github.com/flyfleet/mlplatform/internal/experiments"
	"github.com/flyfleet/mlplatform/internal/featurestore"
	"github.com/flyfleet/mlplatform/internal/inference"
	"github.com/flyfleet/mlplatform/internal/metrics"
)

// Scorer orchestrates: features -> experiment -> model -> prediction -> log.
type Scorer struct {
	features    featurestore.Store
	engine      *inference.Engine
	experiments *experiments.Manager
	bus         eventbus.Bus
	metrics     *metrics.Metrics
}

func NewScorer(fs featurestore.Store, eng *inference.Engine, exp *experiments.Manager, bus eventbus.Bus, m *metrics.Metrics) *Scorer {
	return &Scorer{
		features:    fs,
		engine:      eng,
		experiments: exp,
		bus:         bus,
		metrics:     m,
	}
}

type ScoreRequest struct {
	ModelName    string            `json:"model_name"`
	EntityType   string            `json:"entity_type"`
	EntityID     string            `json:"entity_id"`
	FeatureNames []string          `json:"feature_names"`
	Context      map[string]string `json:"context"` // for experiment assignment
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

func (s *Scorer) Score(ctx context.Context, req *ScoreRequest) (*ScoreResponse, error) {
	start := time.Now()

	// 1. Retrieve features
	fv, err := s.features.GetOnlineFeatures(ctx, req.EntityType, req.EntityID, req.FeatureNames)
	if err != nil {
		s.recordFailure(req.ModelName, "control")
		return nil, err
	}

	featureMap := make(map[string]float64, len(fv.Features))
	for _, f := range fv.Features {
		if v, ok := f.Value.(float64); ok {
			featureMap[f.Name] = v
		}
	}

	// 2. Experiment assignment
	variant := "control"
	experimentName := ""
	if exp := s.experiments.GetActive(req.ModelName); exp != nil {
		variant = exp.Assign(req.EntityID)
		experimentName = exp.Name
	}

	// 3. Predict
	modelName := req.ModelName
	if variant != "control" {
		modelName = req.ModelName + "_" + variant
	}

	score, err := s.engine.Predict(ctx, modelName, featureMap)
	if err != nil {
		s.recordFailure(req.ModelName, variant)
		return nil, err
	}

	latency := time.Since(start).Microseconds()

	resp := &ScoreResponse{
		Score:         score,
		ModelName:     modelName,
		Experiment:    experimentName,
		Variant:       variant,
		Features:      featureMap,
		LatencyMicros: latency,
	}

	// Emit scoring metrics
	s.metrics.ScoreRequests.WithLabelValues(req.ModelName, variant, "ok").Inc()
	s.metrics.ScoreLatency.WithLabelValues(modelName).Observe(time.Since(start).Seconds())

	// 4. Log prediction asynchronously
	go func() {
		_ = s.bus.Publish(context.Background(), "predictions", &eventbus.Event{
			ID:        req.EntityID + ":" + time.Now().Format(time.RFC3339Nano),
			Type:      "prediction",
			Subject:   req.EntityType + ":" + req.EntityID,
			Timestamp: time.Now().UTC(),
			Data:      mustMarshal(resp),
		})
	}()

	return resp, nil
}

// ScoreBatch scores multiple entities concurrently.
// Used by dispatch ranking: "score these 50 drivers, return sorted."
func (s *Scorer) ScoreBatch(ctx context.Context, modelName, entityType string, entityIDs []string, featureNames []string) (map[string]*ScoreResponse, error) {
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
			resp, err := s.Score(ctx, &ScoreRequest{
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

func (s *Scorer) recordFailure(modelName, variant string) {
	s.metrics.ScoreRequests.WithLabelValues(modelName, variant, "error").Inc()
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
