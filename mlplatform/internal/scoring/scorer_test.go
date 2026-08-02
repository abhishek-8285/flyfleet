package scoring

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/flyfleet/mlplatform/internal/eventbus"
	"github.com/flyfleet/mlplatform/internal/experiments"
	"github.com/flyfleet/mlplatform/internal/featurestore"
	"github.com/flyfleet/mlplatform/internal/inference"
	"github.com/flyfleet/mlplatform/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// --- fakes ---

type fakeModel struct {
	name, version string
	score         float64
}

func (f *fakeModel) Name() string                                        { return f.name }
func (f *fakeModel) Version() string                                     { return f.version }
func (f *fakeModel) Predict(context.Context, map[string]float64) (float64, error) {
	return f.score, nil
}
func (f *fakeModel) PredictBatch(context.Context, []map[string]float64) ([]float64, error) {
	return make([]float64, 0), nil
}
func (f *fakeModel) Close() error { return nil }

// errModel always fails Predict.
type errModel struct{ name, version string }

func (e *errModel) Name() string { return e.name }
func (e *errModel) Version() string { return e.version }
func (e *errModel) Predict(context.Context, map[string]float64) (float64, error) {
	return 0, errors.New("predict failed")
}
func (e *errModel) PredictBatch(context.Context, []map[string]float64) ([]float64, error) {
	return nil, errors.New("predict failed")
}
func (e *errModel) Close() error { return nil }

type fakeFeatureStore struct {
	fv  *featurestore.FeatureVector
	err error
}

func (f *fakeFeatureStore) GetOnlineFeatures(_ context.Context, _, _ string, _ []string) (*featurestore.FeatureVector, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.fv, nil
}
func (f *fakeFeatureStore) PutOnlineFeatures(context.Context, *featurestore.FeatureVector) error { return nil }
func (f *fakeFeatureStore) GetOfflineFeatures(context.Context, string, []string, []string, time.Time, time.Time) ([]*featurestore.FeatureVector, error) {
	return nil, nil
}
func (f *fakeFeatureStore) WriteOfflineFeatures(context.Context, []*featurestore.FeatureVector) error { return nil }
func (f *fakeFeatureStore) PointInTimeJoin(context.Context, string, []time.Time, []string) ([]*featurestore.FeatureVector, error) {
	return nil, nil
}
func (f *fakeFeatureStore) RegisterFeature(context.Context, *featurestore.FeatureDefinition) error { return nil }
func (f *fakeFeatureStore) ListFeatures(context.Context, string) ([]*featurestore.FeatureDefinition, error) {
	return nil, nil
}

type fakeBus struct {
	mu       sync.Mutex
	events   []*eventbus.Event
	notify   chan struct{}
	err      error
}

func newFakeBus() *fakeBus {
	return &fakeBus{notify: make(chan struct{}, 16)}
}

func (b *fakeBus) Publish(_ context.Context, subject string, event *eventbus.Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		return b.err
	}
	b.events = append(b.events, event)
	select {
	case b.notify <- struct{}{}:
	default:
	}
	return nil
}

func (b *fakeBus) Subscribe(context.Context, string, string) (<-chan *eventbus.Event, error) {
	return nil, nil
}

func (b *fakeBus) Close() error { return nil }

func (b *fakeBus) waitPublished(t *testing.T, timeout time.Duration) {
	t.Helper()
	select {
	case <-b.notify:
	case <-time.After(timeout):
		t.Fatal("timeout waiting for event publish")
	}
}

func (b *fakeBus) publishedCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.events)
}

func (b *fakeBus) firstEvent() *eventbus.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.events) == 0 {
		return nil
	}
	return b.events[0]
}

// --- helpers ---

func newTestScorer(t *testing.T) (*Scorer, *inference.Engine, *fakeBus, *metrics.Metrics) {
	t.Helper()
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	e := inference.NewEngine(t.TempDir(), m)
	bus := newFakeBus()
	fs := &fakeFeatureStore{
		fv: &featurestore.FeatureVector{
			EntityType: "driver",
			EntityID:   "d1",
			Features: []featurestore.FeatureValue{{
				Name:      "acceptance_rate",
				Value:     0.95,
				Timestamp: time.Now().UTC(),
			}},
		},
	}
	exp := experiments.NewManager()
	s := NewScorer(fs, e, exp, bus, m)
	return s, e, bus, m
}

