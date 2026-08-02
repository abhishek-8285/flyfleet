package dispatch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uber/h3-go/v4"
)

type Dispatcher struct {
	ranker        *Ranker
	fallback      *FallbackRanker
	events        *EventEmitter
	policy        PolicyConfig
	fleetCoreURL  string
	store         *InMemoryStore
	httpClient    *http.Client
}

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
	DispatchID     string           `json:"dispatch_id"`
	RideID         string           `json:"ride_id"`
	DriverID       string           `json:"driver_id"`
	State          RideState        `json:"state"`
	Score          float64          `json:"score"`
	Variant        string           `json:"variant"`
	Strategy       DispatchStrategy `json:"strategy"`
	Attempt        int              `json:"attempt"`
	OfferType      string           `json:"offer_type"`
	TotalCandidates int             `json:"total_candidates"`
	DispatchLatencyMs int64         `json:"dispatch_latency_ms"`
	DispatchedAt   *time.Time       `json:"dispatched_at,omitempty"`
}

type TripOffer struct {
	ID        string `json:"id"`
	TripID    string `json:"trip_id"`
	DriverID  string `json:"driver_id"`
	Status    string `json:"status"`
}

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

func NewDispatcher(mlURL, fleetCoreURL string, emitter *EventEmitter, store *InMemoryStore) *Dispatcher {
	return &Dispatcher{
		ranker:       NewRanker(mlURL),
		fallback:     NewFallbackRanker(),
		events:       emitter,
		policy:       DefaultPolicy,
		fleetCoreURL: fleetCoreURL,
		store:        store,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Dispatch is the main entry point. Called when a rider requests a ride.
// Returns the matched driver ID, or an error if no driver could be matched.
// Supports idempotency via the Idempotency-Key header (passed in ctx via IdempotencyKey).
func (d *Dispatcher) Dispatch(ctx context.Context, req *RideRequest) (*DispatchResult, error) {
	start := time.Now()

	rideID := req.RideID
	if rideID == "" {
		rideID = "ride-" + time.Now().UTC().Format("20060102-150405-") + req.RiderID
		req.RideID = rideID
	}

	idempotencyKey := IdempotencyKeyFromContext(ctx)
	if idempotencyKey != "" {
		if record, ok := d.store.GetIdempotencyResult(idempotencyKey); ok {
			log.Info().Str("idempotency_key", idempotencyKey).Str("ride_id", record.RideID).Msg("idempotency hit")
			return record.Result, nil
		}
	}

	ride := &RideStatus{
		RideID:       rideID,
		State:        RideRequested,
		RiderID:      req.RiderID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		StateEntered: time.Now(),
	}

	if err := d.store.CreateRide(ride); err != nil {
		return nil, fmt.Errorf("create ride: %w", err)
	}

	if idempotencyKey != "" {
		d.store.StoreIdempotencyResult(idempotencyKey, rideID, "", nil)
	}

	if err := ride.TransitionTo(DriverSearching); err != nil {
		return nil, fmt.Errorf("transition to searching: %w", err)
	}
	if err := d.store.UpdateRide(ride); err != nil {
		return nil, fmt.Errorf("update ride state: %w", err)
	}

	drivers, err := d.findNearbyDrivers(ctx, req)
	if err != nil {
		ride.TransitionTo(RideFailed)
		d.store.UpdateRide(ride)
		return nil, fmt.Errorf("find drivers: %w", err)
	}

	if len(drivers) == 0 {
		ride.TransitionTo(RideExpired)
		d.store.UpdateRide(ride)
		return nil, ErrNoDriversAvailable
	}

	log.Info().
		Str("ride_id", req.RideID).
		Int("candidates", len(drivers)).
		Msg("dispatching ride")

	ranked, err := d.ranker.Rank(ctx, req, drivers)
	if err != nil {
		log.Warn().Err(err).Str("ride_id", req.RideID).Msg("ml ranker failed, using fallback")
		ranked = d.fallback.Rank(req, drivers)
		d.events.EmitFallbackUsed(ctx, req.RideID)
	}

	strategy, candidates := Decide(ranked, d.policy)

	log.Info().
		Str("ride_id", req.RideID).
		Int("strategy", int(strategy)).
		Int("candidates", len(candidates)).
		Float64("top_score", ranked[0].Score).
		Msg("dispatch strategy selected")

	var result *DispatchResult
	switch strategy {
	case Sequential:
		result, err = d.sequentialDispatch(ctx, req, candidates)
	case Parallel:
		result, err = d.parallelDispatch(ctx, req, candidates)
	case Widen:
		result, err = d.widenDispatch(ctx, req, ranked)
	}

	if err != nil {
		ride.TransitionTo(RideFailed)
		d.store.UpdateRide(ride)
		return nil, err
	}

	if err := ride.TransitionTo(DriverAssigned); err != nil {
		return nil, fmt.Errorf("transition to assigned: %w", err)
	}
	ride.DriverID = result.DriverID
	if err := d.store.UpdateRide(ride); err != nil {
		return nil, fmt.Errorf("update ride state: %w", err)
	}

	dispatchID := "dispatch-" + rideID
	result.DispatchID = dispatchID
	result.RideID = rideID
	result.State = DriverAssigned
	result.Strategy = strategy
	result.TotalCandidates = len(drivers)
	result.DispatchLatencyMs = time.Since(start).Milliseconds()
	now := time.Now()
	result.DispatchedAt = &now

	assignment := &DispatchAssignment{
		DispatchID: dispatchID,
		RideID:     rideID,
		DriverID:   result.DriverID,
		Status:     DriverAssigned,
		Score:      result.Score,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	d.store.CreateAssignment(assignment)

	d.events.EmitDispatchResult(ctx, req, result)

	log.Info().
		Str("ride_id", req.RideID).
		Str("driver_id", result.DriverID).
		Float64("score", result.Score).
		Int64("latency_ms", result.DispatchLatencyMs).
		Msg("ride dispatched")

	return result, nil
}

func (d *Dispatcher) sequentialDispatch(ctx context.Context, req *RideRequest, candidates []RankedDriver) (*DispatchResult, error) {
	for i, c := range candidates {
		d.events.EmitOfferSent(ctx, req.RideID, c.DriverID, i+1)

		accepted, err := d.offerRide(ctx, req, c.DriverID, d.policy.OfferTimeoutSeconds)
		if err != nil {
			continue
		}
		if accepted {
			return &DispatchResult{
				DriverID:  c.DriverID,
				Score:     c.Score,
				Variant:   c.Variant,
				Attempt:   i + 1,
				OfferType: "sequential",
			}, nil
		}

		d.events.EmitOfferRejected(ctx, req.RideID, c.DriverID)
	}

	return nil, ErrAllDriversRejected
}

func (d *Dispatcher) parallelDispatch(ctx context.Context, req *RideRequest, candidates []RankedDriver) (*DispatchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.policy.OfferTimeoutSeconds)*time.Second)
	defer cancel()

	type offerResult struct {
		driver   RankedDriver
		accepted bool
	}

	resultCh := make(chan offerResult, len(candidates))

	for _, c := range candidates {
		go func(candidate RankedDriver) {
			accepted, _ := d.offerRide(ctx, req, candidate.DriverID, d.policy.OfferTimeoutSeconds)
			resultCh <- offerResult{driver: candidate, accepted: accepted}
		}(c)
	}

	for range candidates {
		res := <-resultCh
		if res.accepted {
			cancel()
			return &DispatchResult{
				DriverID:  res.driver.DriverID,
				Score:     res.driver.Score,
				Variant:   res.driver.Variant,
				OfferType: "parallel",
			}, nil
		}
	}

	return nil, ErrAllDriversRejected
}

func (d *Dispatcher) widenDispatch(ctx context.Context, req *RideRequest, currentRanked []RankedDriver) (*DispatchResult, error) {
	radius := req.SearchRadiusKm

	for radius < d.policy.MaxRadiusKm {
		radius += d.policy.WidenRadiusStepKm

		widerDrivers, err := d.findDriversInRadius(ctx, req, radius)
		if err != nil || len(widerDrivers) == 0 {
			continue
		}

		ranked, err := d.ranker.Rank(ctx, req, widerDrivers)
		if err != nil {
			ranked = d.fallback.Rank(req, widerDrivers)
		}

		if len(ranked) > 0 && ranked[0].Score >= d.policy.MediumConfidenceThreshold {
			return d.sequentialDispatch(ctx, req, ranked[:3])
		}
	}

	if len(currentRanked) > 0 {
		return d.sequentialDispatch(ctx, req, currentRanked[:1])
	}

	return nil, ErrNoDriversAvailable
}

// offerRide creates a trip in fleet-core and offers it to the driver.
// Returns true if the driver accepts within the timeout.
func (d *Dispatcher) offerRide(ctx context.Context, req *RideRequest, driverID string, timeoutSec int) (bool, error) {
	tripID, err := d.createTrip(ctx, req)
	if err != nil {
		log.Warn().Err(err).Str("ride_id", req.RideID).Msg("failed to create trip in fleet-core")
		return false, err
	}

	offerURL := fmt.Sprintf("%s/trips/%s/offer", d.fleetCoreURL, tripID)
	offerBody := fmt.Sprintf(`{"driver_id":"%s"}`, driverID)
	offerReq, _ := http.NewRequestWithContext(ctx, "POST", offerURL, bytes.NewReader([]byte(offerBody)))
	offerReq.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(offerReq)
	if err != nil {
		return false, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("offer request failed: %d %s", resp.StatusCode, string(body))
	}

	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	pollURL := fmt.Sprintf("%s/trips/offers?driver_id=%s", d.fleetCoreURL, driverID)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		pollReq, _ := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
		pollResp, err := d.httpClient.Do(pollReq)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		body, _ := io.ReadAll(pollResp.Body)
		pollResp.Body.Close()

		if pollResp.StatusCode == http.StatusOK {
			var offers []TripOffer
			if err := json.Unmarshal(body, &offers); err == nil && len(offers) > 0 {
				for _, offer := range offers {
					if offer.DriverID == driverID && offer.TripID == tripID {
						switch offer.Status {
						case "ACCEPTED":
							return true, nil
						case "REJECTED", "EXPIRED":
							return false, nil
						}
					}
				}
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false, nil
}

// findNearbyDrivers uses H3 spatial indexing to find drivers near the pickup.
// It converts the pickup location to an H3 cell, finds neighboring cells
// within the search radius, and queries fleet-core for drivers in those cells.
func (d *Dispatcher) findNearbyDrivers(ctx context.Context, req *RideRequest) ([]string, error) {
	pickupCell, err := h3.NewLatLng(req.Lat, req.Lng).Cell(8)
	if err != nil {
		return nil, fmt.Errorf("h3 cell: %w", err)
	}

	k := h3RingForRadius(req.SearchRadiusKm)

	// Get all H3 cells in the disk (all cells within k rings).
	cells, err := h3.GridDisk(pickupCell, k)
	if err != nil {
		return nil, fmt.Errorf("h3 grid disk: %w", err)
	}

	// Query fleet-core for drivers in these cells.
	driverIDs, err := d.queryFleetCoreByCells(ctx, cells, req)
	if err != nil {
		return nil, err
	}

	// If H3 disk query returned too few drivers, widen by one more ring.
	if len(driverIDs) < 3 && k < 10 {
		cells, err = h3.GridDisk(pickupCell, k+1)
		if err != nil {
			return nil, fmt.Errorf("h3 grid disk widen: %w", err)
		}
		driverIDs, err = d.queryFleetCoreByCells(ctx, cells, req)
		if err != nil {
			return nil, err
		}
	}

	return driverIDs, nil
}

// findDriversInRadius uses H3 disk expansion to find drivers within a custom radius.
func (d *Dispatcher) findDriversInRadius(ctx context.Context, req *RideRequest, radiusKm float64) ([]string, error) {
	pickupCell, err := h3.NewLatLng(req.Lat, req.Lng).Cell(8)
	if err != nil {
		return nil, fmt.Errorf("h3 cell: %w", err)
	}
	k := h3RingForRadius(radiusKm)
	cells, err := h3.GridDisk(pickupCell, k)
	if err != nil {
		return nil, fmt.Errorf("h3 grid disk: %w", err)
	}
	return d.queryFleetCoreByCells(ctx, cells, req)
}

// queryFleetCoreByCells queries fleet-core's suggestion engine for drivers
// in the given H3 cells. It converts cells to strings and passes the pickup
// location and radius; fleet-core handles the actual spatial filtering.
func (d *Dispatcher) queryFleetCoreByCells(ctx context.Context, cells []h3.Cell, req *RideRequest) ([]string, error) {
	cellStrings := make([]string, len(cells))
	for i, c := range cells {
		cellStrings[i] = h3.IndexToString(uint64(c))
	}

	suggestReq := map[string]any{
		"pickup_lat": req.Lat,
		"pickup_lng": req.Lng,
		"radius_km":  req.SearchRadiusKm,
		"h3_cells":   cellStrings,
		"top_n":      50,
	}
	body, err := json.Marshal(suggestReq)
	if err != nil {
		return nil, err
	}

	resp, err := d.httpClient.Post(
		fmt.Sprintf("%s/dispatch/suggest", d.fleetCoreURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fleet-core suggest returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var suggestions []FleetCoreSuggestion
	if err := json.Unmarshal(data, &suggestions); err != nil {
		return nil, err
	}

	driverIDs := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		driverIDs = append(driverIDs, s.DriverID)
	}
	return driverIDs, nil
}

// h3RingForRadius returns the number of H3 rings needed to cover a given radius in km.
// Resolution 8 cells are approximately 1.4km across, so each ring adds ~1.4km.
func h3RingForRadius(radiusKm float64) int {
	// Resolution 8 hexagon width ≈ 1.4km. Each ring adds one hexagon width.
	rings := int(radiusKm / 1.4)
	if rings < 1 {
		return 1
	}
	if rings > 10 {
		return 10
	}
	return rings
}

// OfferRide offers a ride to a specific driver and waits for acceptance.
// It atomically claims the driver before offering to prevent double assignment.
func (d *Dispatcher) OfferRide(ctx context.Context, req *RideRequest, driverID string, timeoutSec int) (bool, error) {
	lock := d.store.GetDriverLock(driverID)
	lock.Lock()
	defer lock.Unlock()

	if err := d.store.ClaimDriver(driverID); err != nil {
		return false, err
	}

	defer func() {
		if err := d.store.ReleaseDriver(driverID); err != nil {
			log.Warn().Str("driver_id", driverID).Err(err).Msg("failed to release driver")
		}
	}()

	return d.offerRide(ctx, req, driverID, timeoutSec)
}

// FindNearbyDrivers finds drivers near the pickup location using H3 spatial indexing.
func (d *Dispatcher) FindNearbyDrivers(ctx context.Context, req *RideRequest) ([]string, error) {
	return d.findNearbyDrivers(ctx, req)
}

// GetDispatchResult retrieves the dispatch result for a ride by ride ID.
func (d *Dispatcher) GetDispatchResult(ctx context.Context, rideID string) (*DispatchResult, error) {
	ride, err := d.store.GetRide(rideID)
	if err != nil {
		return nil, err
	}

	if ride.State != DriverAssigned && ride.State != DriverAccepted && ride.State != RideCompleted {
		return nil, fmt.Errorf("ride %s is in state %s, no driver assigned yet", rideID, ride.State)
	}

	assignment, err := d.store.GetAssignment("dispatch-" + rideID)
	if err != nil {
		return nil, err
	}

	return &DispatchResult{
		DispatchID: assignment.DispatchID,
		RideID:     ride.RideID,
		DriverID:   assignment.DriverID,
		State:      ride.State,
		Score:      assignment.Score,
		DispatchedAt: &ride.StateEntered,
	}, nil
}

// createTrip creates a minimal trip in fleet-core for the ride request.
func (d *Dispatcher) createTrip(ctx context.Context, req *RideRequest) (string, error) {
	tripReq := map[string]any{
		"title":               "Ride " + req.RideID,
		"description":         "Rider requested ride via dispatch",
		"pickup_lat":          req.Lat,
		"pickup_lng":          req.Lng,
		"rider_id":            req.RiderID,
		"search_radius_km":    req.SearchRadiusKm,
		"surge_multiplier":    req.SurgeMultiplier,
	}
	body, err := json.Marshal(tripReq)
	if err != nil {
		return "", err
	}

	resp, err := d.httpClient.Post(
		fmt.Sprintf("%s/trips", d.fleetCoreURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fleet-core create trip returned %d: %s", resp.StatusCode, string(respBody))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var trip struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &trip); err != nil {
		return "", err
	}
	return trip.ID, nil
}

type idempotencyKeyCtxKey struct{}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtxKey{}, key)
}

func IdempotencyKeyFromContext(ctx context.Context) string {
	key, _ := ctx.Value(idempotencyKeyCtxKey{}).(string)
	return key
}
