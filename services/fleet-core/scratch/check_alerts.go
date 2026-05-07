package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT column_name FROM information_schema.columns WHERE table_name = 'alerts'")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Columns in 'alerts' table:")
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("- " + columnName)
	}
}
