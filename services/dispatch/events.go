package dispatch

import (
	"context"
	"time"

	ml "github.com/flyfleet/mlplatform/client"
)

// EventEmitter publishes dispatch lifecycle events to mlplatform.
// These events feed the feature computation pipeline and training data.
type EventEmitter struct {
	ml     *ml.Client
	source string
}

func NewEventEmitter(mlURL string) *EventEmitter {
	return &EventEmitter{
		ml:     ml.New(mlURL, ml.WithTimeout(2*time.Second)),
		source: "dispatch-service",
	}
}

func (e *EventEmitter) EmitDispatchResult(ctx context.Context, req *RideRequest, result *DispatchResult) {
	e.ml.Emit(ctx, &ml.EmitRequest{
		EventType: "dispatch.completed",
		Source:    e.source,
		Subject:   "ride:" + req.RideID,
		Data: map[string]any{
			"ride_id":              req.RideID,
			"driver_id":            result.DriverID,
			"rider_id":             req.RiderID,
			"score":                result.Score,
			"variant":              result.Variant,
			"strategy":             result.OfferType,
			"attempt":              result.Attempt,
			"total_candidates":     result.TotalCandidates,
			"dispatch_latency_ms":  result.DispatchLatencyMs,
			"zone_id":              req.ZoneID,
			"surge_multiplier":     req.SurgeMultiplier,
		},
	})
}

func (e *EventEmitter) EmitOfferSent(ctx context.Context, rideID, driverID string, attempt int) {
	e.ml.Emit(ctx, &ml.EmitRequest{
		EventType: "dispatch.offer_sent",
		Source:    e.source,
		Subject:   "driver:" + driverID,
		Data: map[string]any{
			"ride_id":   rideID,
			"driver_id": driverID,
			"attempt":   attempt,
		},
	})
}

func (e *EventEmitter) EmitOfferRejected(ctx context.Context, rideID, driverID string) {
	e.ml.Emit(ctx, &ml.EmitRequest{
		EventType: "driver.rejected",
		Source:    e.source,
		Subject:   "driver:" + driverID,
		Data: map[string]any{
			"ride_id":   rideID,
			"driver_id": driverID,
			"action":    "rejected",
		},
	})
}

func (e *EventEmitter) EmitFallbackUsed(ctx context.Context, rideID string) {
	e.ml.Emit(ctx, &ml.EmitRequest{
		EventType: "dispatch.fallback_used",
		Source:    e.source,
		Subject:   "ride:" + rideID,
		Data: map[string]any{
			"ride_id": rideID,
			"reason":  "ml_platform_unavailable",
		},
	})
}
