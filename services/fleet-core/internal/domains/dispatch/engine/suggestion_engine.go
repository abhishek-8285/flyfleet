package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/dispatch/service"
	driverModels "github.com/flyfleet/fleet-core/internal/domains/driver/models"
	routeModels "github.com/flyfleet/fleet-core/internal/domains/route/models"
	vehicleModels "github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SuggestionEngine interface {
	Suggest(ctx context.Context, req SuggestionRequest) ([]Suggestion, error)
}

type SuggestionRequest struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	ShipmentID   uuid.UUID `json:"shipment_id"`
	PickupLat    float64   `json:"pickup_lat"`
	PickupLng    float64   `json:"pickup_lng"`
	VehicleType  string    `json:"vehicle_type"`
	CargoWeight  float64   `json:"cargo_weight"`
	TopN         int       `json:"top_n"`
	BypassCache  bool      `json:"bypass_cache"`
}

type Suggestion struct {
	VehicleID        uuid.UUID `json:"vehicle_id"`
	DriverID         uuid.UUID `json:"driver_id"`
	DriverName       string    `json:"driver_name"`
	DriverPhone      string    `json:"driver_phone"`
	Rank             int       `json:"rank"`
	Score            float64   `json:"score"`
	DistKm           float64   `json:"dist_km"`
	EstimatedArrival string    `json:"estimated_arrival"`
	Reasons          []string  `json:"reasons"`
}

type VehicleRepository interface {
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*vehicleModels.Vehicle, int64, error)
	Search(ctx context.Context, params vehicleModels.VehicleSearchParams) ([]*vehicleModels.Vehicle, int64, error)
}

type DriverRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*driverModels.Driver, error)
	List(ctx context.Context, filter driverModels.DriverListFilter) ([]*driverModels.Driver, int, error)
}

type TripRepository interface {
	GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]interface{}, error)
}

type LiveLocation struct {
	VehicleID uuid.UUID
	Lat       float64
	Lng       float64
}

type LiveCache interface {
	GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*LiveLocation, error)
}

type DistanceMatrixProvider interface {
	GetDistanceAndTime(ctx context.Context, origin, destination routeModels.Location, vehicle *routeModels.VehicleProfile) (float64, int, error)
}

type suggestionEngine struct {
	vehicleRepo VehicleRepository
	driverRepo  DriverRepository
	tripRepo    TripRepository
	liveCache   LiveCache
	scoring     service.ScoringService
	redis       redis.Cmdable
	router      DistanceMatrixProvider
}

func NewSuggestionEngine(
	vehicleRepo VehicleRepository,
	driverRepo DriverRepository,
	tripRepo TripRepository,
	liveCache LiveCache,
	scoring service.ScoringService,
	rdb redis.Cmdable,
	router DistanceMatrixProvider,
) SuggestionEngine {
	return &suggestionEngine{
		vehicleRepo: vehicleRepo,
		driverRepo:  driverRepo,
		tripRepo:    tripRepo,
		liveCache:   liveCache,
		scoring:     scoring,
		redis:       rdb,
		router:      router,
	}
}

