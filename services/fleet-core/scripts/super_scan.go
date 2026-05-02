package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	defaultURL := "postgres://fleet_user:fleet_password@localhost:5432/postgres?sslmode=disable"

	conn, err := pgx.Connect(ctx, defaultURL)
	if err != nil {
		log.Fatalf("❌ Could not connect to default postgres DB on port 5432: %v", err)
	}
	defer conn.Close(ctx)

	// Get all non-template databases
	rows, err := conn.Query(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false")
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err == nil {
			dbs = append(dbs, dbName)
		}
	}

	fmt.Printf("🔍 Found %d databases. Scanning each for any row data...\n\n", len(dbs))

	for _, db := range dbs {
		scanDB(ctx, db)
	}
}

func scanDB(ctx context.Context, dbName string) {
	url := fmt.Sprintf("postgres://fleet_user:fleet_password@localhost:5432/%s?sslmode=disable", dbName)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		// Try with flyfleet credentials just in case
		url = fmt.Sprintf("postgres://flyfleet_admin:flyfleet_secret@localhost:5432/%s?sslmode=disable", dbName)
		conn, err = pgx.Connect(ctx, url)
		if err != nil {
			return
		}
	}
	defer conn.Close(ctx)

	// Get all public tables in the database
	rows, err := conn.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
		}
	}

	if len(tables) == 0 {
		return
	}

	fmt.Printf("📂 Database: %s\n", dbName)
	foundData := false

	for _, table := range tables {
		var count int
		err = conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err == nil && count > 0 {
			fmt.Printf("   ✅ %s: %d rows\n", table, count)
			foundData = true
		}
	}

	if !foundData {
		fmt.Printf("   ⚪ Contains tables, but all are empty.\n")
	}
	fmt.Println()
}
