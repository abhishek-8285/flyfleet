package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/flyfleet/fleet-core/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	seedFile := "migrations/seed_complete_dev_data.sql"
	content, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("Unable to read seed file: %v\n", err)
	}

	fmt.Printf("Applying seed data from %s...\n", seedFile)

	// Execute the entire content as one batch
	_, err = conn.Exec(ctx, string(content))
	if err != nil {
		// If it fails as a batch, try splitting by COMMIT (basic attempt)
		// But usually seed scripts have nested DO blocks, so simple split is risky.
		// We'll just report the error.
		log.Fatalf("Execution failed: %v\n", err)
	}

	fmt.Println("Seed data applied successfully!")

	// Verify vehicles
	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles").Scan(&count)
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
	} else {
		fmt.Printf("Total vehicles in DB now: %d\n", count)
	}

    // List vehicles for delta.io tenant
    tenantID := "550e8400-e29b-41d4-a716-446655440000"
    rows, err := conn.Query(ctx, "SELECT registration_number FROM vehicles WHERE tenant_id = $1", tenantID)
    if err == nil {
        fmt.Printf("Vehicles for tenant %s:\n", tenantID)
        for rows.Next() {
            var reg string
            rows.Scan(&reg)
            fmt.Printf(" - %s\n", reg)
        }
    }
}
