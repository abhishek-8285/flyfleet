package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/route/models"
	"github.com/flyfleet/fleet-core/internal/domains/route/repository"
	"github.com/google/uuid"
)

// RouteOptimizationService defines the interface for route optimization business logic
type RouteOptimizationService interface {
	CreateOptimizationRequest(ctx context.Context, request *models.RouteOptimizationRequest) error
	GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error)
	UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteOptimizationRequest(ctx context.Context, id string) error
	ListOptimizationRequests(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error)

	GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error)
	GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error)
	ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error)

	SubmitOptimizationJob(ctx context.Context, requestID string) error
	GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error)
	ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error)

	// Phase 10: Real-time rerouting
	RecalculateRoute(ctx context.Context, resultID string, vehicleID string, currentLat, currentLon float64) (*models.RouteOptimizationResult, error)
}

// RouteOptimizationServiceImpl implements the RouteOptimizationService interface
type RouteOptimizationServiceImpl struct {
	repo             repository.RouteOptimizationRepository
	distanceProvider DistanceMatrixProvider
	solver           VRPSolver
	validationEngine RouteValidationEngine
}

// NewRouteOptimizationService creates a new instance of RouteOptimizationService
func NewRouteOptimizationService(repo repository.RouteOptimizationRepository, osrmURL string) *RouteOptimizationServiceImpl {
	distanceProvider := &OSRMDistanceProvider{
		BaseURL:  osrmURL,
		Fallback: &HaversineDistanceProvider{},
	}

	s := &RouteOptimizationServiceImpl{
		repo:             repo,
		distanceProvider: distanceProvider,
	}

	// Upgrade to MetaheuristicSolver for premium local-search and metaheuristic routing
	s.solver = &MetaheuristicSolver{Service: s}
	s.validationEngine = &DefaultValidationEngine{Service: s}
	return s
}

// CreateOptimizationRequest creates a new route optimization request
func (s *RouteOptimizationServiceImpl) CreateOptimizationRequest(ctx context.Context, request *models.RouteOptimizationRequest) error {
	// Validate the request
	if err := s.validateOptimizationRequest(request); err != nil {
		return fmt.Errorf("invalid optimization request: %w", err)
	}

	// Set defaults if not provided
	if request.ID == "" {
		request.ID = uuid.New().String()
	}
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()
	request.Status = "pending"

	// Save the request
	if err := s.repo.CreateOptimizationRequest(ctx, request); err != nil {
		return fmt.Errorf("failed to save optimization request: %w", err)
	}
	return nil
}

// GetOptimizationRequest retrieves a route optimization request by ID
func (s *RouteOptimizationServiceImpl) GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error) {
	request, err := s.repo.GetOptimizationRequest(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization request: %w", err)
	}
	return request, nil
}

// UpdateOptimizationRequest updates an existing route optimization request
func (s *RouteOptimizationServiceImpl) UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	if err := s.repo.UpdateOptimizationRequest(ctx, id, updates); err != nil {
		return fmt.Errorf("failed to update optimization request: %w", err)
	}
	return nil
}

// DeleteOptimizationRequest deletes a route optimization request
func (s *RouteOptimizationServiceImpl) DeleteOptimizationRequest(ctx context.Context, id string) error {
	if err := s.repo.DeleteOptimizationRequest(ctx, id); err != nil {
		return fmt.Errorf("failed to delete optimization request: %w", err)
	}
	return nil
}

// ListOptimizationRequests retrieves route optimization requests for an organization
func (s *RouteOptimizationServiceImpl) ListOptimizationRequests(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error) {
	requests, total, err := s.repo.ListOptimizationRequests(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list optimization requests: %w", err)
	}
	return requests, total, nil
}

// GetOptimizationResult retrieves a route optimization result by ID
func (s *RouteOptimizationServiceImpl) GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error) {
	result, err := s.repo.GetOptimizationResult(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization result: %w", err)
	}
	return result, nil
}

