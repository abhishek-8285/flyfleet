package engine

import (
	"context"

	"github.com/google/uuid"
)

type SuggestionEngine interface {
	Suggest(ctx context.Context, req SuggestionRequest) ([]Suggestion, error)
}

type SuggestionRequest struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	ShipmentID   uuid.UUID `json:"shipment_id"`
	PickupLat    float64   `json:"pickup_lat"`
	PickupLng    float64   `json:"pickup_lng"`
	VehicleType  string    `json:"vehicle_type"`
	CargoWeight  float64   `json:"cargo_weight"`
	TopN         int       `json:"top_n"`
}

type Suggestion struct {
	VehicleID   uuid.UUID `json:"vehicle_id"`
	DriverID    uuid.UUID `json:"driver_id"`
	Rank        int       `json:"rank"`
	Score       float64   `json:"score"`
	DistKm      float64   `json:"dist_km"`
	EstimatedArrival string `json:"estimated_arrival"`
	Reasons     []string  `json:"reasons"`
}

type suggestionEngine struct {
	// dependencies
}

func NewSuggestionEngine() SuggestionEngine {
	return &suggestionEngine{}
}

func (s *suggestionEngine) Suggest(ctx context.Context, req SuggestionRequest) ([]Suggestion, error) {
	// Implementation will use scoring and availability services
	return []Suggestion{}, nil
}
