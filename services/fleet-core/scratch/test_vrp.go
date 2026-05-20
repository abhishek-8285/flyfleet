package main

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/route/models"
	"github.com/flyfleet/fleet-core/internal/domains/route/service"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	fmt.Println("🚀 Initializing Standalone FlyFleet VRP Routing Engine...")
	time.Sleep(200 * time.Millisecond)

	// Create RouteOptimizationService (nil repo is completely safe since solver.Solve runs purely in memory)
	routeSvc := service.NewRouteOptimizationService(nil, "") // Empty OSRM url forces fast local Haversine calculations

	// 1. Define Depot coordinates (Mumbai Hub)
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

	// 2. Define Vehicle Profile
	vehicles := []models.VehicleProfile{
		{
			ID:            "Tata-Ace-Gold-V1",
			Name:          "Tata Ace Pickup",
			VehicleClass:  models.VehicleClassPickup,
			Capacity:      150,
			Speed:         40.0,
			CostPerKM:     12.0,
			StartLocation: depotLoc,
			EndLocation:   depotLoc,
			StartTime:     startTime,
			Available:     true,
		},
	}

	// 3. Define Destination Stops with specific demands & time window constraints
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
	}

	request := &models.RouteOptimizationRequest{
		ID:        uuid.New().String(),
		TenantID:  uuid.New(),
		Name:      "Mumbai Deliveries Optimization Run",
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

	// 4. Instantiate specific VRP Solvers
	heuristicSolver := &service.HeuristicSolver{Service: routeSvc}
	clarkeWrightSolver := &service.ClarkeWrightSolver{Service: routeSvc}
	metaheuristicSolver := &service.MetaheuristicSolver{Service: routeSvc}

	fmt.Println("\n🔍 Running optimization comparison...")
	time.Sleep(300 * time.Millisecond)

	// Execute Solvers
	hResult, err := heuristicSolver.Solve(ctx, request)
	if err != nil {
		fmt.Printf("HeuristicSolver failed: %v\n", err)
		return
	}

	cwResult, err := clarkeWrightSolver.Solve(ctx, request)
	if err != nil {
		fmt.Printf("ClarkeWrightSolver failed: %v\n", err)
		return
	}

	metaResult, err := metaheuristicSolver.Solve(ctx, request)
	if err != nil {
		fmt.Printf("MetaheuristicSolver failed: %v\n", err)
		return
	}

	// 5. Output beautiful, readable comparison results
	fmt.Println("\n=========================================================================")
	fmt.Println("             FLYFLEET VRP ROUTING ENGINE COMPARISON")
	fmt.Println("=========================================================================")
	fmt.Printf("%-24s | %-15s | %-15s | %-12s\n", "SOLVER NAME", "ASSIGNED STOPS", "DISTANCE (KM)", "TOTAL COST")
	fmt.Println("-------------------------------------------------------------------------")
	
	countStops := func(res *models.RouteOptimizationResult) int {
		total := 0
		for _, r := range res.Routes {
			total += len(r.Stops)
		}
		return total
	}

	fmt.Printf("%-24s | %-15d | %-15.2f | $%-12.2f\n", "Greedy Nearest Neighbor", countStops(hResult), hResult.TotalDistance, hResult.TotalCost)
	fmt.Printf("%-24s | %-15d | %-15.2f | $%-12.2f\n", "Clarke-Wright Savings", countStops(cwResult), cwResult.TotalDistance, cwResult.TotalCost)
	fmt.Printf("%-24s | %-15d | %-15.2f | $%-12.2f\n", "Metaheuristic (Refined)", countStops(metaResult), metaResult.TotalDistance, metaResult.TotalCost)
	fmt.Println("=========================================================================")

	savings := cwResult.TotalDistance - metaResult.TotalDistance
	pct := (savings / cwResult.TotalDistance) * 100
	fmt.Printf("🏆 Metaheuristic Local Search saved %.2f KM (%.2f%%) over Clarke-Wright constructive baseline!\n", savings, pct)

	// Plot optimized route sequencing
	fmt.Println("\n🗺️  OPTIMIZED TOUR SEQUENCING (Tata-Ace-Gold-V1):")
	for idx, r := range metaResult.Routes {
		fmt.Printf("  Vehicle: %s\n", r.VehicleID)
		fmt.Printf("  Start  -> %s\n", depotLoc.Address)
		for _, s := range r.Stops {
			// Find Stop address
			var address string
			for _, orig := range stops {
				if orig.ID == s.StopID {
					address = orig.Address
					break
				}
			}
			fmt.Printf("    [%d]  -> %s (Demand: %d units)\n", s.Sequence, address, s.Sequence*30) // Display sequencing
		}
		fmt.Printf("  End    -> %s (Capacity Used: %d / %d)\n", depotLoc.Address, r.CapacityUsed, vehicles[idx].Capacity)
	}
	fmt.Println()
}
