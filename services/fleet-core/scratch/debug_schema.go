package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	tables := []string{"fuel_logs", "fuel_transactions", "toll_transactions", "expense_reports", "trips", "drivers", "vehicles"}

	for _, table := range tables {
		fmt.Printf("\n--- Table: %s ---\n", table)
		rows, err := pool.Query(context.Background(), `
			SELECT table_schema, column_name, data_type 
			FROM information_schema.columns 
			WHERE table_name = $1 
			ORDER BY table_schema, ordinal_position`, table)
		if err != nil {
			log.Printf("Error querying table %s: %v", table, err)
			continue
		}
		
		found := false
		for rows.Next() {
			var schema, column, dataType string
			if err := rows.Scan(&schema, &column, &dataType); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("[%s] %s (%s)\n", schema, column, dataType)
			found = true
		}
		if !found {
			fmt.Println("Table not found!")
		}
		rows.Close()
	}
}
