package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/fleetcore")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	var tenantID string
	var gpsDeviceID *string
	err = conn.QueryRow(ctx, "SELECT tenant_id, gps_device_id FROM vehicles WHERE id = '8c476a51-77a7-4b9f-853e-ce1a1df3939a'").Scan(&tenantID, &gpsDeviceID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("TenantID: %s\n", tenantID)
	if gpsDeviceID != nil {
		fmt.Printf("GPSDeviceID: %s\n", *gpsDeviceID)
	} else {
		fmt.Println("GPSDeviceID: <nil>")
	}
}
