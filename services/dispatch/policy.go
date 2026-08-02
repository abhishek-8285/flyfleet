package dispatch

// DispatchPolicy decides HOW to dispatch based on model confidence.
// This is where the product logic lives — not in the model, not in the
// platform, but in the decision layer that interprets the score.

type DispatchStrategy int

const (
	// Sequential: offer to one driver at a time, wait for response.
	// Used when the model is confident. Minimizes wasted offers.
	Sequential DispatchStrategy = iota

	// Parallel: offer to top N drivers simultaneously, first accept wins.
	// Used when the model is uncertain. Minimizes rider wait time.
	Parallel

	// Widen: expand search radius and re-rank.
	// Used when all scores are low. No good candidate nearby.
	Widen
)

type PolicyConfig struct {
	HighConfidenceThreshold  float64 `yaml:"high_confidence"`
	MediumConfidenceThreshold float64 `yaml:"medium_confidence"`
	ParallelOfferCount       int     `yaml:"parallel_count"`
	MaxSequentialOffers      int     `yaml:"max_sequential"`
	OfferTimeoutSeconds      int     `yaml:"offer_timeout_s"`
	WidenRadiusStepKm        float64 `yaml:"widen_step_km"`
	MaxRadiusKm              float64 `yaml:"max_radius_km"`
}

var DefaultPolicy = PolicyConfig{
	HighConfidenceThreshold:  0.85,
	MediumConfidenceThreshold: 0.60,
	ParallelOfferCount:       3,
	MaxSequentialOffers:      5,
	OfferTimeoutSeconds:      15,
	WidenRadiusStepKm:        2.0,
	MaxRadiusKm:              10.0,
}

// Decide selects a dispatch strategy based on the top candidate's score.
func Decide(ranked []RankedDriver, cfg PolicyConfig) (DispatchStrategy, []RankedDriver) {
	if len(ranked) == 0 {
		return Widen, nil
	}

	topScore := ranked[0].Score

	switch {
	case topScore >= cfg.HighConfidenceThreshold:
		limit := cfg.MaxSequentialOffers
		if limit > len(ranked) {
			limit = len(ranked)
		}
		return Sequential, ranked[:limit]

	case topScore >= cfg.MediumConfidenceThreshold:
		limit := cfg.ParallelOfferCount
		if limit > len(ranked) {
			limit = len(ranked)
		}
		return Parallel, ranked[:limit]

	default:
		return Widen, ranked
	}
}