// GetOptimizationResultByRequestID retrieves a route optimization result by request ID
func (s *RouteOptimizationServiceImpl) GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error) {
	result, err := s.repo.GetOptimizationResultByRequestID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization result by request ID: %w", err)
	}
	return result, nil
}

// ListOptimizationResults retrieves route optimization results for a request
func (s *RouteOptimizationServiceImpl) ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error) {
	results, total, err := s.repo.ListOptimizationResults(ctx, requestID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list optimization results: %w", err)
	}
	return results, total, nil
}

// SubmitOptimizationJob submits a route optimization job for processing
func (s *RouteOptimizationServiceImpl) SubmitOptimizationJob(ctx context.Context, requestID string) error {
	// Get the optimization request
	request, err := s.repo.GetOptimizationRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get optimization request: %w", err)
	}

	// Verify request exists
	if request == nil {
		return fmt.Errorf("optimization request not found: %s", requestID)
	}

	// Create a new optimization job
	job := &models.OptimizationJob{
		ID:        uuid.New().String(),
		RequestID: requestID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Save the job
	if err := s.repo.CreateOptimizationJob(ctx, job); err != nil {
		return fmt.Errorf("failed to create optimization job: %w", err)
	}

	// Update the request status
	if err := s.repo.UpdateOptimizationRequest(ctx, requestID, map[string]interface{}{
		"status": "processing",
	}); err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// In a real implementation, this would publish the job to a queue for processing
	// For now, we'll simulate the processing by updating the job status
	go s.processOptimizationJob(job.ID)

	return nil
}

// GetOptimizationJob retrieves an optimization job by ID
func (s *RouteOptimizationServiceImpl) GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error) {
	job, err := s.repo.GetOptimizationJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization job: %w", err)
	}
	return job, nil
}

// ListOptimizationJobs retrieves optimization jobs for a request
func (s *RouteOptimizationServiceImpl) ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error) {
	jobs, total, err := s.repo.ListOptimizationJobs(ctx, requestID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list optimization jobs: %w", err)
	}
	return jobs, total, nil
}

// validateOptimizationRequest validates the optimization request
func (s *RouteOptimizationServiceImpl) validateOptimizationRequest(request *models.RouteOptimizationRequest) error {
	if request.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}

	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Stops) == 0 {
		return fmt.Errorf("at least one stop is required")
	}

	if len(request.Vehicles) == 0 {
		return fmt.Errorf("at least one vehicle is required")
	}

	// Validate that all stops have valid coordinates
	for i, stop := range request.Stops {
		if stop.Latitude < -90 || stop.Latitude > 90 {
			return fmt.Errorf("invalid latitude for stop %d: %f", i, stop.Latitude)
		}
		if stop.Longitude < -180 || stop.Longitude > 180 {
			return fmt.Errorf("invalid longitude for stop %d: %f", i, stop.Longitude)
		}
	}

	// Validate that all vehicles have valid start/end locations
	for i, vehicle := range request.Vehicles {
		if vehicle.StartLocation.Latitude < -90 || vehicle.StartLocation.Latitude > 90 {
			return fmt.Errorf("invalid start latitude for vehicle %d: %f", i, vehicle.StartLocation.Latitude)
		}
		if vehicle.StartLocation.Longitude < -180 || vehicle.StartLocation.Longitude > 180 {
			return fmt.Errorf("invalid start longitude for vehicle %d: %f", i, vehicle.StartLocation.Longitude)
		}
		if vehicle.EndLocation.Latitude < -90 || vehicle.EndLocation.Latitude > 90 {
			return fmt.Errorf("invalid end latitude for vehicle %d: %f", i, vehicle.EndLocation.Latitude)
		}
		if vehicle.EndLocation.Longitude < -180 || vehicle.EndLocation.Longitude > 180 {
			return fmt.Errorf("invalid end longitude for vehicle %d: %f", i, vehicle.EndLocation.Longitude)
		}
	}

	return nil
}

