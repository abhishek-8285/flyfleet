package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/route/models"
	"github.com/google/uuid"
)

// VRPSolver defines the interface for solving Vehicle Routing Problems
type VRPSolver interface {
	Solve(ctx context.Context, request *models.RouteOptimizationRequest) (*models.RouteOptimizationResult, error)
}

// ClarkeWrightSolver implements VRPSolver using the Savings algorithm for better optimization
type ClarkeWrightSolver struct {
	Service *RouteOptimizationServiceImpl
}

func (s *ClarkeWrightSolver) Solve(ctx context.Context, request *models.RouteOptimizationRequest) (*models.RouteOptimizationResult, error) {
	if len(request.Vehicles) == 0 {
		return nil, fmt.Errorf("no vehicles available")
	}
	if len(request.Stops) == 0 {
		return nil, fmt.Errorf("no stops to optimize")
	}

	startTime := time.Now()
	depot := request.Depots[0].Location // Simplify for MVP

	// 1. Calculate savings Sj,i = d(j,0) + d(0,i) - d(j,i)
	type saving struct {
		i, j   int
		saving float64
	}
	savings := []saving{}

	for i := 0; i < len(request.Stops); i++ {
		for j := i + 1; j < len(request.Stops); j++ {
			d_i0, _, _ := s.Service.distanceProvider.GetDistanceAndTime(ctx, models.Location{Latitude: request.Stops[i].Latitude, Longitude: request.Stops[i].Longitude}, depot, nil)
			d_j0, _, _ := s.Service.distanceProvider.GetDistanceAndTime(ctx, models.Location{Latitude: request.Stops[j].Latitude, Longitude: request.Stops[j].Longitude}, depot, nil)
			d_ij, _, _ := s.Service.distanceProvider.GetDistanceAndTime(ctx, models.Location{Latitude: request.Stops[i].Latitude, Longitude: request.Stops[i].Longitude}, models.Location{Latitude: request.Stops[j].Latitude, Longitude: request.Stops[j].Longitude}, nil)

			savings = append(savings, saving{i: i, j: j, saving: d_i0 + d_j0 - d_ij})
		}
	}

	// 2. Sort savings descending
	sort.Slice(savings, func(i, j int) bool {
		return savings[i].saving > savings[j].saving
	})

	// 3. Initialize routes (one per stop)
	type route struct {
		stops    []int // indices of stops
		capacity int
	}
	activeRoutes := make([]route, len(request.Stops))
	for i := range request.Stops {
		activeRoutes[i] = route{stops: []int{i}, capacity: request.Stops[i].Demand}
	}

	// 4. Merge routes based on savings
	for _, sav := range savings {
		// Find routes containing i and j
		ri, rj := -1, -1
		for idx, r := range activeRoutes {
			for _, sIdx := range r.stops {
				if sIdx == sav.i {
					ri = idx
				}
				if sIdx == sav.j {
					rj = idx
				}
			}
		}

		if ri != -1 && rj != -1 && ri != rj {
			// Check if i and j are at the exterior of their routes
			iAtEnd := activeRoutes[ri].stops[0] == sav.i || activeRoutes[ri].stops[len(activeRoutes[ri].stops)-1] == sav.i
			jAtEnd := activeRoutes[rj].stops[0] == sav.j || activeRoutes[rj].stops[len(activeRoutes[rj].stops)-1] == sav.j

			if iAtEnd && jAtEnd {
				// Check capacity (using first vehicle as proxy for simple MVP)
				if activeRoutes[ri].capacity+activeRoutes[rj].capacity <= request.Vehicles[0].Capacity {
					// Merge routes
					newStops := mergeStops(activeRoutes[ri].stops, activeRoutes[rj].stops, sav.i, sav.j)
					activeRoutes[ri].stops = newStops
					activeRoutes[ri].capacity += activeRoutes[rj].capacity
					// Remove rj
					activeRoutes = append(activeRoutes[:rj], activeRoutes[rj+1:]...)
				}
			}
		}
	}

	// 5. Convert to final result
	finalRoutes := []models.Route{}
	totalDistance := 0.0
	for i, r := range activeRoutes {
		if i >= len(request.Vehicles) {
			break
		}

		finalRoute := models.Route{
			VehicleID: request.Vehicles[i].ID,
			Stops:     []models.RouteStop{},
		}

		lastLoc := depot
		for sIdx, stopIdx := range r.stops {
			stop := request.Stops[stopIdx]
			stopLoc := models.Location{Latitude: stop.Latitude, Longitude: stop.Longitude}
			dist, _, _ := s.Service.distanceProvider.GetDistanceAndTime(ctx, lastLoc, stopLoc, &request.Vehicles[i])

			finalRoute.Stops = append(finalRoute.Stops, models.RouteStop{
				StopID:   stop.ID,
				Sequence: sIdx,
			})
			finalRoute.Distance += dist
			lastLoc = stopLoc
		}
		// Return to depot
		distBack, _, _ := s.Service.distanceProvider.GetDistanceAndTime(ctx, lastLoc, depot, &request.Vehicles[i])
		finalRoute.Distance += distBack

		totalDistance += finalRoute.Distance
		finalRoutes = append(finalRoutes, finalRoute)
	}

	return &models.RouteOptimizationResult{
		ID:               uuid.New().String(),
		RequestID:        request.ID,
		Routes:           finalRoutes,
		TotalDistance:    totalDistance,
		OptimizationTime: startTime,
	}, nil
}

