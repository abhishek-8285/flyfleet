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

	rows, err := db.Query("SELECT id, registration_number FROM vehicles LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Vehicles in DB:")
	for rows.Next() {
		var id, reg string
		var regNull sql.NullString
		if err := rows.Scan(&id, &regNull); err != nil {
			log.Fatal(err)
		}
		if regNull.Valid {
			reg = regNull.String
		} else {
			reg = "<NULL>"
		}
		fmt.Printf("ID: %s, Reg: %s\n", id, reg)
	}
}
