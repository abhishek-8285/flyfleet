package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	connStr := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT id, tenant_id, is_deleted, status, title FROM trips")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	fmt.Println("--- TRIPS IN DATABASE ---")
	found := false
	for rows.Next() {
		var id, tenantID, status, title string
		var isDeleted bool
		err := rows.Scan(&id, &tenantID, &isDeleted, &status, &title)
		if err != nil {
			log.Fatalf("Scan failed: %v\n", err)
		}
		fmt.Printf("ID: %s | Tenant: %s | Deleted: %v | Status: %s | Title: %s\n", id, tenantID, isDeleted, status, title)
		found = true
	}
	if !found {
		fmt.Println("No trips found.")
	}
}
