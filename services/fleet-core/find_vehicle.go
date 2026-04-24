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

	var id string
	var tenantID string
	err = conn.QueryRow(context.Background(), "SELECT id, tenant_id FROM vehicles WHERE registration_number = $1", "DL1LAH8783").Scan(&id, &tenantID)
	if err != nil {
		log.Fatalf("Vehicle not found: %v\n", err)
	}

	fmt.Printf("Vehicle DL1LAH8783 -> ID: %s, TenantID: %s\n", id, tenantID)
}
