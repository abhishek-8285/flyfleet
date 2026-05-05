package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	tables := []string{"drivers", "driver_attendance", "trips"}

	for _, table := range tables {
		fmt.Printf("\n--- Table: %s ---\n", table)
		query := fmt.Sprintf(`
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_name = '%s' 
			ORDER BY ordinal_position`, table)

		rows, err := conn.Query(context.Background(), query)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var colName, dataType string
			err := rows.Scan(&colName, &dataType)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%-20s | %s\n", colName, dataType)
		}
		rows.Close()
	}
}