// --- tests ---

func TestScorer_BasicScore(t *testing.T) {
	s, e, bus, m := newTestScorer(t)

	if err := e.LoadModel(&fakeModel{name: "churn", version: "v1"}, true); err != nil {
		t.Fatalf("load model: %v", err)
	}

	resp, err := s.Score(context.Background(), &ScoreRequest{
		ModelName:    "churn",
		EntityType:   "driver",
		EntityID:     "d1",
		FeatureNames: []string{"acceptance_rate"},
	})
	if err != nil {
		t.Fatalf("score: %v", err)
	}

	if resp.Score != 0.0 {
		t.Errorf("expected score 0.0, got %v", resp.Score)
	}
	if resp.ModelName != "churn" {
		t.Errorf("expected model churn, got %q", resp.ModelName)
	}
	if resp.Variant != "control" {
		t.Errorf("expected variant control, got %q", resp.Variant)
	}
	if resp.Experiment != "" {
		t.Errorf("expected empty experiment, got %q", resp.Experiment)
	}
	if resp.Features["acceptance_rate"] != 0.95 {
		t.Errorf("expected feature 0.95, got %v", resp.Features["acceptance_rate"])
	}
	if resp.LatencyMicros < 0 {
		t.Errorf("expected non-negative latency, got %d", resp.LatencyMicros)
	}

	// OK result metric
	if got, want := testutil.ToFloat64(m.ScoreRequests.WithLabelValues("churn", "control", "ok")), 1.0; got != want {
		t.Errorf("score_requests ok: got %v want %v", got, want)
	}

	// Prediction event published
	bus.waitPublished(t, time.Second)
	evt := bus.firstEvent()
	if evt == nil {
		t.Fatal("expected a published event")
	}
	if evt.Subject != "driver:d1" {
		t.Errorf("expected subject driver:d1, got %q", evt.Subject)
	}
	if evt.Type != "prediction" {
		t.Errorf("expected type prediction, got %q", evt.Type)
	}
}

func TestScorer_WithExperiment_TreatmentVariant(t *testing.T) {
	s, e, _, _ := newTestScorer(t)

	// 100% treatment weight → every entity gets "treatment".
	s.experiments.Register(&experiments.Experiment{
		Name:      "churn-ab",
		ModelName: "churn",
		Variants: []experiments.Variant{
			{Name: "control", Weight: 0.0},
			{Name: "treatment", Weight: 1.0},
		},
		Status: "running",
	})

	// The treatment variant routes to model "churn_treatment".
	if err := e.LoadModel(&fakeModel{name: "churn_treatment", version: "v1"}, true); err != nil {
		t.Fatalf("load model: %v", err)
	}

	resp, err := s.Score(context.Background(), &ScoreRequest{
		ModelName:    "churn",
		EntityType:   "driver",
		EntityID:     "d1",
		FeatureNames: []string{"acceptance_rate"},
	})
	if err != nil {
		t.Fatalf("score: %v", err)
	}

	if resp.ModelName != "churn_treatment" {
		t.Errorf("expected model churn_treatment, got %q", resp.ModelName)
	}
	if resp.Variant != "treatment" {
		t.Errorf("expected variant treatment, got %q", resp.Variant)
	}
	if resp.Experiment != "churn-ab" {
		t.Errorf("expected experiment churn-ab, got %q", resp.Experiment)
	}
}

