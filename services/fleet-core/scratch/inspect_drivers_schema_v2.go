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

	rows, err := conn.Query(context.Background(), `
		SELECT column_name, is_generated, column_default, is_nullable
		FROM information_schema.columns 
		WHERE table_name = 'drivers'
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Columns in drivers table:")
	for rows.Next() {
		var name, isGenerated, isNullable string
		var colDefault *string
		if err := rows.Scan(&name, &isGenerated, &colDefault, &isNullable); err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}
		defStr := "NULL"
		if colDefault != nil {
			defStr = *colDefault
		}
		fmt.Printf("- %s (Generated: %s, Default: %s, Nullable: %s)\n", name, isGenerated, defStr, isNullable)
	}
}
