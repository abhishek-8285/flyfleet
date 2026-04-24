package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Creating telemetry_points table...")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS telemetry_points (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			device_id    VARCHAR(100) NOT NULL,
			lat          DOUBLE PRECISION NOT NULL,
			lng          DOUBLE PRECISION NOT NULL,
			speed        DOUBLE PRECISION DEFAULT 0,
			recorded_at  TIMESTAMP WITH TIME ZONE NOT NULL,
			event_id     UUID NOT NULL,
			tenant_id    UUID NOT NULL,
			received_at  TIMESTAMP WITH TIME ZONE,
			vehicle_id   UUID,
			heading      DOUBLE PRECISION,
			ignition     BOOLEAN,
			fuel_level   DOUBLE PRECISION,
			raw_data     JSONB
		)
	`)
	if err != nil {
		log.Fatalf("Create table failed: %v", err)
	}
	fmt.Println("Table created successfully!")
}
