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

	rows, err := db.Query("SELECT id, first_name, last_name FROM drivers LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Drivers:")
	for rows.Next() {
		var id, fn, ln string
		if err := rows.Scan(&id, &fn, &ln); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %s, Name: %s %s\n", id, fn, ln)
	}
}
