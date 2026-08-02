package dispatch

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type RideRequest struct {
	RideID          string             `json:"ride_id"`
	RiderID         string             `json:"rider_id"`
	ZoneID          string             `json:"zone_id"`
	Lat             float64            `json:"lat"`
	Lng             float64            `json:"lng"`
	SearchRadiusKm  float64            `json:"search_radius_km"`
	SurgeMultiplier float64            `json:"surge_multiplier"`
	DriverDistances map[string]float64 `json:"driver_distances"`
	DriverRatings   map[string]float64 `json:"driver_ratings"`
}

type DispatchResult struct {
	DriverID          string           `json:"driver_id"`
	Score             float64          `json:"score"`
	Variant           string           `json:"variant"`
	Strategy          DispatchStrategy `json:"strategy"`
	Attempt           int              `json:"attempt"`
	OfferType         string           `json:"offer_type"`
	TotalCandidates   int              `json:"total_candidates"`
	DispatchLatencyMs int64            `json:"dispatch_latency_ms"`
}

// TripOffer represents a trip offer from fleet-core.
type TripOffer struct {
	ID        string `json:"id"`
	TripID    string `json:"trip_id"`
	DriverID  string `json:"driver_id"`
	Status    string `json:"status"` // OFFERED, ACCEPTED, REJECTED, EXPIRED
}

// FleetCoreSuggestion represents a driver suggestion from fleet-core's suggestion engine.
type FleetCoreSuggestion struct {
	DriverID   string  `json:"driver_id"`
	DriverName string  `json:"driver_name"`
	Score      float64 `json:"score"`
	DistKm     float64 `json:"dist_km"`
}

var (
	ErrNoDriversAvailable = fmt.Errorf("no drivers available")
	ErrAllDriversRejected = fmt.Errorf("all drivers rejected")
)

// Start wires the dispatcher into your existing ride request handler.
// Call this from your main API server's route setup.
func Start(mlURL, fleetCoreURL string) *Dispatcher {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	emitter := NewEventEmitter(mlURL)
	dispatcher := NewDispatcher(mlURL, fleetCoreURL, emitter)

	log.Info().Str("mlplatform_url", mlURL).Str("fleetcore_url", fleetCoreURL).Msg("dispatch ranker initialized")

	return dispatcher
}
