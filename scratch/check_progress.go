package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/fleet_db")
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		return
	}
	defer conn.Close(ctx)

	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations_legacy").Scan(&count)
	if err != nil {
		// Table might not exist yet if 001 failed
		os.WriteFile("migration_count.txt", []byte("0 (table not found)"), 0644)
		return
	}

	os.WriteFile("migration_count.txt", []byte(fmt.Sprintf("%d", count)), 0644)
}
