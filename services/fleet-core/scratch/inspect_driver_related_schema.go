package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	tables := []string{"trips", "driver_attendance", "drivers"}

	for _, table := range tables {
		fmt.Printf("\n--- %s TABLE SCHEMA ---\n", table)
		rows, err := conn.Query(context.Background(), fmt.Sprintf(`
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_name = '%s'
			ORDER BY ordinal_position;
		`, table))
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var col, dtype, nullable string
			if err := rows.Scan(&col, &dtype, &nullable); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%-30s | %-20s | %-10s\n", col, dtype, nullable)
		}
		rows.Close()
	}
}
