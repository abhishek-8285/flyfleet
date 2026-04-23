package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://flyfleet_admin:flyfleet_secret@localhost:5432/flyfleet?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query column information for vehicles table
	rows, err := db.Query(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'vehicles' 
		AND column_name = 'speed_limit'
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		var name, dtype string
		if err := rows.Scan(&name, &dtype); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Column '%s' exists with type '%s'\n", name, dtype)
	} else {
		fmt.Println("Column 'speed_limit' does NOT exist in 'vehicles' table.")
	}
}
