package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	connStr := "postgres://user:password@localhost:5432/fleet_core?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var exists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicles_legacy')").Scan(&exists)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Table vehicles_legacy exists: %v\n", exists)

	if exists {
		var count int
		err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM vehicles_legacy").Scan(&count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Count failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("vehicles_legacy row count: %d\n", count)
	}

	err = conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicles')").Scan(&exists)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Table vehicles exists: %v\n", exists)
}
