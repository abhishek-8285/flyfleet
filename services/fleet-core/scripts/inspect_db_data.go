package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	// DB on 5432 (Native)
	url5432 := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	// DB on 5433 (Docker)
	url5433 := "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"

	fmt.Println("🔍 Inspecting DB on port 5432...")
	inspect(ctx, url5432, "5432")

	fmt.Println("\n🔍 Inspecting DB on port 5433...")
	inspect(ctx, url5433, "5433")
}

func inspect(ctx context.Context, url, port string) {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Printf("❌ Could not connect to DB on port %s: %v\n", port, err)
		return
	}
	defer conn.Close(ctx)

	var vehicleCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles").Scan(&vehicleCount)
	if err != nil {
		fmt.Printf("⚠️  vehicles table issue on port %s: %v\n", port, err)
	} else {
		fmt.Printf("🚗 Vehicles count on port %s: %d\n", port, vehicleCount)
	}

	var driverCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM drivers").Scan(&driverCount)
	if err != nil {
		fmt.Printf("⚠️  drivers table issue on port %s: %v\n", port, err)
	} else {
		fmt.Printf("👨 Drivers count on port %s: %d\n", port, driverCount)
	}

	// Print sample data if available
	if vehicleCount > 0 {
		var id string
		var regNo *string
		err = conn.QueryRow(ctx, "SELECT id, registration_no FROM vehicles LIMIT 1").Scan(&id, &regNo)
		if err == nil {
			rNo := "NULL"
			if regNo != nil {
				rNo = *regNo
			}
			fmt.Printf("   💡 Sample Vehicle on %s: ID=%s, RegNo=%s\n", port, id, rNo)
		}
	}
}
