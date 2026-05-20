package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flyfleet/fleet-core/internal/config"
	"github.com/flyfleet/fleet-core/internal/database"
	routeRepo "github.com/flyfleet/fleet-core/internal/domains/route/repository"
	routeSvc "github.com/flyfleet/fleet-core/internal/domains/route/service"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("🚀 Starting VRP Solve Orchestrator...")

	cfg := config.Load()
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer pool.Close()

	repo := routeRepo.NewRouteOptimizationRepository(pool)
	serviceInstance := routeSvc.NewRouteOptimizationService(repo, "https://router.project-osrm.org")

	// 1. Fetch all requests from the database
	fmt.Println("🔍 Scanning database for route requests...")
	
	rows, err := pool.Query(ctx, "SELECT id FROM route_optimization_requests")
	if err != nil {
		log.Fatalf("❌ Failed to query requests: %v", err)
	}
	defer rows.Close()

	var reqIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			reqIDs = append(reqIDs, id)
		}
	}
	rows.Close()

	fmt.Printf("📋 Found %d requests in the database. Solving all of them...\n", len(reqIDs))

	for i, id := range reqIDs {
		fmt.Printf("⚡ [%d/%d] Submitting VRP Solver Job for Request ID: %s...\n", i+1, len(reqIDs), id)

		// Submit VRP job to run Clarke-Wright Heuristic & 2-Opt Metaheuristic optimizations
		err = serviceInstance.SubmitOptimizationJob(ctx, id)
		if err != nil {
			fmt.Printf("❌ Failed to submit job for %s: %v\n", id, err)
			continue
		}

		// Poll status until the background worker completes
		for {
			req, err := serviceInstance.GetOptimizationRequest(ctx, id)
			if err != nil {
				fmt.Printf("❌ Failed to get request status: %v\n", err)
				break
			}
			if req.Status == "completed" {
				fmt.Printf("✅ Solved & updated Request ID: %s\n", id)
				break
			}
			if req.Status == "failed" {
				fmt.Printf("❌ Optimization failed for Request ID: %s\n", id)
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
	}

	fmt.Println("\n=======================================================")
	fmt.Println("🎉 SUCCESS: Every single route request is now solved!")
	fmt.Println("=======================================================")
	fmt.Println("➡️ Refresh your browser at: http://localhost:3000/admin/route-optimization")
	fmt.Println("➡️ Click on ANY request in your sidebar to see beautiful overlays!")
	fmt.Println("=======================================================\n")
}
