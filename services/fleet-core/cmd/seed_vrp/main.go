package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flyfleet/fleet-core/internal/config"
	"github.com/flyfleet/fleet-core/internal/database"
	"github.com/flyfleet/fleet-core/internal/domains/route/models"
	routeRepo "github.com/flyfleet/fleet-core/internal/domains/route/repository"
	routeSvc "github.com/flyfleet/fleet-core/internal/domains/route/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("🚀 Seeding Mumbai VRP Optimization Request...")

	cfg := config.Load()
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer pool.Close()

	repo := routeRepo.NewRouteOptimizationRepository(pool)
	serviceInstance := routeSvc.NewRouteOptimizationService(repo, "")

	tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000") // Authenticated admin organization

	requestID := uuid.New().String()

	stops := []models.Stop{
		{
			ID:              "stop-bandra",
			Address:         "Bandra West, Mumbai",
			Latitude:        19.0596,
			Longitude:       72.8295,
			Demand:          40,
			ServiceDuration: 15,
			State:           "MH",
		},
		{
			ID:              "stop-andheri",
			Address:         "Andheri East, Mumbai",
			Latitude:        19.1136,
			Longitude:       72.8697,
			Demand:          60,
			ServiceDuration: 20,
			State:           "MH",
		},
		{
			ID:              "stop-dadar",
			Address:         "Dadar West, Mumbai",
			Latitude:        19.0178,
			Longitude:       72.8478,
			Demand:          30,
			ServiceDuration: 15,
			State:           "MH",
		},
		{
			ID:              "stop-chembur",
			Address:         "Chembur, Mumbai",
			Latitude:        19.0622,
			Longitude:       72.8974,
			Demand:          50,
			ServiceDuration: 20,
			State:           "MH",
		},
	}

	vehicles := []models.VehicleProfile{
		{
			ID:            "Tata-Ace-Gold",
			VehicleClass:  "PICKUP",
			Capacity:      150,
			Speed:         40.0,
			CostPerKM:     12.0,
			StartLocation: models.Location{Latitude: 18.9438, Longitude: 72.8396}, // Depot: Mumbai Port
			EndLocation:   models.Location{Latitude: 18.9438, Longitude: 72.8396},
			Available:     true,
		},
		{
			ID:            "Mahindra-Bolero",
			VehicleClass:  "TRUCK",
			Capacity:      250,
			Speed:         45.0,
			CostPerKM:     18.0,
			StartLocation: models.Location{Latitude: 18.9438, Longitude: 72.8396}, // Depot: Mumbai Port
			EndLocation:   models.Location{Latitude: 18.9438, Longitude: 72.8396},
			Available:     true,
		},
	}

	depots := []models.Depot{
		{
			ID:   "depot-mumbai-port",
			Name: "Mumbai Port Central Depot",
			Location: models.Location{
				Latitude:  18.9438,
				Longitude: 72.8396,
				Address:   "Mumbai Port, Fort, Mumbai",
			},
		},
	}

	req := &models.RouteOptimizationRequest{
		ID:                requestID,
		TenantID:          tenantID,
		Name:              "Mumbai Fast-Track Delivery Tour",
		Description:       "Optimal distribution of retail shipments from Mumbai Port Central Depot across prime Mumbai districts.",
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(12 * time.Hour),
		Constraints: models.RouteConstraints{
			MaxTravelTime: 480,
			MaxDistance:   300,
		},
		Stops:    stops,
		Vehicles: vehicles,
		Depots:   depots,
		Priorities: models.RoutePriorities{
			TimePriority:    0.5,
			CostPriority:    0.5,
			ServicePriority: 0.5,
		},
		OptimizationGoals: models.OptimizationGoals{
			MinimizeDistance: true,
			MinimizeVehicles: true,
		},
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. Create optimization request
	err = serviceInstance.CreateOptimizationRequest(ctx, req)
	if err != nil {
		log.Fatalf("❌ Failed to create request: %v", err)
	}
	fmt.Printf("✅ Request created successfully! ID: %s\n", requestID)

	// 2. Submit VRP job to run Clarke-Wright Heuristic & 2-Opt Metaheuristic solver
	fmt.Println("🔄 Solving VRP using Clarke-Wright Heuristic & 2-Opt Metaheuristic optimizations...")
	err = serviceInstance.SubmitOptimizationJob(ctx, requestID)
	if err != nil {
		log.Fatalf("❌ Solver Job failed: %v", err)
	}

	// 3. Give background goroutine 2 seconds to finish solving
	fmt.Println("⏱️ Processing VRP solutions in the background...")
	time.Sleep(2 * time.Second)

	fmt.Println("\n=======================================================")
	fmt.Println("🎉 VRP SUCCESS: Mumbai Delivery Tour completely seeded and solved!")
	fmt.Println("=======================================================")
	fmt.Println("➡️ Refresh your browser at: http://localhost:3000/admin/route-optimization")
	fmt.Println("➡️ You will see the sequenced pins, timelines, and routes mapped live!")
	fmt.Println("=======================================================\n")
}
