package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	fmt.Println("--- Inspecting 'invitations' table ---")
	rows, err := conn.Query(ctx, `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'invitations'
	`)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var name, dtype string
		if err := rows.Scan(&name, &dtype); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Column: %s (%s)\n", name, dtype)
	}

	if !found {
		fmt.Println("Table 'invitations' does not exist.")
	}

	fmt.Println("\n--- Inspecting 'tenants' vs 'organizations' ---")
	var hasTenants, hasOrgs bool
	conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tenants')").Scan(&hasTenants)
	conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organizations')").Scan(&hasOrgs)
	fmt.Printf("Table 'tenants' exists: %v\n", hasTenants)
	fmt.Printf("Table 'organizations' exists: %v\n", hasOrgs)

	if hasOrgs {
		var isView bool
		conn.QueryRow(ctx, "SELECT table_type = 'VIEW' FROM information_schema.tables WHERE table_name = 'organizations'").Scan(&isView)
		fmt.Printf("Table 'organizations' is a VIEW: %v\n", isView)
	}
}
