package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://fleet_user:fleet_password@localhost:5432/fleet_db")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	var count int
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM telemetry_points WHERE device_id = $1", "1234567890").Scan(&count)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	fmt.Printf("Simulated rows (Device: 1234567890) in telemetry_points: %d\n", count)
}
