package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	queries := []string{
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS employee_code VARCHAR(50)",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS aadhaar_number VARCHAR(20)",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS pan_number VARCHAR(20)",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_issue_date DATE",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS employment_type VARCHAR(50)",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS organization_id UUID",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS full_name VARCHAR(255)",
	}

	for _, q := range queries {
		_, err := conn.Exec(context.Background(), q)
		if err != nil {
			log.Printf("Failed to execute query %q: %v\n", q, err)
		} else {
			log.Printf("Successfully executed query %q\n", q)
		}
	}
}
