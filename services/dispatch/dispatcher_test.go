package dispatch

import (
	"testing"
)

func TestDecide_HighConfidence(t *testing.T) {
	ranked := []RankedDriver{
		{DriverID: "1", Score: 0.92},
		{DriverID: "2", Score: 0.85},
		{DriverID: "3", Score: 0.70},
	}

	strategy, candidates := Decide(ranked, DefaultPolicy)

	if strategy != Sequential {
		t.Errorf("expected Sequential, got %d", strategy)
	}
	if len(candidates) != 3 {
		t.Errorf("expected 3 candidates, got %d", len(candidates))
	}
}

func TestDecide_MediumConfidence(t *testing.T) {
	ranked := []RankedDriver{
		{DriverID: "1", Score: 0.72},
		{DriverID: "2", Score: 0.68},
		{DriverID: "3", Score: 0.65},
		{DriverID: "4", Score: 0.60},
	}

	strategy, candidates := Decide(ranked, DefaultPolicy)

	if strategy != Parallel {
		t.Errorf("expected Parallel, got %d", strategy)
	}
	if len(candidates) != DefaultPolicy.ParallelOfferCount {
		t.Errorf("expected %d candidates, got %d", DefaultPolicy.ParallelOfferCount, len(candidates))
	}
}

func TestDecide_LowConfidence(t *testing.T) {
	ranked := []RankedDriver{
		{DriverID: "1", Score: 0.40},
		{DriverID: "2", Score: 0.35},
	}

	strategy, _ := Decide(ranked, DefaultPolicy)

	if strategy != Widen {
		t.Errorf("expected Widen, got %d", strategy)
	}
}

func TestDecide_Empty(t *testing.T) {
	strategy, candidates := Decide(nil, DefaultPolicy)

	if strategy != Widen {
		t.Errorf("expected Widen for empty, got %d", strategy)
	}
	if candidates != nil {
		t.Error("expected nil candidates for empty input")
	}
}

func TestFallbackRanker(t *testing.T) {
	fb := NewFallbackRanker()

	req := &RideRequest{
		RideID: "ride-1",
		DriverDistances: map[string]float64{
			"d1": 1.2,
			"d2": 3.5,
			"d3": 0.8,
		},
		DriverRatings: map[string]float64{
			"d1": 4.8,
			"d2": 4.2,
			"d3": 3.9,
		},
	}

	ranked := fb.Rank(req, []string{"d1", "d2", "d3"})

	if len(ranked) != 3 {
		t.Fatalf("expected 3 ranked, got %d", len(ranked))
	}

	if ranked[0].DriverID != "d3" {
		t.Errorf("expected d3 first, got %s", ranked[0].DriverID)
	}

	for _, r := range ranked {
		if r.Variant != "fallback" {
			t.Errorf("expected variant fallback, got %s", r.Variant)
		}
	}
}

func TestHaversine(t *testing.T) {
	dist := haversineKm(40.7580, -73.9855, 40.7061, -73.9969)
	if dist < 5.0 || dist > 6.5 {
		t.Errorf("expected ~5.7km, got %.2f", dist)
	}
}
