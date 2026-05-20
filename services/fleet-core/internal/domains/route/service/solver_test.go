package service

import (
	"context"
	
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/route/models"
	"github.com/flyfleet/fleet-core/internal/domains/route/repository"
	"github.com/google/uuid"
)

// MockDistanceProvider implements DistanceMatrixProvider for testing purposes
type MockDistanceProvider struct{}

func (m *MockDistanceProvider) GetDistanceAndTime(ctx context.Context, origin, destination models.Location, vehicle *models.VehicleProfile) (float64, int, error) {
	dist := HaversineDistance(origin.Latitude, origin.Longitude, destination.Latitude, destination.Longitude)
	speed := 40.0
	if vehicle != nil && vehicle.Speed > 0 {
		speed = vehicle.Speed
	}
	travelTime := int((dist / speed) * 60)
	return dist, travelTime, nil
}

func (m *MockDistanceProvider) GetRoutePolyline(ctx context.Context, points []models.Location) ([][]float64, error) {
	result := make([][]float64, len(points))
	for i, pt := range points {
		result[i] = []float64{pt.Latitude, pt.Longitude}
	}
	return result, nil
}

// MockRouteOptimizationRepository implements repository.RouteOptimizationRepository for testing purposes
type MockRouteOptimizationRepository struct{}

var _ repository.RouteOptimizationRepository = (*MockRouteOptimizationRepository)(nil)

func (m *MockRouteOptimizationRepository) CreateOptimizationRequest(ctx context.Context, request *models.RouteOptimizationRequest) error {
	return nil
}
func (m *MockRouteOptimizationRepository) GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error) {
	return nil, nil
}
func (m *MockRouteOptimizationRepository) UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error {
	return nil
}
func (m *MockRouteOptimizationRepository) DeleteOptimizationRequest(ctx context.Context, id string) error {
	return nil
}
func (m *MockRouteOptimizationRepository) ListOptimizationRequests(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error) {
	return nil, 0, nil
}
func (m *MockRouteOptimizationRepository) CreateOptimizationResult(ctx context.Context, result *models.RouteOptimizationResult) error {
	return nil
}
func (m *MockRouteOptimizationRepository) GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error) {
	return nil, nil
}
func (m *MockRouteOptimizationRepository) GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error) {
	return nil, nil
}
func (m *MockRouteOptimizationRepository) UpdateOptimizationResult(ctx context.Context, id string, updates map[string]interface{}) error {
	return nil
}
func (m *MockRouteOptimizationRepository) DeleteOptimizationResult(ctx context.Context, id string) error {
	return nil
}
func (m *MockRouteOptimizationRepository) ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error) {
	return nil, 0, nil
}
func (m *MockRouteOptimizationRepository) CreateOptimizationJob(ctx context.Context, job *models.OptimizationJob) error {
	return nil
}
func (m *MockRouteOptimizationRepository) GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error) {
	return nil, nil
}
func (m *MockRouteOptimizationRepository) UpdateOptimizationJob(ctx context.Context, id string, updates map[string]interface{}) error {
	return nil
}
func (m *MockRouteOptimizationRepository) UpdateOptimizationJobStatus(ctx context.Context, id, status string, completedAt *time.Time, error *string) error {
	return nil
}
func (m *MockRouteOptimizationRepository) ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error) {
	return nil, 0, nil
}
func (m *MockRouteOptimizationRepository) ListRoadRestrictions(ctx context.Context, roadIDs []string) ([]*models.RoadRestriction, error) {
	return nil, nil
}
func (m *MockRouteOptimizationRepository) ListRoadTimeRestrictions(ctx context.Context, roadIDs []string) ([]*models.RoadTimeRestriction, error) {
	return nil, nil
}
func (m *MockRouteOptimizationRepository) GetVehiclePermits(ctx context.Context, vehicleID string) ([]*models.VehiclePermit, error) {
	return nil, nil
}
func (m *MockRouteOptimizationRepository) HealthCheck(ctx context.Context) error {
	return nil
}
func (m *MockRouteOptimizationRepository) Ping(ctx context.Context) error {
	return nil
}

// MockValidationEngine implements RouteValidationEngine for testing purposes
type MockValidationEngine struct{}

func (m *MockValidationEngine) ValidateAssignment(ctx context.Context, vehicle *models.VehicleProfile, stop *models.Stop, currentCapacity int, arrivalTime time.Time) (bool, string) {
	if currentCapacity < stop.Demand {
		return false, "insufficient capacity"
	}
	if stop.TimeWindowEnd != nil && arrivalTime.After(*stop.TimeWindowEnd) {
		return false, "time window missed"
	}
	return true, ""
}