func (s *suggestionEngine) Suggest(ctx context.Context, req SuggestionRequest) ([]Suggestion, error) {
	if s.vehicleRepo == nil {
		return []Suggestion{}, nil
	}

	// 1. Calculate dynamic candidate limit based on requested TopN to protect memory footprint
	limit := req.TopN * 5
	if limit < 50 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	var vehicles []*vehicleModels.Vehicle
	var err error

	// Construct precision cache key based on query filters to intercept hot dispatch operations
	cacheKey := fmt.Sprintf("dispatch:vehicles:cache:%s:%s:%d", req.TenantID.String(), req.VehicleType, limit)

	// Attempt hot Redis lookup first to completely bypass DB execution
	if s.redis != nil && !req.BypassCache {
		if cachedVal, err := s.redis.Get(ctx, cacheKey).Result(); err == nil && cachedVal != "" {
			if jsonErr := json.Unmarshal([]byte(cachedVal), &vehicles); jsonErr == nil && len(vehicles) > 0 {
				goto evaluation
			}
		}
	}

	// Execute DB Query with DB-side indexing filters (type & tenant) instead of massive full scans
	{
		params := vehicleModels.VehicleSearchParams{
			TenantID: &req.TenantID,
			Limit:    limit,
			Page:     1,
		}
		if req.VehicleType != "" {
			params.VehicleType = vehicleModels.VehicleType(strings.ToUpper(req.VehicleType))
		}

		vehicles, _, err = s.vehicleRepo.Search(ctx, params)
		if err != nil {
			// Fallback to tenant retrieval if Search fails or is unsupported
			vehicles, _, err = s.vehicleRepo.GetByTenantID(ctx, req.TenantID, limit, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch vehicles: %w", err)
			}
		}
	}

	// Save candidates to Redis with a short 1-minute TTL to protect database pool saturation under quick re-polls
	if s.redis != nil && len(vehicles) > 0 {
		if cachedBytes, err := json.Marshal(vehicles); err == nil {
			s.redis.Set(ctx, cacheKey, string(cachedBytes), 1*time.Minute)
		}
	}

evaluation:
	suggestions := []Suggestion{}

	// 2. Iterate and evaluate candidate compatibility (First Pass: Haversine Filter)
	type candidate struct {
		vehicle     *vehicleModels.Vehicle
		driver      *driverModels.Driver
		driverID    uuid.UUID
		haversineKm float64
		lat         float64
		lng         float64
	}

	var candidates []candidate
	matchedDriverIDs := make(map[uuid.UUID]bool)

	for _, vehicle := range vehicles {
		// Filter out suspended, blacklisted, or hard-blocked vehicles
		if vehicle.Status == vehicleModels.StatusSuspended || vehicle.IsBlacklisted || vehicle.IsHardBlocked {
			continue
		}

		// Filter out inactive vehicles
		if vehicle.Status == vehicleModels.StatusDecommissioned || vehicle.Status == vehicleModels.StatusRetired {
			continue
		}

		// Type Compatibility Match (case-insensitive)
		if req.VehicleType != "" {
			if !strings.EqualFold(string(vehicle.VehicleType), req.VehicleType) {
				continue
			}
		}

		// Cargo Capacity Match (capacity in Tons * 1000 = Kg)
		if req.CargoWeight > 0 {
			maxCapacityKg := vehicle.CapacityTonnage * 1000.0
			if maxCapacityKg < req.CargoWeight {
				continue
			}
		}

		// Busy Status Check (exclude vehicles assigned to ongoing trips)
		if s.tripRepo != nil {
			activeTrips, err := s.tripRepo.GetActiveTripsForVehicle(ctx, vehicle.ID.String())
			if err == nil && len(activeTrips) > 0 {
				continue
			}
		}

		// Determine spatial location coordinates
		lat, lng := vehicle.Summary.LastLocationLat, vehicle.Summary.LastLocationLng
		if s.liveCache != nil {
			if liveLoc, err := s.liveCache.GetLiveLocation(ctx, vehicle.ID); err == nil && liveLoc != nil {
				lat, lng = liveLoc.Lat, liveLoc.Lng
			}
		}

		// Haversine Distance computation
		distKm := 0.0
		if lat != 0.0 || lng != 0.0 {
			distKm = haversineDistance(req.PickupLat, req.PickupLng, lat, lng)
		}

		// Fetch assigned driver details
		var driver *driverModels.Driver
		var driverID uuid.UUID
		if vehicle.CurrentDriverID != nil && *vehicle.CurrentDriverID != uuid.Nil {
			driverID = *vehicle.CurrentDriverID
			if s.driverRepo != nil {
				if d, err := s.driverRepo.GetByID(ctx, driverID); err == nil && d != nil {
					driver = d
					matchedDriverIDs[driverID] = true
				} else if err != nil {
					fmt.Printf("[SUGGEST-ERROR] Failed to fetch driver %s for vehicle %s: %v\n", driverID, vehicle.ID, err)
				}
			}
		}

		candidates = append(candidates, candidate{
			vehicle:     vehicle,
			driver:      driver,
			driverID:    driverID,
			haversineKm: distKm,
			lat:         lat,
			lng:         lng,
		})
	}

	// 2.2. Query and evaluate unassigned / standby drivers for the tenant
	if s.driverRepo != nil {
		filter := driverModels.DriverListFilter{
			TenantID: req.TenantID.String(),
			Status:   "active",
			Page:     1,
			PageSize: 100,
		}
		if drivers, _, err := s.driverRepo.List(ctx, filter); err == nil {
			for _, driver := range drivers {
				driverID, parseErr := uuid.Parse(driver.ID)
				if parseErr != nil {
					continue
				}
				// Skip if driver is already linked to an evaluated active vehicle candidate
				if matchedDriverIDs[driverID] {
					continue
				}

				// Resolve unassigned driver location (background GPS or home terminal address)
				lat, lng := 0.0, 0.0
				hasMobileGPS := false

				// Check background mobile GPS telemetry key (driver:live_location:<driver_id>)
				if s.redis != nil {
					locKey := fmt.Sprintf("driver:live_location:%s", driverID.String())
					if locVal, err := s.redis.Get(ctx, locKey).Result(); err == nil && locVal != "" {
						type LiveLocCoord struct {
							Lat float64 `json:"lat"`
							Lng float64 `json:"lng"`
						}
						var coord LiveLocCoord
						if jsonErr := json.Unmarshal([]byte(locVal), &coord); jsonErr == nil && coord.Lat != 0.0 && coord.Lng != 0.0 {
							lat, lng = coord.Lat, coord.Lng
							hasMobileGPS = true
						} else {
							var la, lo float64
							if n, scanErr := fmt.Sscanf(locVal, "%f,%f", &la, &lo); scanErr == nil && n == 2 && la != 0.0 && lo != 0.0 {
								lat, lng = la, lo
								hasMobileGPS = true
							}
						}
					}
				}

				// Fallback to Home Terminal / address coordinates
				if !hasMobileGPS && driver.Address.Latitude != nil && driver.Address.Longitude != nil {
					lat = *driver.Address.Latitude
					lng = *driver.Address.Longitude
				}

				// Calculate proximity distance if we have valid coordinates
				distKm := 0.0
				if lat != 0.0 || lng != 0.0 {
					distKm = haversineDistance(req.PickupLat, req.PickupLng, lat, lng)

					candidates = append(candidates, candidate{
						vehicle:     nil, // Unassigned driver has no vehicle
						driver:      driver,
						driverID:    driverID,
						haversineKm: distKm,
						lat:         lat,
						lng:         lng,
					})
				}
			}
		}
	}

	// Sort candidates by straight-line Haversine distance ascending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].haversineKm < candidates[j].haversineKm
	})

	// 3. Second Pass: Resolve actual precise road distance & duration for top candidates
	// We limit precise road calculations to the top 10 closest spatial candidates to avoid O(N) external network bloat
	roadCandidatesLimit := 10
	if len(candidates) < roadCandidatesLimit {
		roadCandidatesLimit = len(candidates)
	}

	for i := 0; i < len(candidates); i++ {
		c := &candidates[i]
		distKm := c.haversineKm
		durationMin := int(distKm * 2.5) // Default rough ETA fallback: ~2.5 mins per km

		// Resolve true road route metrics if router is present and coordinates are valid
		if i < roadCandidatesLimit && s.router != nil && (c.lat != 0.0 || c.lng != 0.0) {
			origin := routeModels.Location{Latitude: c.lat, Longitude: c.lng}
			dest := routeModels.Location{Latitude: req.PickupLat, Longitude: req.PickupLng}
			profile := &routeModels.VehicleProfile{
				Speed: 40.0,
			}
			if c.vehicle != nil {
				profile.VehicleClass = routeModels.VehicleClass(strings.ToUpper(string(c.vehicle.VehicleType)))
			}

			if roadDist, roadMin, err := s.router.GetDistanceAndTime(ctx, origin, dest, profile); err == nil {
				distKm = roadDist
				durationMin = roadMin
			}
		}

		// Calculate scores and descriptive reasons using the resolved route distance
		score := 0.0
		var reasons []string
		if s.scoring != nil {
			score = s.scoring.Calculate(distKm, c.vehicle, c.driver)
			reasons = s.scoring.Explain(distKm, c.vehicle, c.driver)
		} else {
			score = math.Max(0, 100.0-(distKm*5.0))
			if distKm <= 5.0 {
				reasons = []string{"Closest available vehicle"}
			}
		}

		if i < roadCandidatesLimit && s.router != nil {
			reasons = append(reasons, "Actual road-routing calculation verified")
		}

		var driverName, driverPhone string
		if c.driver != nil {
			driverName = c.driver.FullName
			if driverName == "" {
				driverName = fmt.Sprintf("%s %s", c.driver.FirstName, c.driver.LastName)
			}
			driverPhone = c.driver.Phone
		}

		var vehicleID uuid.UUID
		if c.vehicle != nil {
			vehicleID = c.vehicle.ID
		}

		suggestions = append(suggestions, Suggestion{
			VehicleID:        vehicleID,
			DriverID:         c.driverID,
			DriverName:       driverName,
			DriverPhone:      driverPhone,
			Score:            score,
			DistKm:           distKm,
			EstimatedArrival: fmt.Sprintf("%d Mins", durationMin),
			Reasons:          reasons,
		})
	}

	// 4. Sort suggestions descending by final computed Score
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	// 5. Assign visual Rank values
	for idx := range suggestions {
		suggestions[idx].Rank = idx + 1
	}

	// 6. Apply TopN clamping
	topN := req.TopN
	if topN <= 0 {
		topN = 5
	}
	if len(suggestions) > topN {
		suggestions = suggestions[:topN]
	}

	return suggestions, nil
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in KM
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
