package service

import (
	"context"
	"fmt"
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

	// Perform the actual optimization (simulated)
	result, err := s.performOptimization(request)
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

// performOptimization performs the actual route optimization algorithm
func (s *RouteOptimizationServiceImpl) performOptimization(request *models.RouteOptimizationRequest) (*models.RouteOptimizationResult, error) {
	// This is a simplified simulation of the optimization process
	// In a real implementation, this would use sophisticated algorithms like:
	// - Genetic algorithms
	// - Ant colony optimization
	// - Simulated annealing
	// - Google OR-Tools
	// - Custom heuristics for VRP (Vehicle Routing Problem)

	result := &models.RouteOptimizationResult{
		ID:              uuid.New().String(),
		RequestID:       request.ID,
		OptimizationTime: time.Now(),
		TotalDistance:   0, // Will be calculated
		TotalTime:       0, // Will be calculated
		TotalCost:       0, // Will be calculated
		TotalVehiclesUsed: 0,
		TotalStops:      len(request.Stops),
		Routes:          []models.Route{},
		PerformanceMetrics: models.PerformanceMetrics{
			ImprovementPercentage: 15.5, // Example improvement
			ExecutionTimeMs:       1200, // Example execution time
			ConvergenceIterations: 250,  // Example iterations
			ObjectiveValue:        420.5,// Example objective value
		},
	}

	// Calculate a basic optimization (this is a simplified version)
	// In reality, this would be much more complex
	totalDistance := 0.0
	totalTime := 0
	totalCost := 0.0
	vehiclesUsed := 0

	// Assign stops to vehicles in a round-robin fashion as a simple approach
	numVehicles := len(request.Vehicles)
	if numVehicles == 0 {
		return nil, fmt.Errorf("no vehicles available for optimization")
	}

	for i, stop := range request.Stops {
		vehicleIdx := i % numVehicles
		vehicle := request.Vehicles[vehicleIdx]

		// Create a route for this vehicle if it doesn't exist yet
		if len(result.Routes) <= vehicleIdx {
			// Extend the slice
			for len(result.Routes) <= vehicleIdx {
				result.Routes = append(result.Routes, models.Route{
					VehicleID: vehicle.ID,
					Stops:     []models.RouteStop{},
					StartTime: request.StartTime,
				})
			}
		}

		// Add the stop to the vehicle's route
		route := &result.Routes[vehicleIdx]
		
		// Calculate arrival time based on travel time from previous stop
		arrivalTime := request.StartTime
		if len(route.Stops) > 0 {
			// In a real implementation, calculate actual travel time between stops
			// For now, we'll just add a fixed amount
			lastStop := route.Stops[len(route.Stops)-1]
			arrivalTime = lastStop.DepartureTime.Add(30 * time.Minute) // 30 min travel time
		}

		departureTime := arrivalTime.Add(time.Duration(stop.ServiceDuration) * time.Minute)

		// Add the stop to the route
		routeStop := models.RouteStop{
			StopID:        stop.ID,
			Sequence:      len(route.Stops),
			ArrivalTime:   arrivalTime,
			DepartureTime: departureTime,
			ServiceTime:   stop.ServiceDuration,
			WaitingTime:   0, // Simplified
			Load:          stop.Demand, // Simplified
		}
		
		route.Stops = append(route.Stops, routeStop)

		// Update route metrics
		route.Distance += 10.0 // Simplified distance calculation
		route.Time += stop.ServiceDuration + 30 // Service time + travel time
		route.Cost += 15.0 // Simplified cost calculation
	}

	// Finalize route calculations
	for i := range result.Routes {
		route := &result.Routes[i]
		if len(route.Stops) > 0 {
			route.EndTime = route.Stops[len(route.Stops)-1].DepartureTime
		}
		route.OptimizationScore = 0.85 // Example score
		
		// Add to totals
		totalDistance += route.Distance
		totalTime += route.Time
		totalCost += route.Cost
		vehiclesUsed++
	}

	result.TotalDistance = totalDistance
	result.TotalTime = totalTime
	result.TotalCost = totalCost
	result.TotalVehiclesUsed = vehiclesUsed

	return result, nil
}