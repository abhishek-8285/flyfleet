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

	var billingExists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM billing_device_assignments WHERE device_id = '1234567890')").Scan(&billingExists)
	if err != nil {
		log.Fatal(err)
	}

	var trackingExists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM device_registry WHERE device_id = '1234567890')").Scan(&trackingExists)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Billing exists: %v\n", billingExists)
	fmt.Printf("Tracking exists: %v\n", trackingExists)
}
