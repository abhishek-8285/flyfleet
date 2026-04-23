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

	var reg string
	var driverID sql.NullString
	err = db.QueryRow("SELECT registration_number, current_driver_id FROM vehicles WHERE id = '8c476a51-77a7-4b9f-853e-ce1a1df3939a'").Scan(&reg, &driverID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Vehicle: %s, CurrentDriverID: %s\n", reg, driverID.String)
}