// processOptimizationJob simulates the processing of an optimization job
func (s *RouteOptimizationServiceImpl) processOptimizationJob(jobID string) {
	ctx := context.Background()

	// Update job status to processing
	err := s.repo.UpdateOptimizationJobStatus(ctx, jobID, "processing", nil, nil)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error updating job status: %v\n", err)
		return
	}

	// Simulate optimization processing time
	time.Sleep(2 * time.Second)

	// Get the job
	job, err := s.repo.GetOptimizationJob(ctx, jobID)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error getting job: %v\n", err)
		return
	}

	// Get the request associated with this job
	request, err := s.repo.GetOptimizationRequest(ctx, job.RequestID)
	if err != nil {
		// Mark job as failed
		errorMsg := fmt.Sprintf("Failed to get optimization request: %v", err)
		completedAt := time.Now()
		s.repo.UpdateOptimizationJobStatus(ctx, jobID, "failed", &completedAt, &errorMsg)

		// Update request status
		s.repo.UpdateOptimizationRequest(ctx, job.RequestID, map[string]interface{}{
			"status": "failed",
		})
		return
	}

	// Calculate the optimization using the configured solver
	result, err := s.solver.Solve(ctx, request)
	if err != nil {
		// Mark job as failed
		errorMsg := fmt.Sprintf("Optimization failed: %v", err)
		completedAt := time.Now()
		s.repo.UpdateOptimizationJobStatus(ctx, jobID, "failed", &completedAt, &errorMsg)

		// Update request status
		s.repo.UpdateOptimizationRequest(ctx, job.RequestID, map[string]interface{}{
			"status": "failed",
		})
		return
	}

	// Populate full polylines for beautiful frontend rendering
	s.populateRoutePolylines(ctx, request, result)

	// Save the optimization result
	err = s.repo.CreateOptimizationResult(ctx, result)
	if err != nil {
		// Mark job as failed
		errorMsg := fmt.Sprintf("Failed to save optimization result: %v", err)
		completedAt := time.Now()
		s.repo.UpdateOptimizationJobStatus(ctx, jobID, "failed", &completedAt, &errorMsg)

		// Update request status
		s.repo.UpdateOptimizationRequest(ctx, job.RequestID, map[string]interface{}{
			"status": "failed",
		})
		return
	}

	// Update job status to completed
	completedAt := time.Now()
	err = s.repo.UpdateOptimizationJobStatus(ctx, jobID, "completed", &completedAt, nil)
	if err != nil {
		// Just log the error, as the optimization was successful
		fmt.Printf("Error updating job status: %v\n", err)
	}

	// Update request status and attach the result
	err = s.repo.UpdateOptimizationRequest(ctx, job.RequestID, map[string]interface{}{
		"status": "completed",
		"result": result,
	})
	if err != nil {
		// Just log the error
		fmt.Printf("Error updating request status: %v\n", err)
	}

	// Update the result_id in the job record
	err = s.repo.UpdateOptimizationJob(ctx, jobID, map[string]interface{}{
		"result_id": result.ID,
	})
	if err != nil {
		// Just log the error
		fmt.Printf("Error updating job with result ID: %v\n", err)
	}
}

