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

	fmt.Println("--- VEHICLES TABLE SCHEMA ---")
	rows, err := conn.Query(context.Background(), `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = 'vehicles'
		ORDER BY ordinal_position;
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var col, dtype, nullable string
		var def *string
		if err := rows.Scan(&col, &dtype, &nullable, &def); err != nil {
			log.Fatal(err)
		}
		defVal := "NULL"
		if def != nil {
			defVal = *def
		}
		fmt.Printf("%-30s | %-20s | %-10s | %s\n", col, dtype, nullable, defVal)
	}
}