func TestVRPSolversComparison(t *testing.T) {
	ctx := context.Background()

	// 1. Setup Mock Service Context
	serviceImpl := &RouteOptimizationServiceImpl{
		distanceProvider: &MockDistanceProvider{},
		repo:             &MockRouteOptimizationRepository{},
	}
	serviceImpl.validationEngine = &MockValidationEngine{}

	// 2. Define Depot, Vehicles, and Stops (Mumbai, India region)
	depotLoc := models.Location{
		Latitude:  19.0760, // Mumbai Center
		Longitude: 72.8777,
		Address:   "Mumbai Central Hub",
		State:     "Maharashtra",
	}

	startTime := time.Now().Add(2 * time.Hour)
	timeWindowEnd1 := startTime.Add(3 * time.Hour)
	timeWindowEnd2 := startTime.Add(4 * time.Hour)
	timeWindowEnd3 := startTime.Add(5 * time.Hour)

	vehicles := []models.VehicleProfile{
		{
			ID:            "V1",
			Name:          "Tata Ace Gold",
			VehicleClass:  models.VehicleClassPickup,
			Capacity:      150,
			Speed:         40.0,
			CostPerKM:     12.0,
			StartLocation: depotLoc,
			EndLocation:   depotLoc,
			StartTime:     startTime,
			Available:     true,
		},
		{
			ID:            "V2",
			Name:          "Mahindra Bolero Pickup",
			VehicleClass:  models.VehicleClassPickup,
			Capacity:      150,
			Speed:         45.0,
			CostPerKM:     15.0,
			StartLocation: depotLoc,
			EndLocation:   depotLoc,
			StartTime:     startTime,
			Available:     true,
		},
	}

	stops := []models.Stop{
		{
			ID:              "Stop-1-Bandra",
			Address:         "Bandra West, Mumbai",
			State:           "Maharashtra",
			Latitude:        19.0596,
			Longitude:       72.8295,
			Demand:          40,
			ServiceDuration: 15,
			TimeWindowEnd:   &timeWindowEnd1,
		},
		{
			ID:              "Stop-2-Andheri",
			Address:         "Andheri East, Mumbai",
			State:           "Maharashtra",
			Latitude:        19.1136,
			Longitude:       72.8697,
			Demand:          60,
			ServiceDuration: 20,
			TimeWindowEnd:   &timeWindowEnd2,
		},
		{
			ID:              "Stop-3-Thane",
			Address:         "Thane West, Mumbai",
			State:           "Maharashtra",
			Latitude:        19.2183,
			Longitude:       72.9781,
			Demand:          50,
			ServiceDuration: 25,
			TimeWindowEnd:   &timeWindowEnd3,
		},
		{
			ID:              "Stop-4-NaviMumbai",
			Address:         "Vashi, Navi Mumbai",
			State:           "Maharashtra",
			Latitude:        19.0330,
			Longitude:       73.0297,
			Demand:          30,
			ServiceDuration: 15,
		},
		{
			ID:              "Stop-5-Chembur",
			Address:         "Chembur, Mumbai",
			State:           "Maharashtra",
			Latitude:        19.0617,
			Longitude:       72.8970,
			Demand:          40,
			ServiceDuration: 10,
		},
	}

	request := &models.RouteOptimizationRequest{
		ID:        uuid.New().String(),
		TenantID:  uuid.New(),
		Name:      "Mumbai Deliveries Optimization",
		StartTime: startTime,
		Vehicles:  vehicles,
		Stops:     stops,
		Depots: []models.Depot{
			{
				ID:       "Depot-Mumbai",
				Name:     "Mumbai Central Depot",
				Location: depotLoc,
			},
		},
		OptimizationGoals: models.OptimizationGoals{
			MinimizeDistance:   true,
			RespectTimeWindows: true,
		},
	}

	// 3. Instantiate Solvers
	heuristicSolver := &HeuristicSolver{Service: serviceImpl}
	clarkeWrightSolver := &ClarkeWrightSolver{Service: serviceImpl}
	metaheuristicSolver := &MetaheuristicSolver{Service: serviceImpl}

	// 4. Solve using Heuristic (Greedy NN)
	hResult, err := heuristicSolver.Solve(ctx, request)
	if err != nil {
		t.Fatalf("HeuristicSolver failed: %v", err)
	}

	// 5. Solve using Clarke-Wright Savings
	cwResult, err := clarkeWrightSolver.Solve(ctx, request)
	if err != nil {
		t.Fatalf("ClarkeWrightSolver failed: %v", err)
	}

	// 6. Solve using Metaheuristic (2-Opt + Exchange Refined)
	metaResult, err := metaheuristicSolver.Solve(ctx, request)
	if err != nil {
		t.Fatalf("MetaheuristicSolver failed: %v", err)
	}

	// 7. Verify correctness and completeness
	if len(hResult.Routes) == 0 {
		t.Error("HeuristicSolver returned 0 routes")
	}
	if len(cwResult.Routes) == 0 {
		t.Error("ClarkeWrightSolver returned 0 routes")
	}
	if len(metaResult.Routes) == 0 {
		t.Error("MetaheuristicSolver returned 0 routes")
	}

	// Calculate total stops assigned by each solver
	countStops := func(res *models.RouteOptimizationResult) int {
		total := 0
		for _, r := range res.Routes {
			total += len(r.Stops)
		}
		return total
	}

	t.Logf("--- Optimization Comparison Results ---")
	t.Logf("Solver Type       | Assigned Stops | Total Distance (KM) | Est. Cost (INR)")
	t.Logf("-------------------------------------------------------------------------")
	t.Logf("Greedy NN Heur    | %14d | %19.2f | %15.2f", countStops(hResult), hResult.TotalDistance, hResult.TotalCost)
	t.Logf("Clarke-Wright CW  | %14d | %19.2f | %15.2f", countStops(cwResult), cwResult.TotalDistance, cwResult.TotalCost)
	t.Logf("Metaheuristic LSearch| %14d | %19.2f | %15.2f", countStops(metaResult), metaResult.TotalDistance, metaResult.TotalCost)

	// In VRP, Metaheuristic should always be equal or better than the constructive CW baseline it refines!
	if metaResult.TotalDistance > cwResult.TotalDistance {
		t.Errorf("Metaheuristic Solver degraded constructive solution: CW %.2f KM vs Meta %.2f KM", cwResult.TotalDistance, metaResult.TotalDistance)
	} else {
		savings := cwResult.TotalDistance - metaResult.TotalDistance
		pct := (savings / cwResult.TotalDistance) * 100
		t.Logf("🏆 Metaheuristic local search saved %.2f KM (%.2f%%) over Clarke-Wright constructive baseline!", savings, pct)
	}
}

