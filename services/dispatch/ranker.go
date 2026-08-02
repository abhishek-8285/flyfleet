package dispatch

import (
	"context"
	"sort"
	"time"

	ml "github.com/flyfleet/mlplatform/client"
)

// Ranker scores candidate drivers and returns them in dispatch order.
type Ranker struct {
	ml       *ml.Client
	features []string
	timeout  time.Duration
}

func NewRanker(mlURL string) *Ranker {
	return &Ranker{
		ml: ml.New(mlURL,
			ml.WithTimeout(3*time.Second),
			ml.WithRetry(2, 100*time.Millisecond, 1*time.Second),
		),
		features: []string{
			"distance_to_rider_km",
			"acceptance_rate_7d",
			"rides_completed_today",
			"minutes_since_last_ride",
			"driver_rating",
			"is_heading_toward_rider",
			"surge_multiplier",
			"time_of_day_hour",
			"day_of_week",
			"zone_demand_density",
			"zone_supply_density",
			"supply_demand_ratio",
			"effective_distance",
		},
		timeout: 2 * time.Second,
	}
}

type RankedDriver struct {
	DriverID string  `json:"driver_id"`
	Score    float64 `json:"score"`
	Distance float64 `json:"distance_km"`
	Variant  string  `json:"variant"`
}

// Rank returns drivers sorted by predicted success probability.
func (r *Ranker) Rank(ctx context.Context, rideReq *RideRequest, driverIDs []string) ([]RankedDriver, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	scores, err := r.ml.ScoreBatch(ctx,
		"dispatch_ranker",
		"driver",
		driverIDs,
		r.features,
	)
	if err != nil {
		return nil, err
	}

	ranked := make([]RankedDriver, 0, len(scores))
	for id, resp := range scores {
		ranked = append(ranked, RankedDriver{
			DriverID: id,
			Score:    resp.Score,
			Variant:  resp.Variant,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	return ranked, nil
}
