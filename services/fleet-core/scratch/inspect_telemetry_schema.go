package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT column_name, is_nullable, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'telemetry_points'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Columns in telemetry_points:")
	for rows.Next() {
		var name, nullable, dtype string
		if err := rows.Scan(&name, &nullable, &dtype); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %s (%s, %s)\n", name, dtype, nullable)
	}
}
