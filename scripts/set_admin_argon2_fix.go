package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/argon2"
)

const (
	Argon2Iterations  = 3
	Argon2Memory      = 64 * 1024
	Argon2Parallelism = 4
	Argon2SaltLength  = 16
	Argon2KeyLength   = 32
)

func main() {
	password := "password"
	email := "admin@flyfleet.com"
	
	salt := make([]byte, Argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		log.Fatal(err)
	}

	hash := argon2.IDKey([]byte(password), salt, uint32(Argon2Iterations), uint32(Argon2Memory), uint8(Argon2Parallelism), uint32(Argon2KeyLength))

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash)

	db, err := sql.Open("postgres", "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Use ILIKE or Lower to be absolutely sure
	query := "UPDATE users SET password_hash = $1 WHERE LOWER(email) = LOWER($2)"
	res, err := db.Exec(query, encoded, email)
	if err != nil {
		log.Fatal(err)
	}
	count, _ := res.RowsAffected()
	fmt.Printf("Updated %d rows with Argon2id hash for %s.\n", count, email)
	
	if count == 0 {
		fmt.Println("Warning: No rows were updated. Checking if user exists...")
		var dbEmail string
		err := db.QueryRow("SELECT email FROM users WHERE LOWER(email) = LOWER($1)", email).Scan(&dbEmail)
		if err != nil {
			fmt.Printf("Error finding user: %v\n", err)
		} else {
			fmt.Printf("User exists in DB as: %s\n", dbEmail)
		}
	}
}
