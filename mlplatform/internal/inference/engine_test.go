package inference

import (
	"context"
	"testing"

	"github.com/flyfleet/mlplatform/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeModel struct {
	name, version string
}

func (f *fakeModel) Name() string                                       { return f.name }
func (f *fakeModel) Version() string                                    { return f.version }
func (f *fakeModel) Predict(context.Context, map[string]float64) (float64, error) {
	return 0.42, nil
}
func (f *fakeModel) PredictBatch(context.Context, []map[string]float64) ([]float64, error) {
	return make([]float64, 0), nil
}
func (f *fakeModel) Close() error { return nil }

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	return NewEngine(t.TempDir(), m)
}

func TestEngine_LoadAndPredict(t *testing.T) {
	e := newTestEngine(t)
	if err := e.LoadModel(&fakeModel{name: "churn", version: "v1"}, true); err != nil {
		t.Fatalf("load: %v", err)
	}

	score, err := e.Predict(context.Background(), "churn", map[string]float64{"x": 1})
	if err != nil {
		t.Fatalf("predict: %v", err)
	}
	if score != 0.42 {
		t.Fatalf("expected score 0.42, got %v", score)
	}
}

func TestEngine_HotSwap(t *testing.T) {
	e := newTestEngine(t)
	if err := e.LoadModel(&fakeModel{name: "churn", version: "v1"}, true); err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := e.LoadModel(&fakeModel{name: "churn", version: "v2"}, false); err != nil {
		t.Fatalf("load v2: %v", err)
	}
	if err := e.HotSwap("churn", "v2"); err != nil {
		t.Fatalf("hotswap: %v", err)
	}
}

func TestEngine_PredictUnknownModel(t *testing.T) {
	e := newTestEngine(t)
	if _, err := e.Predict(context.Background(), "nope", nil); err != ErrModelNotFound {
		t.Fatalf("expected ErrModelNotFound, got %v", err)
	}
}

func TestEngine_ActiveModelsGauge(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	e := NewEngine(t.TempDir(), m)
	if err := e.LoadModel(&fakeModel{name: "churn", version: "v1"}, true); err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, want := testutil.ToFloat64(m.ActiveModels), 1.0; got != want {
		t.Fatalf("active models gauge: got %v want %v", got, want)
	}
}
