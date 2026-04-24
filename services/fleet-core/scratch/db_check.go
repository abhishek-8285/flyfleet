package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	url := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	fmt.Println("Checking vehicles in DB...")
	rows, err := conn.Query(context.Background(), "SELECT id, tenant_id, gps_device_id, registration_number FROM vehicles LIMIT 10")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var id, tenantID, gpsDeviceID, reg string
		err := rows.Scan(&id, &tenantID, &gpsDeviceID, &reg)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}
		fmt.Printf("Vehicle ID: %s | Tenant ID: %s | GPS Device ID: %s | Reg: %s\n", id, tenantID, gpsDeviceID, reg)
	}
}
