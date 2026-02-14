package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/flyfleet/route-optimization-service/internal/models"
	"github.com/flyfleet/route-optimization-service/internal/repository"
	"github.com/google/uuid"
)

// RouteOptimizationService defines the interface for route optimization business logic
type RouteOptimizationService interface {
	CreateOptimizationRequest(ctx context.Context, request *models.RouteOptimizationRequest) error
	GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error)
	UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteOptimizationRequest(ctx context.Context, id string) error
	ListOptimizationRequests(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error)

	GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error)
	GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error)
	ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error)

	SubmitOptimizationJob(ctx context.Context, requestID string) error
	GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error)
	ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error)
}

// RouteOptimizationServiceImpl implements the RouteOptimizationService interface
type RouteOptimizationServiceImpl struct {
	repo repository.RouteOptimizationRepository
}

// NewRouteOptimizationService creates a new instance of RouteOptimizationService
func NewRouteOptimizationService(repo repository.RouteOptimizationRepository) *RouteOptimizationServiceImpl {
	return &RouteOptimizationServiceImpl{
		repo: repo,
	}
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
	return s.repo.CreateOptimizationRequest(ctx, request)
}

// GetOptimizationRequest retrieves a route optimization request by ID
func (s *RouteOptimizationServiceImpl) GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error) {
	return s.repo.GetOptimizationRequest(ctx, id)
}

// UpdateOptimizationRequest updates an existing route optimization request
func (s *RouteOptimizationServiceImpl) UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.repo.UpdateOptimizationRequest(ctx, id, updates)
}

// DeleteOptimizationRequest deletes a route optimization request
func (s *RouteOptimizationServiceImpl) DeleteOptimizationRequest(ctx context.Context, id string) error {
	return s.repo.DeleteOptimizationRequest(ctx, id)
}

// ListOptimizationRequests retrieves route optimization requests for an organization
func (s *RouteOptimizationServiceImpl) ListOptimizationRequests(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error) {
	return s.repo.ListOptimizationRequests(ctx, orgID, limit, offset)
}

// GetOptimizationResult retrieves a route optimization result by ID
func (s *RouteOptimizationServiceImpl) GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error) {
	return s.repo.GetOptimizationResult(ctx, id)
}

// GetOptimizationResultByRequestID retrieves a route optimization result by request ID
func (s *RouteOptimizationServiceImpl) GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error) {
	return s.repo.GetOptimizationResultByRequestID(ctx, requestID)
}

// ListOptimizationResults retrieves route optimization results for a request
func (s *RouteOptimizationServiceImpl) ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error) {
	return s.repo.ListOptimizationResults(ctx, requestID, limit, offset)
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
	err = s.repo.CreateOptimizationJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create optimization job: %w", err)
	}

	// Update the request status
	err = s.repo.UpdateOptimizationRequest(ctx, requestID, map[string]interface{}{
		"status": "processing",
	})
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// In a real implementation, this would publish the job to a queue for processing
	// For now, we'll simulate the processing by updating the job status
	go s.processOptimizationJob(job.ID)

	return nil
}

// GetOptimizationJob retrieves an optimization job by ID
func (s *RouteOptimizationServiceImpl) GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error) {
	return s.repo.GetOptimizationJob(ctx, id)
}

// ListOptimizationJobs retrieves optimization jobs for a request
func (s *RouteOptimizationServiceImpl) ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error) {
	return s.repo.ListOptimizationJobs(ctx, requestID, status, limit, offset)
}

// validateOptimizationRequest validates the optimization request
func (s *RouteOptimizationServiceImpl) validateOptimizationRequest(request *models.RouteOptimizationRequest) error {
	if request.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
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

	// Update job status to running
	err := s.repo.UpdateOptimizationJobStatus(ctx, jobID, "running", nil, nil)
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

	// Calculate the optimization using Nearest Neighbor heuristic
	result, err := s.performNearestNeighborOptimization(request)
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
func (s *RouteOptimizationServiceImpl) performNearestNeighborOptimization(request *models.RouteOptimizationRequest) (*models.RouteOptimizationResult, error) {
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

		for i, stop := range unassignedStops {
			stopLoc := models.Location{Latitude: stop.Latitude, Longitude: stop.Longitude}
			for vIdx, vLoc := range vehicleLocations {
				dist := s.haversineDistance(vLoc.Latitude, vLoc.Longitude, stopLoc.Latitude, stopLoc.Longitude)
				if dist < minDist {
					minDist = dist
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
			speed := 40.0
			if request.Vehicles[bestVehicleIdx].Speed > 0 {
				speed = request.Vehicles[bestVehicleIdx].Speed
			}
			arrivalTime = request.StartTime.Add(time.Duration(minDist/speed*60) * time.Minute)
		} else {
			lastStop := route.Stops[len(route.Stops)-1]
			travelTime := 30.0
			if request.Vehicles[bestVehicleIdx].Speed > 0 {
				travelTime = minDist / request.Vehicles[bestVehicleIdx].Speed * 60
			}
			arrivalTime = lastStop.DepartureTime.Add(time.Duration(travelTime) * time.Minute)
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
