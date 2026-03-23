package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Use UPSERT for simplicity
	query := `
		INSERT INTO users (email, password_hash, name, status)
		VALUES ('admin@flyfleet.com', $1, 'Admin User', 'active')
		ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
	`
	res, err := db.Exec(query, string(hash))
	if err != nil {
		log.Fatal(err)
	}
	count, _ := res.RowsAffected()
	fmt.Printf("Updated/Inserted %d rows with bcrypt hash.\n", count)
}