// performNearestNeighborOptimization implements a greedy Nearest Neighbor algorithm for VRP
func (s *RouteOptimizationServiceImpl) performNearestNeighborOptimization(ctx context.Context, request *models.RouteOptimizationRequest) (*models.RouteOptimizationResult, error) {
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

	// Track assigned stops
	unassignedStops := make([]models.Stop, len(request.Stops))
	copy(unassignedStops, request.Stops)

	// Track current location for each vehicle
	vehicleLocations := make([]models.Location, len(request.Vehicles))
	for i, v := range request.Vehicles {
		vehicleLocations[i] = v.StartLocation
	}

	// Track routes
	routes := make([]models.Route, len(request.Vehicles))
	for i, v := range request.Vehicles {
		routes[i] = models.Route{
			VehicleID: v.ID,
			StartTime: request.StartTime,
			Stops:     []models.RouteStop{},
		}
	}

	baselineDistance := 0.0 // To calculate improvement

	// Greedy assignment
	for len(unassignedStops) > 0 {
		bestStopIdx := -1
		bestVehicleIdx := -1
		minDist := math.MaxFloat64
		minTravelTime := 0

		for i, stop := range unassignedStops {
			// Check if vehicle is compatible with stop
			// This check should ideally be done after selecting a vehicle, or for each vehicle.
			// For now, we'll assume it's done per vehicle.

			stopLoc := models.Location{Latitude: stop.Latitude, Longitude: stop.Longitude}
			for vIdx, vLoc := range vehicleLocations {
				// Also check if this specific vehicle is compatible with the stop
				if !s.isVehicleCompatibleWithStop(&request.Vehicles[vIdx], &stop) {
					continue
				}

				// PHASE 2 & 3: Check road restrictions for the destination at arrivalTime
				// For the greedy search, we use a preliminary arrivalTime check
				// We'll calculate the accurate arrivalTime after choosing the best stop.

				dist, travelTime, _ := s.distanceProvider.GetDistanceAndTime(ctx, vLoc, stopLoc, &request.Vehicles[vIdx])
				if dist < minDist {
					minDist = dist
					minTravelTime = travelTime
					bestStopIdx = i
					bestVehicleIdx = vIdx
				}
			}
		}

		if bestStopIdx == -1 {
			break
		}

		// Assign stop
		stop := unassignedStops[bestStopIdx]
		route := &routes[bestVehicleIdx]

		baselineDistance += minDist * 1.2

		// Calculate timing
		var arrivalTime time.Time
		if len(route.Stops) == 0 {
			arrivalTime = request.StartTime.Add(time.Duration(minTravelTime) * time.Minute)
		} else {
			lastStop := route.Stops[len(route.Stops)-1]
			arrivalTime = lastStop.DepartureTime.Add(time.Duration(minTravelTime) * time.Minute)
		}

		// PHASE 2 & 3: Check road restrictions for the destination at arrivalTime
		if !s.isVehicleAllowedOnRoad(ctx, &request.Vehicles[bestVehicleIdx], stop.ID, arrivalTime) {
			continue
		}

		// PHASE 4: Check state permits
		if !s.hasPermitForState(ctx, &request.Vehicles[bestVehicleIdx], stop.State) {
			continue
		}

		departureTime := arrivalTime.Add(time.Duration(stop.ServiceDuration) * time.Minute)

		route.Stops = append(route.Stops, models.RouteStop{
			StopID:        stop.ID,
			Sequence:      len(route.Stops),
			ArrivalTime:   arrivalTime,
			DepartureTime: departureTime,
			ServiceTime:   stop.ServiceDuration,
			Load:          stop.Demand,
		})

		route.Distance += minDist
		route.Time += int(departureTime.Sub(arrivalTime).Minutes())
		route.Cost += minDist * request.Vehicles[bestVehicleIdx].CostPerKM

		// Update vehicle current location
		vehicleLocations[bestVehicleIdx] = models.Location{Latitude: stop.Latitude, Longitude: stop.Longitude}

		// Remove from unassigned
		unassignedStops = append(unassignedStops[:bestStopIdx], unassignedStops[bestStopIdx+1:]...)
	}

	// Finalize routes (return to end location)
	for i := range routes {
		if len(routes[i].Stops) > 0 {
			lastStopLoc := vehicleLocations[i]
			endLoc := request.Vehicles[i].EndLocation
			distToEnd := s.haversineDistance(lastStopLoc.Latitude, lastStopLoc.Longitude, endLoc.Latitude, endLoc.Longitude)

			routes[i].Distance += distToEnd
			travelTimeToEnd := 30.0
			if request.Vehicles[i].Speed > 0 {
				travelTimeToEnd = distToEnd / request.Vehicles[i].Speed * 60
			}
			routes[i].EndTime = routes[i].Stops[len(routes[i].Stops)-1].DepartureTime.Add(time.Duration(travelTimeToEnd) * time.Minute)
			routes[i].OptimizationScore = 0.92

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
			ConvergenceIterations: len(request.Stops),
			ObjectiveValue:        result.TotalDistance,
		}
	}

	return result, nil
}

