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

func generateArgon2Hash(password string) (string, error) {
	salt := make([]byte, Argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, uint32(Argon2Iterations), uint32(Argon2Memory), uint8(Argon2Parallelism), uint32(Argon2KeyLength))

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash)

	return encoded, nil
}

func main() {
	hash, err := generateArgon2Hash("password")
	if err != nil {
		log.Fatal(err)
	}

	dbURL := "host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	email := "admin@flyfleet.com"
	_, err = db.Exec("UPDATE users SET password_hash = $1 WHERE email = $2", hash, email)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated %s with hash: %s\n", email, hash)
}
