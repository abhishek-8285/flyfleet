package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var tripCount int
	err = db.QueryRow("SELECT count(*) FROM trips WHERE id = '550e8400-e29b-41d4-a716-446655440020'").Scan(&tripCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Trips with ID 550e8400-e29b-41d4-a716-446655440020: %d\n", tripCount)

	var telemetryCount int
	err = db.QueryRow("SELECT count(*) FROM telemetry_points WHERE trip_id = '550e8400-e29b-41d4-a716-446655440020'").Scan(&telemetryCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Telemetry points for trip 550e8400-e29b-41d4-a716-446655440020: %d\n", telemetryCount)
}
