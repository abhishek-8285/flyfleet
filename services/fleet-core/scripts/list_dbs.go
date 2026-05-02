package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	// Connect to default 'postgres' DB on port 5432
	url := "postgres://fleet_user:fleet_password@localhost:5432/postgres?sslmode=disable"
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("❌ Could not connect to default postgres DB on port 5432: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("📋 Listing all databases on port 5432...")
	rows, err := conn.Query(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false")
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err == nil {
			fmt.Printf("   - %s\n", dbName)
			// Let's also check vehicles count in this database
			checkCount(ctx, dbName)
		}
	}
}

func checkCount(ctx context.Context, dbName string) {
	url := fmt.Sprintf("postgres://fleet_user:fleet_password@localhost:5432/%s?sslmode=disable", dbName)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return
	}
	defer conn.Close(ctx)

	var vehicleCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles").Scan(&vehicleCount)
	if err == nil && vehicleCount > 0 {
		fmt.Printf("      🚗 Vehicles found: %d!\n", vehicleCount)
		var id string
		var regNo *string
		err = conn.QueryRow(ctx, "SELECT id, registration_no FROM vehicles LIMIT 1").Scan(&id, &regNo)
		if err == nil {
			rNo := "NULL"
			if regNo != nil {
				rNo = *regNo
			}
			fmt.Printf("      💡 Sample Vehicle: ID=%s, RegNo=%s\n", id, rNo)
		}
	}
}