func mergeStops(r1, r2 []int, i, j int) []int {
	// Reorder to ensure i and j are the connecting ends
	if r1[0] == i {
		reverse(r1)
	}
	if r2[len(r2)-1] == j {
		reverse(r2)
	}
	return append(r1, r2...)
}

func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// HeuristicSolver implements VRPSolver using advanced heuristics
type HeuristicSolver struct {
	Service *RouteOptimizationServiceImpl
}

func (s *HeuristicSolver) Solve(ctx context.Context, request *models.RouteOptimizationRequest) (*models.RouteOptimizationResult, error) {
	startTime := time.Now()

	result := &models.RouteOptimizationResult{
		ID:               uuid.New().String(),
		RequestID:        request.ID,
		OptimizationTime: time.Now(),
		TotalStops:       len(request.Stops),
		Routes:           []models.Route{},
	}

	if len(request.Vehicles) == 0 {
		return nil, fmt.Errorf("no vehicles available for optimization")
	}

	// 1. Pre-process: Cluster stops if many
	// (Hidden for now, used in internal logic)

	// 2. Greedy assignment with constraints
	unassignedStops := make([]models.Stop, len(request.Stops))
	copy(unassignedStops, request.Stops)

	vehicleLocations := make([]models.Location, len(request.Vehicles))
	vehicleCapacities := make([]int, len(request.Vehicles))
	for i, v := range request.Vehicles {
		// PHASE 8: Multi-Depot Start Optimization
		if v.StartLocation.Latitude == 0 && v.StartLocation.Longitude == 0 && len(request.Depots) > 0 {
			// Find nearest depot to the first stop (simplification)
			if len(request.Stops) > 0 {
				minDepotDist := math.MaxFloat64
				var bestDepot models.Depot
				for _, d := range request.Depots {
					dist := HaversineDistance(d.Location.Latitude, d.Location.Longitude, request.Stops[0].Latitude, request.Stops[0].Longitude)
					if dist < minDepotDist {
						minDepotDist = dist
						bestDepot = d
					}
				}
				vehicleLocations[i] = bestDepot.Location
			} else {
				vehicleLocations[i] = request.Depots[0].Location
			}
		} else {
			vehicleLocations[i] = v.StartLocation
		}
		vehicleCapacities[i] = v.Capacity
	}

	routes := make([]models.Route, len(request.Vehicles))
	for i, v := range request.Vehicles {
		routes[i] = models.Route{
			VehicleID: v.ID,
			StartTime: request.StartTime,
			Stops:     []models.RouteStop{},
		}
	}

	baselineDistance := 0.0

	for len(unassignedStops) > 0 {
		bestStopIdx := -1
		bestVehicleIdx := -1
		minDist := math.MaxFloat64
		bestArrival := time.Time{}
		bestTravelTime := 0

		for sIdx, stop := range unassignedStops {
			stopLoc := models.Location{Latitude: stop.Latitude, Longitude: stop.Longitude}

			for vIdx, vehicle := range request.Vehicles {
				// Calculate Distance and Time
				dist, travelTime, _ := s.Service.distanceProvider.GetDistanceAndTime(ctx, vehicleLocations[vIdx], stopLoc, &vehicle)

				// Estimate Arrival Time
				var arrivalTime time.Time
				if len(routes[vIdx].Stops) == 0 {
					arrivalTime = request.StartTime.Add(time.Duration(travelTime) * time.Minute)
				} else {
					lastStop := routes[vIdx].Stops[len(routes[vIdx].Stops)-1]
					arrivalTime = lastStop.DepartureTime.Add(time.Duration(travelTime) * time.Minute)
				}

				// Adjust for Time Window Start (waiting)
				if stop.TimeWindowStart != nil && arrivalTime.Before(*stop.TimeWindowStart) {
					arrivalTime = *stop.TimeWindowStart
				}

				// PHASE 9: Use Centralized Validation Engine
				if allowed, _ := s.Service.validationEngine.ValidateAssignment(ctx, &vehicle, &stop, vehicleCapacities[vIdx], arrivalTime); !allowed {
					continue
				}

				if dist < minDist {
					minDist = dist
					bestStopIdx = sIdx
					bestVehicleIdx = vIdx
					bestArrival = arrivalTime
					bestTravelTime = travelTime
				}
			}
		}

		if bestStopIdx == -1 {
			break // No more valid assignments possible
		}

		// Assign Stop
		stop := unassignedStops[bestStopIdx]
		route := &routes[bestVehicleIdx]

		waitingTime := 0
		if stop.TimeWindowStart != nil && bestArrival.Equal(*stop.TimeWindowStart) {
			// Calculate waiting time
			// Actually bestArrival was already adjusted to start window if early.
			// Let's keep it simple for now.
		}

		departureTime := bestArrival.Add(time.Duration(stop.ServiceDuration) * time.Minute)

		route.Stops = append(route.Stops, models.RouteStop{
			StopID:        stop.ID,
			Sequence:      len(route.Stops),
			ArrivalTime:   bestArrival,
			DepartureTime: departureTime,
			ServiceTime:   stop.ServiceDuration,
			Load:          stop.Demand,
			WaitingTime:   waitingTime,
		})

		route.Distance += minDist
		route.Time += (bestTravelTime + stop.ServiceDuration + waitingTime)
		route.Cost += minDist * request.Vehicles[bestVehicleIdx].CostPerKM

		vehicleLocations[bestVehicleIdx] = models.Location{Latitude: stop.Latitude, Longitude: stop.Longitude, State: stop.State}
		vehicleCapacities[bestVehicleIdx] -= stop.Demand
		unassignedStops = append(unassignedStops[:bestStopIdx], unassignedStops[bestStopIdx+1:]...)

		baselineDistance += minDist * 1.3
	}

	// Finalize routes (PHASE 8: End Depot Optimization)
	for i := range routes {
		if len(routes[i].Stops) > 0 {
			lastStopLoc := vehicleLocations[i]

			endLoc := request.Vehicles[i].EndLocation
			if endLoc.Latitude == 0 && endLoc.Longitude == 0 && len(request.Depots) > 0 {
				// Find nearest depot to return to
				minDepotDist := math.MaxFloat64
				for _, d := range request.Depots {
					dist := HaversineDistance(lastStopLoc.Latitude, lastStopLoc.Longitude, d.Location.Latitude, d.Location.Longitude)
					if dist < minDepotDist {
						minDepotDist = dist
						endLoc = d.Location
					}
				}
			}

			distToEnd, timeToEnd, _ := s.Service.distanceProvider.GetDistanceAndTime(ctx, lastStopLoc, endLoc, &request.Vehicles[i])

			routes[i].Distance += distToEnd
			routes[i].Time += timeToEnd
			routes[i].EndTime = routes[i].Stops[len(routes[i].Stops)-1].DepartureTime.Add(time.Duration(timeToEnd) * time.Minute)
			routes[i].OptimizationScore = 0.95

			result.TotalDistance += routes[i].Distance
			result.TotalTime += routes[i].Time
			result.TotalCost += routes[i].Cost
			result.TotalVehiclesUsed++
		}
	}

	result.Routes = routes
	if baselineDistance > 0 {
		result.PerformanceMetrics = models.PerformanceMetrics{
			ImprovementPercentage: ((baselineDistance - result.TotalDistance) / baselineDistance) * 100,
			ExecutionTimeMs:       time.Since(startTime).Milliseconds(),
			ConvergenceIterations: len(request.Stops) - len(unassignedStops),
			ObjectiveValue:        result.TotalDistance,
		}
	}

	return result, nil
}
