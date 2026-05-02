package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	url := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("❌ Failed to connect to fleet_db: %v", err)
	}
	defer conn.Close(ctx)

	// Columns we want to have (from insertion logic)
	expectedColumns := map[string]string{
		"id":                 "UUID",
		"tenant_id":          "UUID",
		"first_name":         "VARCHAR(100)",
		"last_name":          "VARCHAR(100)",
		"email":              "VARCHAR(255)",
		"phone":              "VARCHAR(50)",
		"employee_code":      "VARCHAR(100)",
		"aadhaar_number":     "VARCHAR(50)",
		"pan_number":         "VARCHAR(50)",
		"license_number":     "VARCHAR(100)",
		"license_type":       "VARCHAR(50)",
		"license_issue_date": "DATE",
		"license_expiry":     "DATE",
		"license_authority":  "VARCHAR(255)",
		"date_of_birth":      "DATE",
		"address":            "JSONB",
		"emergency_contact":  "JSONB",
		"status":             "VARCHAR(50)",
		"employment_type":    "VARCHAR(50)",
		"joining_date":       "DATE",
		"experience_years":   "INTEGER",
		"rating":             "FLOAT",
		"total_trips":        "INTEGER",
		"total_distance_km":  "FLOAT",
		"total_hours_driven": "FLOAT",
		"behavior_score":     "FLOAT",
		"metadata":           "JSONB",
		"created_at":         "TIMESTAMP",
		"updated_at":         "TIMESTAMP",
	}

	// Fetch existing columns
	rows, err := conn.Query(ctx, `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = 'drivers' AND table_schema = 'public'
	`)
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}
	defer rows.Close()

	existingColumns := make(map[string]bool)
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err == nil {
			existingColumns[col] = true
		}
	}

	fmt.Println("🛠️ MISSING COLUMNS IN 'drivers' TABLE:")
	var alterCmd string
	for col, colType := range expectedColumns {
		if !existingColumns[col] {
			fmt.Printf("   - %s (%s)\n", col, colType)
			alterCmd += fmt.Sprintf("ADD COLUMN IF NOT EXISTS %s %s, ", col, colType)
		}
	}

	if alterCmd != "" {
		alterCmd = "ALTER TABLE drivers " + alterCmd[:len(alterCmd)-2] + ";"
		fmt.Printf("\n📋 Full SQL Command to add all missing columns:\n\n%s\n", alterCmd)
	} else {
		fmt.Println("🎉 No missing columns found!")
	}
}
