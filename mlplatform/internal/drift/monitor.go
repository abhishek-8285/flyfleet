package drift

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Monitor tracks feature distributions and alerts on drift.
type Monitor struct {
	mu         sync.Mutex
	windows    map[string]*FeatureWindow // key: "model:feature"
	threshold  float64
	windowSize int
	alertFn    func(feature string, score float64)
}

type FeatureWindow struct {
	Values   []float64
	Baseline []float64 // reference distribution
	Index    int
	Full     bool
}

func NewMonitor(threshold float64, windowSize int) *Monitor {
	return &Monitor{
		windows:    make(map[string]*FeatureWindow),
		threshold:  threshold,
		windowSize: windowSize,
	}
}

func (m *Monitor) OnAlert(fn func(feature string, score float64)) {
	m.alertFn = fn
}

// Observe records a feature value and checks for drift.
func (m *Monitor) Observe(ctx context.Context, model, feature string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := model + ":" + feature
	w, ok := m.windows[key]
	if !ok {
		w = &FeatureWindow{
			Values: make([]float64, m.windowSize),
		}
		m.windows[key] = w
	}

	w.Values[w.Index] = value
	w.Index = (w.Index + 1) % m.windowSize
	if w.Index == 0 {
		w.Full = true
	}

	if w.Full && len(w.Baseline) > 0 {
		score := psi(w.Baseline, w.Values)
		if score > m.threshold {
			log.Warn().
				Str("feature", key).
				Float64("psi", score).
				Msg("drift detected")
			if m.alertFn != nil {
				m.alertFn(key, score)
			}
		}
	}
}

// SetBaseline sets the reference distribution for a feature.
func (m *Monitor) SetBaseline(model, feature string, values []float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := model + ":" + feature
	w, ok := m.windows[key]
	if !ok {
		w = &FeatureWindow{Values: make([]float64, m.windowSize)}
		m.windows[key] = w
	}
	w.Baseline = values
}

// PSI (Population Stability Index) — simplified.
func psi(baseline, current []float64) float64 {
	bins := 10
	bHist := histogram(baseline, bins)
	cHist := histogram(current, bins)

	var score float64
	for i := range bHist {
		b := math.Max(bHist[i], 0.0001)
		c := math.Max(cHist[i], 0.0001)
		score += (c - b) * math.Log(c/b)
	}
	return score
}

// PSI computes the Population Stability Index between two feature distributions.
// Exported so batch activities and tests can reuse the same statistic the drift
// monitor uses internally.
func PSI(baseline, current []float64) float64 {
	return psi(baseline, current)
}

func histogram(values []float64, bins int) []float64 {
	if len(values) == 0 {
		return make([]float64, bins)
	}
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if min == max {
		h := make([]float64, bins)
		h[0] = 1.0
		return h
	}

	h := make([]float64, bins)
	width := (max - min) / float64(bins)
	for _, v := range values {
		idx := int((v - min) / width)
		if idx >= bins {
			idx = bins - 1
		}
		h[idx]++
	}
	total := float64(len(values))
	for i := range h {
		h[i] /= total
	}
	return h
}

// CheckInterval exposes the configured drift check cadence.
func CheckInterval() time.Duration {
	return time.Hour
}
