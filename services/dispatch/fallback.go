package dispatch

import (
	"math"
	"sort"
)

// FallbackRanker is the legacy hardcoded formula.
// This runs when mlplatform is unreachable.
// It must always work. No dependencies. No network calls.
type FallbackRanker struct{}

func NewFallbackRanker() *FallbackRanker {
	return &FallbackRanker{}
}

func (f *FallbackRanker) Rank(req *RideRequest, driverIDs []string) []RankedDriver {
	ranked := make([]RankedDriver, 0, len(driverIDs))

	for _, id := range driverIDs {
		dist := req.DriverDistances[id]
		rating := req.DriverRatings[id]

		score := -(dist*1.0 + (5.0-rating)*0.2)

		ranked = append(ranked, RankedDriver{
			DriverID: id,
			Score:    score,
			Distance: dist,
			Variant:  "fallback",
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	return ranked
}

// haversineKm computes distance between two lat/lng points.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
