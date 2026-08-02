package experiments

import (
	"hash/fnv"
	"sync"
)

type Experiment struct {
	Name      string    `json:"name"`
	ModelName string    `json:"model_name"`
	Variants  []Variant `json:"variants"`
	Status    string    `json:"status"` // "running", "paused", "completed"
}

type Variant struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"` // 0.0 - 1.0
}

type Manager struct {
	mu          sync.RWMutex
	experiments map[string]*Experiment // key: model_name
}

func NewManager() *Manager {
	return &Manager{experiments: make(map[string]*Experiment)}
}

func (m *Manager) Register(exp *Experiment) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.experiments[exp.ModelName] = exp
}

func (m *Manager) GetActive(modelName string) *Experiment {
	m.mu.RLock()
	defer m.mu.RUnlock()
	exp, ok := m.experiments[modelName]
	if !ok || exp.Status != "running" {
		return nil
	}
	return exp
}

// Assign deterministically assigns an entity to a variant.
func (e *Experiment) Assign(entityID string) string {
	h := fnv.New32a()
	h.Write([]byte(e.Name + ":" + entityID))
	bucket := float64(h.Sum32()%10000) / 10000.0

	cumulative := 0.0
	for _, v := range e.Variants {
		cumulative += v.Weight
		if bucket < cumulative {
			return v.Name
		}
	}
	return e.Variants[len(e.Variants)-1].Name
}
