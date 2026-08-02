package streaming

import (
	"context"
	"encoding/json"

	"github.com/flyfleet/mlplatform/internal/eventbus"
	"github.com/flyfleet/mlplatform/internal/featurestore"
	"github.com/flyfleet/mlplatform/internal/warehouse"
)

// FeaturePipeline computes real-time features from event streams.
type FeaturePipeline struct {
	store     featurestore.Store
	warehouse *warehouse.ClickHouseStore
}

func NewFeaturePipeline(store featurestore.Store, wh *warehouse.ClickHouseStore) *FeaturePipeline {
	return &FeaturePipeline{store: store, warehouse: wh}
}

// Example: compute driver acceptance rate from OrderCreated/DriverAccepted events.
func (fp *FeaturePipeline) HandleDriverEvent(ctx context.Context, evt *eventbus.Event) error {
	var payload struct {
		DriverID string `json:"driver_id"`
		Action   string `json:"action"` // "accepted", "rejected"
	}
	if err := json.Unmarshal(evt.Data, &payload); err != nil {
		return err
	}

	// Get current counters
	fv, err := fp.store.GetOnlineFeatures(ctx, "driver", payload.DriverID, []string{
		"total_offers", "total_accepted", "acceptance_rate",
	})
	if err != nil {
		return err
	}

	totalOffers := getFloat(fv, "total_offers") + 1
	totalAccepted := getFloat(fv, "total_accepted")
	if payload.Action == "accepted" {
		totalAccepted++
	}
	rate := totalAccepted / totalOffers

	// Write updated features
	return fp.store.PutOnlineFeatures(ctx, &featurestore.FeatureVector{
		EntityID:   payload.DriverID,
		EntityType: "driver",
		Features: []featurestore.FeatureValue{
			{Name: "total_offers", Value: totalOffers, Timestamp: evt.Timestamp},
			{Name: "total_accepted", Value: totalAccepted, Timestamp: evt.Timestamp},
			{Name: "acceptance_rate", Value: rate, Timestamp: evt.Timestamp},
		},
	})
}

func getFloat(fv *featurestore.FeatureVector, name string) float64 {
	if fv == nil {
		return 0
	}
	for _, f := range fv.Features {
		if f.Name == name {
			v, _ := f.Value.(float64)
			return v
		}
	}
	return 0
}
