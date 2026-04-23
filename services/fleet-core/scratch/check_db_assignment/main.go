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

	rows, err := db.Query(`
		SELECT registration_number, current_driver_id 
		FROM vehicles 
		WHERE current_driver_id IS NOT NULL
		ORDER BY updated_at DESC
		LIMIT 5
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Recently updated vehicles with drivers:")
	for rows.Next() {
		var reg string
		var driverID sql.NullString
		if err := rows.Scan(&reg, &driverID); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Reg: %s, DriverID: %s\n", reg, driverID.String)
	}
}
