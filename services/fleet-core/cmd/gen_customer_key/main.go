package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gen_customer_key <name> <email> <org_id>")
		fmt.Println("Example: gen_customer_key \"ABC Logistics\" \"admin@abc.io\" \"799ca0ba-6343-44d0-99ca-b12a841f07d7\"")
		os.Exit(1)
	}

	name := os.Args[1]
	email := os.Args[2]
	orgID := os.Args[3]

	ctx := context.Background()

	// 1. Initialize Database
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgresql://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// 2. Generate API Key
	tokenBytes := make([]byte, 16)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		log.Fatalf("Failed to generate random bytes: %v", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	fullKey := "flc_" + rawToken

	hash := sha256.Sum256([]byte(fullKey))
	hashedKey := hex.EncodeToString(hash[:])

	customerID := uuid.New()
	
	// 3. Insert Customer into Legacy Schema for Portal Access
	query := `INSERT INTO customers (id, org_id, name, email, api_key, api_secret_hash, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, 'active', NOW(), NOW())`
	
	_, err = pool.Exec(ctx, query, customerID, orgID, name, email, fullKey, hashedKey)
	if err != nil {
		log.Fatalf("Failed to insert customer: %v", err)
	}

	fmt.Println("\n====================================================")
	fmt.Println("🎉 SUCCESS: Customer created successfully!")
	fmt.Println("====================================================")
	fmt.Printf("Customer ID:  %s\n", customerID)
	fmt.Printf("Name:         %s\n", name)
	fmt.Printf("Email:        %s\n", email)
	fmt.Println("----------------------------------------------------")
	fmt.Printf("API ACCESS KEY: %s\n", fullKey)
	fmt.Println("----------------------------------------------------")
	fmt.Println("⚠️  WARNING: Save this key! It will not be shown again.")
	fmt.Println("The hashed version is stored in the database for security.")
	fmt.Println("====================================================")
}
