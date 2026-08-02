package inference

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/flyfleet/mlplatform/internal/metrics"
)

// Model is a loaded, ready-to-serve model.
type Model interface {
	Name() string
	Version() string
	Predict(ctx context.Context, features map[string]float64) (float64, error)
	PredictBatch(ctx context.Context, batch []map[string]float64) ([]float64, error)
	Close() error
}

// Engine manages model lifecycle: loading, hot-swapping, routing.
type Engine struct {
	mu        sync.RWMutex
	models    map[string]Model // key: "name:version"
	active    map[string]string // key: "name" -> active version
	previous  map[string]string // key: "name" -> previous active version (for rollback)
	modelDir  string
	metrics   *metrics.Metrics
}

func NewEngine(modelDir string, m *metrics.Metrics) *Engine {
	return &Engine{
		models:   make(map[string]Model),
		active:   make(map[string]string),
		previous: make(map[string]string),
		modelDir: modelDir,
		metrics:  m,
	}
}

// LoadModel loads a model from disk and registers it.
func (e *Engine) LoadModel(model Model, activate bool) error {
	key := model.Name() + ":" + model.Version()

	e.mu.Lock()
	defer e.mu.Unlock()

	e.models[key] = model
	if activate {
		e.active[model.Name()] = model.Version()
		e.metrics.ActiveModels.Set(float64(len(e.active)))
	}
	return nil
}

// HotSwap atomically replaces the active model version.
func (e *Engine) HotSwap(name, newVersion string) error {
	key := name + ":" + newVersion

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.models[key]; !ok {
		return ErrModelNotFound
	}

	// Record previous version for rollback.
	if oldVer, ok := e.active[name]; ok {
		e.previous[name] = oldVer
	}

	// Deactivate old
	if oldVer, ok := e.active[name]; ok {
		oldKey := name + ":" + oldVer
		if old, ok := e.models[oldKey]; ok {
			old.Close()
			delete(e.models, oldKey)
		}
	}

	e.active[name] = newVersion
	e.metrics.ActiveModels.Set(float64(len(e.active)))
	return nil
}

// List returns all loaded models with their active status.
func (e *Engine) List() []ModelInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var out []ModelInfo
	for key, m := range e.models {
		parts := strings.SplitN(key, ":", 2)
		version := ""
		if len(parts) == 2 {
			version = parts[1]
		}
		out = append(out, ModelInfo{
			Name:    m.Name(),
			Version: version,
			Active:  e.active[m.Name()] == version,
		})
	}
	return out
}

// PreviousVersion returns the version that was active before the current one.
// Used for rollback.
func (e *Engine) PreviousVersion(name string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	prev, ok := e.previous[name]
	if !ok {
		return "", fmt.Errorf("no previous version for %s", name)
	}
	return prev, nil
}

// ModelInfo describes a loaded model version.
type ModelInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Active  bool   `json:"active"`
}

// Predict routes to the active model version.
func (e *Engine) Predict(ctx context.Context, modelName string, features map[string]float64) (float64, error) {
	e.mu.RLock()
	version, ok := e.active[modelName]
	if !ok {
		e.mu.RUnlock()
		return 0, ErrModelNotFound
	}
	model := e.models[modelName+":"+version]
	e.mu.RUnlock()

	return model.Predict(ctx, features)
}

func (e *Engine) PredictBatch(ctx context.Context, modelName string, batch []map[string]float64) ([]float64, error) {
	e.mu.RLock()
	version, ok := e.active[modelName]
	if !ok {
		e.mu.RUnlock()
		return nil, ErrModelNotFound
	}
	model := e.models[modelName+":"+version]
	e.mu.RUnlock()

	return model.PredictBatch(ctx, batch)
}

var ErrModelNotFound = fmt.Errorf("model not found")