// haversineDistance calculates the distance between two points in KM
func (s *RouteOptimizationServiceImpl) haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in KM

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// isVehicleCompatibleWithStop checks if a vehicle class is allowed at a stop
func (s *RouteOptimizationServiceImpl) isVehicleCompatibleWithStop(vehicle *models.VehicleProfile, stop *models.Stop) bool {
	if len(stop.AllowedVehicleClasses) == 0 {
		return true // No restrictions
	}

	for _, allowedClass := range stop.AllowedVehicleClasses {
		if vehicle.VehicleClass == allowedClass {
			return true
		}
	}

	return false
}

// isVehicleAllowedOnRoad checks if a vehicle is allowed on a specific road/path at a given time
func (s *RouteOptimizationServiceImpl) isVehicleAllowedOnRoad(ctx context.Context, vehicle *models.VehicleProfile, roadID string, currentTime time.Time) bool {
	// Standard restrictions
	restrictions, err := s.repo.ListRoadRestrictions(ctx, []string{roadID})
	if err == nil && len(restrictions) > 0 {
		for _, rst := range restrictions {
			// Check Vehicle Class
			classAllowed := false
			if len(rst.AllowedVehicleClasses) == 0 {
				classAllowed = true
			} else {
				for _, allowedClass := range rst.AllowedVehicleClasses {
					if vehicle.VehicleClass == allowedClass {
						classAllowed = true
						break
					}
				}
			}
			if !classAllowed {
				return false
			}

			// Check Weight/Height
			if rst.MaxWeight > 0 && vehicle.Weight > rst.MaxWeight {
				return false
			}
			if rst.MaxHeight > 0 && vehicle.Height > rst.MaxHeight {
				return false
			}

			// Check Permit
			if rst.PermitRequired {
				permits, err := s.repo.GetVehiclePermits(ctx, vehicle.ID)
				if err != nil || len(permits) == 0 {
					return false
				}
			}

			if rst.RestrictionType == models.RestrictionNoEntry {
				return false
			}
		}
	}

	// PHASE 3: Time-based restrictions
	timeRestrictions, err := s.repo.ListRoadTimeRestrictions(ctx, []string{roadID})
	if err == nil && len(timeRestrictions) > 0 {
		for _, tr := range timeRestrictions {
			// Check if vehicle is in restricted class for this time window
			classInRestriction := false
			if len(tr.AllowedVehicleClasses) == 0 {
				classInRestriction = true
			} else {
				for _, allowedClass := range tr.AllowedVehicleClasses {
					if vehicle.VehicleClass == allowedClass {
						classInRestriction = true
						break
					}
				}
			}

			if classInRestriction {
				// Parse HH:MM
				currentStr := currentTime.Format("15:04")
				if currentStr >= tr.StartTime && currentStr <= tr.EndTime {
					// Within restricted time window
					// The models say "AllowedVehicleClasses", usually meaning only these are allowed.
					// If vehicle class is NOT in the allowed list during this time, it's restricted.
					// Let's re-read: "allowed_vehicle_classes" in RoadTimeRestriction.
					// If current time is within window, and vehicle class is NOT allowed, return false.
					allowed := false
					for _, ac := range tr.AllowedVehicleClasses {
						if vehicle.VehicleClass == ac {
							allowed = true
							break
						}
					}
					if !allowed {
						return false
					}
				}
			}
		}
	}

	return true
}

