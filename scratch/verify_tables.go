package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fleet_user:fleet_password@localhost:5432/flyfleet?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	rows, err := pool.Query(context.Background(), "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Existing Tables in Database:")
	fmt.Println("---------------------------")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("- %s\n", tableName)
	}
}
