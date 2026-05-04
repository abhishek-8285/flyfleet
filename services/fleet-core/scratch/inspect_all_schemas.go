package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Using the connection string format from the project
	connStr := "postgres://postgres:postgres@localhost:5432/fleet_db?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	tables := []string{"users", "drivers", "tenants", "vehicles"}

	for _, table := range tables {
		fmt.Printf("\nColumns in %s table:\n", table)
		rows, err := conn.Query(context.Background(), fmt.Sprintf("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position", table))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Query failed for %s: %v\n", table, err)
			continue
		}
		
		for rows.Next() {
			var name, dataType, nullable string
			if err := rows.Scan(&name, &dataType, &nullable); err != nil {
				fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
				continue
			}
			fmt.Printf("- %-20s %-20s (Nullable: %s)\n", name, dataType, nullable)
		}
		rows.Close()
	}
}