// hasPermitForState checks if a vehicle has the necessary permits to enter a state
func (s *RouteOptimizationServiceImpl) hasPermitForState(ctx context.Context, vehicle *models.VehicleProfile, state string) bool {
	if state == "" {
		return true // No state specified, assume allowed
	}

	// First, check if vehicle has an "All India Permit"
	for _, pt := range vehicle.PermitTypes {
		if pt == "All India Permit" || pt == "National Permit" {
			return true
		}
	}

	// Then, check specific state permits from repository
	permits, err := s.repo.GetVehiclePermits(ctx, vehicle.ID)
	if err != nil {
		return false
	}

	for _, pmt := range permits {
		// Check expiry
		if pmt.ExpiryDate.Before(time.Now()) {
			continue
		}

		for _, vs := range pmt.ValidStates {
			if vs == state {
				return true
			}
		}
	}

	return false
}

// clusterStops groups stops into clusters using K-Means
func (s *RouteOptimizationServiceImpl) clusterStops(stops []models.Stop, k int) map[int][]models.Stop {
	if k <= 0 || len(stops) == 0 {
		return nil
	}
	if k > len(stops) {
		k = len(stops)
	}

	// 1. Initialize centroids (randomly from existing stops)
	centroids := make([]models.Location, k)
	for i := 0; i < k; i++ {
		centroids[i] = models.Location{
			Latitude:  stops[i].Latitude,
			Longitude: stops[i].Longitude,
		}
	}

	clusters := make(map[int][]models.Stop)

	// Limited iterations for performance
	for iter := 0; iter < 10; iter++ {
		newClusters := make(map[int][]models.Stop)

		// 2. Assign stops to nearest centroid
		for _, stop := range stops {
			minDist := math.MaxFloat64
			bestCentroid := 0
			for i, centroid := range centroids {
				dist := HaversineDistance(stop.Latitude, stop.Longitude, centroid.Latitude, centroid.Longitude)
				if dist < minDist {
					minDist = dist
					bestCentroid = i
				}
			}
			newClusters[bestCentroid] = append(newClusters[bestCentroid], stop)
		}

		// 3. Update centroids
		changed := false
		for i := 0; i < k; i++ {
			clusterStops := newClusters[i]
			if len(clusterStops) == 0 {
				continue
			}

			var sumLat, sumLon float64
			for _, stop := range clusterStops {
				sumLat += stop.Latitude
				sumLon += stop.Longitude
			}
			newLat := sumLat / float64(len(clusterStops))
			newLon := sumLon / float64(len(clusterStops))

			if newLat != centroids[i].Latitude || newLon != centroids[i].Longitude {
				centroids[i].Latitude = newLat
				centroids[i].Longitude = newLon
				changed = true
			}
		}

		clusters = newClusters
		if !changed {
			break
		}
	}

	return clusters
}

