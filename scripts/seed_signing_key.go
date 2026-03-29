package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	kid := uuid.New().String()

	db, err := sql.Open("postgres", "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS signing_keys (
			kid text PRIMARY KEY,
			private_key_enc text,
			public_key text,
			status text,
			algorithm text,
			created_at timestamp with time zone,
			activated_at timestamp with time zone,
			expires_at timestamp with time zone
		);
	`)
	if err != nil {
		log.Fatal("Table creation err: ", err)
	}

	_, err = db.Exec(`
		INSERT INTO signing_keys (kid, private_key_enc, public_key, status, algorithm, created_at, activated_at)
		VALUES ($1, $2, $3, 'active', 'RS256', now(), now())
	`, kid, string(privPEM), string(pubPEM))
	if err != nil {
		log.Fatal("Insert err: ", err)
	}

	fmt.Println("Inserted active signing key successfully!")
}
