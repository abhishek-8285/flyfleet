package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func checkSchema(connStr string) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to open %s: %v\n", connStr, err)
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'vehicle_documents'
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Printf("Failed to query %s: %v\n", connStr, err)
		return
	}
	defer rows.Close()

	fmt.Printf("Columns in vehicle_documents table (%s):\n", connStr)
	found := false
	for rows.Next() {
		found = true
		var name, dtype string
		if err := rows.Scan(&name, &dtype); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %s: %s\n", name, dtype)
	}
	if !found {
		fmt.Println("  Table NOT FOUND")
	}
}

func main() {
	fmt.Println("Checking database schema...")
	checkSchema("postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	checkSchema("postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable")
}