func TestHaversineDistanceAccuracy(t *testing.T) {
	// Mumbai Central to Bandra West is ~11.5 KM
	lat1, lon1 := 19.0760, 72.8777
	lat2, lon2 := 19.0596, 72.8295

	dist := HaversineDistance(lat1, lon1, lat2, lon2)
	t.Logf("Calculated Haversine distance: %.2f KM", dist)

	if dist < 5.0 || dist > 20.0 {
		t.Errorf("Haversine distance out of realistic range: got %.2f KM", dist)
	}
}

func TestMetaheuristicTimeWindowsRespect(t *testing.T) {
	ctx := context.Background()

	serviceImpl := &RouteOptimizationServiceImpl{
		distanceProvider: &MockDistanceProvider{},
		repo:             &MockRouteOptimizationRepository{},
	}

	depotLoc := models.Location{
		Latitude:  19.0760,
		Longitude: 72.8777,
	}

	startTime := time.Now()
	// Extreme case: Stop-1 must be visited immediately, Stop-2 has tight time window
	twEnd1 := startTime.Add(10 * time.Minute) // Infeasible if we detour

	stops := []models.Stop{
		{
			ID:              "Stop-1",
			Latitude:        19.0596,
			Longitude:       72.8295,
			Demand:          10,
			ServiceDuration: 5,
			TimeWindowEnd:   &twEnd1,
		},
		{
			ID:              "Stop-2",
			Latitude:        19.1136,
			Longitude:       72.8697,
			Demand:          10,
			ServiceDuration: 5,
		},
	}

	vehicles := []models.VehicleProfile{
		{
			ID:            "V1",
			Capacity:      100,
			Speed:         60.0, // 1 KM/min
			CostPerKM:     10.0,
			StartLocation: depotLoc,
			EndLocation:   depotLoc,
			StartTime:     startTime,
			Available:     true,
		},
	}

	request := &models.RouteOptimizationRequest{
		ID:        uuid.New().String(),
		TenantID:  uuid.New(),
		Name:      "Tight Time Window Test",
		StartTime: startTime,
		Vehicles:  vehicles,
		Stops:     stops,
		Depots: []models.Depot{
			{
				ID:       "Depot",
				Location: depotLoc,
			},
		},
		OptimizationGoals: models.OptimizationGoals{
			RespectTimeWindows: true,
		},
	}

	metaheuristicSolver := &MetaheuristicSolver{Service: serviceImpl}
	result, err := metaheuristicSolver.Solve(ctx, request)
	if err != nil {
		t.Fatalf("Solver failed: %v", err)
	}

	// Verify that stops in routes are valid or if time windows are violated
	for _, route := range result.Routes {
		currentTime := startTime
		lastLoc := depotLoc
		for _, rs := range route.Stops {
			// Find corresponding stop
			var stop models.Stop
			for _, s := range stops {
				if s.ID == rs.StopID {
					stop = s
					break
				}
			}

			dist := HaversineDistance(lastLoc.Latitude, lastLoc.Longitude, stop.Latitude, stop.Longitude)
			travelTime := int((dist / vehicles[0].Speed) * 60)
			arrivalTime := currentTime.Add(time.Duration(travelTime) * time.Minute)

			if stop.TimeWindowEnd != nil && arrivalTime.After(*stop.TimeWindowEnd) {
				t.Errorf("Time window violated for stop %s: Arrived at %v but deadline was %v", stop.ID, arrivalTime, *stop.TimeWindowEnd)
			}
			currentTime = arrivalTime.Add(time.Duration(stop.ServiceDuration) * time.Minute)
			lastLoc = models.Location{Latitude: stop.Latitude, Longitude: stop.Longitude}
		}
	}
}
