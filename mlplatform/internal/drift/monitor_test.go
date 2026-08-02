package drift

import (
	"context"
	"testing"
)

func TestObserve_NoAlertWhenStable(t *testing.T) {
	alerted := false
	m := NewMonitor(10.0, 10)
	m.OnAlert(func(feature string, score float64) { alerted = true })
	m.SetBaseline("m", "feat", []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		m.Observe(ctx, "m", "feat", 0.5)
	}
	if alerted {
		t.Fatal("expected no drift alert for a stable distribution")
	}
}

func TestObserve_AlertOnDrift(t *testing.T) {
	var alertedFeature string
	var alertedScore float64
	m := NewMonitor(0.05, 10)
	m.OnAlert(func(feature string, score float64) {
		alertedFeature = feature
		alertedScore = score
	})
	// Baseline concentrated at 0; current distribution spread across 0..9.
	m.SetBaseline("m", "feat", []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		m.Observe(ctx, "m", "feat", float64(i))
	}
	if alertedFeature == "" {
		t.Fatal("expected a drift alert for a divergent distribution")
	}
	if alertedFeature != "m:feat" {
		t.Fatalf("expected alert feature m:feat, got %q", alertedFeature)
	}
	if alertedScore <= 0.05 {
		t.Fatalf("expected PSI above threshold 0.05, got %v", alertedScore)
	}
}

func TestPSI_IdenticalDistributions(t *testing.T) {
	a := []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0, 0.1, 0.2}
	if got := PSI(a, a); got != 0 {
		t.Fatalf("expected 0 PSI for identical distributions, got %v", got)
	}
}

func TestPSI_DivergentDistributions(t *testing.T) {
	baseline := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	current := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	if got := PSI(baseline, current); got <= 0 {
		t.Fatalf("expected positive PSI for divergent distributions, got %v", got)
	}
}
