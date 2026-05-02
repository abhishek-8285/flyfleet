package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	url := "postgres://fleet_user:fleet_password@localhost:5432/transport_db?sslmode=disable"

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("❌ Failed to connect to transport_db: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("📋 Listing all tables and row counts in 'transport_db'...")

	rows, err := conn.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`)
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
		}
	}

	for _, table := range tables {
		var count int
		err = conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err == nil {
			fmt.Printf("   - %s: %d rows\n", table, count)
		} else {
			fmt.Printf("   - %s: error querying (%v)\n", table, err)
		}
	}
}
