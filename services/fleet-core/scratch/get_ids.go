package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	var vehicleID, tenantID string
	err = pool.QueryRow(context.Background(), "SELECT id, tenant_id FROM vehicles LIMIT 1").Scan(&vehicleID, &tenantID)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	fmt.Printf("VEHICLE_ID=%s\n", vehicleID)
	fmt.Printf("TENANT_ID=%s\n", tenantID)
}