func TestScorer_WithExperiment_ControlVariant(t *testing.T) {
	s, e, _, _ := newTestScorer(t)

	// 100% control weight → every entity gets "control", model name unchanged.
	s.experiments.Register(&experiments.Experiment{
		Name:      "churn-ab",
		ModelName: "churn",
		Variants: []experiments.Variant{
			{Name: "control", Weight: 1.0},
			{Name: "treatment", Weight: 0.0},
		},
		Status: "running",
	})

	if err := e.LoadModel(&fakeModel{name: "churn", version: "v1"}, true); err != nil {
		t.Fatalf("load model: %v", err)
	}

	resp, err := s.Score(context.Background(), &ScoreRequest{
		ModelName:    "churn",
		EntityType:   "driver",
		EntityID:     "d1",
		FeatureNames: []string{"acceptance_rate"},
	})
	if err != nil {
		t.Fatalf("score: %v", err)
	}

	if resp.Variant != "control" {
		t.Errorf("expected variant control, got %q", resp.Variant)
	}
	if resp.ModelName != "churn" {
		t.Errorf("expected model churn, got %q", resp.ModelName)
	}
}

func TestScorer_FeatureStoreError(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	e := inference.NewEngine(t.TempDir(), m)
	bus := newFakeBus()
	fs := &fakeFeatureStore{err: errors.New("redis down")}
	exp := experiments.NewManager()
	s := NewScorer(fs, e, exp, bus, m)

	_, err := s.Score(context.Background(), &ScoreRequest{
		ModelName:  "churn",
		EntityType: "driver",
		EntityID:   "d1",
	})
	if err == nil {
		t.Fatal("expected error from feature store failure")
	}

	// Error metric recorded with "control" variant (no experiment lookup happened).
	if got := testutil.ToFloat64(m.ScoreRequests.WithLabelValues("churn", "control", "error")); got != 1.0 {
		t.Errorf("expected 1 error metric, got %v", got)
	}
	if bus.publishedCount() != 0 {
		t.Errorf("expected 0 published events on error, got %d", bus.publishedCount())
	}
}

func TestScorer_PredictError(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	e := inference.NewEngine(t.TempDir(), m)
	bus := newFakeBus()
	fs := &fakeFeatureStore{
		fv: &featurestore.FeatureVector{
			EntityType: "driver",
			EntityID:   "d1",
			Features: []featurestore.FeatureValue{{
				Name:  "acceptance_rate",
				Value: 0.95,
			}},
		},
	}
	exp := experiments.NewManager()
	s := NewScorer(fs, e, exp, bus, m)

	// Load a model that always errors.
	if err := e.LoadModel(&errModel{name: "churn", version: "v1"}, true); err != nil {
		t.Fatalf("load model: %v", err)
	}

	_, err := s.Score(context.Background(), &ScoreRequest{
		ModelName:    "churn",
		EntityType:   "driver",
		EntityID:     "d1",
		FeatureNames: []string{"acceptance_rate"},
	})
	if err == nil {
		t.Fatal("expected error from prediction failure")
	}

	if got := testutil.ToFloat64(m.ScoreRequests.WithLabelValues("churn", "control", "error")); got != 1.0 {
		t.Errorf("expected 1 error metric, got %v", got)
	}
}

func TestScorer_SkipsNonFloatFeatures(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	e := inference.NewEngine(t.TempDir(), m)
	fs := &fakeFeatureStore{
		fv: &featurestore.FeatureVector{
			EntityType: "driver",
			EntityID:   "d1",
			Features: []featurestore.FeatureValue{
				{Name: "acceptance_rate", Value: 0.95},
				{Name: "tier", Value: "premium"},        // non-float, should be skipped
				{Name: "ride_count", Value: int64(42)},  // non-float, should be skipped
			},
		},
	}
	exp := experiments.NewManager()
	s := NewScorer(fs, e, exp, newFakeBus(), m)

	if err := e.LoadModel(&fakeModel{name: "churn", version: "v1"}, true); err != nil {
		t.Fatalf("load model: %v", err)
	}

	resp, err := s.Score(context.Background(), &ScoreRequest{
		ModelName:    "churn",
		EntityType:   "driver",
		EntityID:     "d1",
		FeatureNames: []string{"acceptance_rate", "tier", "ride_count"},
	})
	if err != nil {
		t.Fatalf("score: %v", err)
	}

	if len(resp.Features) != 1 {
		t.Errorf("expected 1 float feature in response, got %d", len(resp.Features))
	}
	if _, ok := resp.Features["acceptance_rate"]; !ok {
		t.Errorf("expected acceptance_rate in features map")
	}
}
