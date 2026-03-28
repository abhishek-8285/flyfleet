package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// PEM encode private key
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	// PEM encode public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	})

	// Seed into database
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	kid := fmt.Sprintf("fleet-key-%d", time.Now().Unix())
	query := `
		INSERT INTO signing_keys (kid, private_key_enc, public_key, status, algorithm, created_at, activated_at)
		VALUES ($1, $2, $3, 'active', 'RS256', NOW(), NOW())
	`
	_, err = pool.Exec(ctx, query, kid, string(privPEM), string(pubPEM))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully generated and seeded signing key %s\n", kid)
}