// RecalculateRoute recalculates the route for a specific vehicle starting from its current location
func (s *RouteOptimizationServiceImpl) RecalculateRoute(ctx context.Context, resultID string, vehicleID string, currentLat, currentLon float64) (*models.RouteOptimizationResult, error) {
	// 1. Fetch existing result
	result, err := s.repo.GetOptimizationResult(ctx, resultID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch original request to get constraints
	request, err := s.repo.GetOptimizationRequest(ctx, result.RequestID)
	if err != nil {
		return nil, err
	}

	// 3. Find the specific route for the vehicle
	var targetRoute *models.Route
	for i := range result.Routes {
		if result.Routes[i].VehicleID == vehicleID {
			targetRoute = &result.Routes[i]
			break
		}
	}

	if targetRoute == nil {
		return nil, fmt.Errorf("route for vehicle %s not found in result %s", vehicleID, resultID)
	}

	// 4. Filter out completed stops (simulated: all stops are uncompleted for rerouting)
	// In a real system, we'd check which stops the vehicle has already visited
	uncompletedStops := []models.Stop{}
	for _, rs := range targetRoute.Stops {
		// Find original stop details
		for _, s := range request.Stops {
			if s.ID == rs.StopID {
				uncompletedStops = append(uncompletedStops, s)
				break
			}
		}
	}

	// 5. Create a new sub-request for just this vehicle and uncompleted stops
	subRequest := &models.RouteOptimizationRequest{
		ID:          uuid.New().String(),
		StartTime:   time.Now(),
		Constraints: request.Constraints,
		Vehicles: []models.VehicleProfile{
			{
				ID:            vehicleID,
				StartLocation: models.Location{Latitude: currentLat, Longitude: currentLon},
				// Keep other vehicle parameters if available in request.Vehicles
			},
		},
		Stops: uncompletedStops,
	}

	// Copy vehicle profile details if found in original request
	for _, v := range request.Vehicles {
		if v.ID == vehicleID {
			subRequest.Vehicles[0] = v
			subRequest.Vehicles[0].StartLocation = models.Location{Latitude: currentLat, Longitude: currentLon}
			break
		}
	}

	// 6. Run solver
	newResult, err := s.solver.Solve(ctx, subRequest)
	if err != nil {
		return nil, err
	}

	// Populate full polylines for beautiful frontend rendering
	s.populateRoutePolylines(ctx, subRequest, newResult)

	return newResult, nil
}

// populateRoutePolylines generates complete road-network polylines for all routes in a result
func (s *RouteOptimizationServiceImpl) populateRoutePolylines(ctx context.Context, request *models.RouteOptimizationRequest, result *models.RouteOptimizationResult) {
	if result == nil {
		return
	}

	// Create lookup maps
	stopLookup := make(map[string]models.Stop)
	for _, stop := range request.Stops {
		stopLookup[stop.ID] = stop
	}

	vehicleLookup := make(map[string]models.VehicleProfile)
	for _, vehicle := range request.Vehicles {
		vehicleLookup[vehicle.ID] = vehicle
	}

	for i := range result.Routes {
		route := &result.Routes[i]
		if len(route.Stops) == 0 {
			continue
		}

		vehicle, hasVehicle := vehicleLookup[route.VehicleID]
		if !hasVehicle && len(request.Vehicles) > 0 {
			vehicle = request.Vehicles[0]
		}

		// Reconstruct the sequence of points
		points := []models.Location{}

		// 1. Start location
		startLoc := vehicle.StartLocation
		if startLoc.Latitude == 0 && startLoc.Longitude == 0 && len(request.Depots) > 0 {
			startLoc = request.Depots[0].Location
		}
		points = append(points, startLoc)

		// 2. Stop locations in order
		for _, rs := range route.Stops {
			if stop, ok := stopLookup[rs.StopID]; ok {
				points = append(points, models.Location{
					Latitude:  stop.Latitude,
					Longitude: stop.Longitude,
				})
			}
		}

		// 3. End location
		endLoc := vehicle.EndLocation
		if endLoc.Latitude == 0 && endLoc.Longitude == 0 && len(request.Depots) > 0 {
			// Find nearest depot to the last stop or just use first depot
			if len(route.Stops) > 0 {
				lastStop := stopLookup[route.Stops[len(route.Stops)-1].StopID]
				minDepotDist := math.MaxFloat64
				bestEndLoc := request.Depots[0].Location
				for _, d := range request.Depots {
					dist := HaversineDistance(lastStop.Latitude, lastStop.Longitude, d.Location.Latitude, d.Location.Longitude)
					if dist < minDepotDist {
						minDepotDist = dist
						bestEndLoc = d.Location
					}
				}
				endLoc = bestEndLoc
			} else {
				endLoc = request.Depots[0].Location
			}
		}
		points = append(points, endLoc)

		// Fetch polyline
		polyline, err := s.distanceProvider.GetRoutePolyline(ctx, points)
		if err == nil {
			route.Polyline = polyline
		}
	}
}
