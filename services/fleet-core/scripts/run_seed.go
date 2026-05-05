package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	seedFile := "seeds/seed_complete_dev_data.sql"
	content, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("Unable to read seed file: %v\n", err)
	}

	fmt.Printf("Executing seed script: %s\n", seedFile)
	_, err = conn.Exec(ctx, string(content))
	if err != nil {
		log.Fatalf("Failed to execute seed script: %v\n", err)
	}

	fmt.Println("Seed script executed successfully!")
}
