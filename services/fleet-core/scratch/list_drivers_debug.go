package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := "postgres://flyfleet:flyfleet@localhost:5432/flyfleet?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Query as text to see the raw values
	query := `SELECT id::text, tenant_id::text, current_vehicle_id::text FROM drivers`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Inspecting drivers (raw)...")
	for rows.Next() {
		var id, tenantID, vehicleID *string
		err := rows.Scan(&id, &tenantID, &vehicleID)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		
		idVal := "NULL"
		if id != nil { idVal = *id }
		tVal := "NULL"
		if tenantID != nil { tVal = *tenantID }
		vVal := "NULL"
		if vehicleID != nil { vVal = *vehicleID }
		
		fmt.Printf("ID: [%s], Tenant: [%s], Vehicle: [%s]\n", idVal, tVal, vVal)
	}
}
Copied!
