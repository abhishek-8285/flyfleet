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
		os.WriteFile("db_verification_result.txt", []byte(fmt.Sprintf("FAILURE: %v", err)), 0644)
		return
	}
	defer conn.Close(ctx)

	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
	if err != nil {
		os.WriteFile("db_verification_result.txt", []byte(fmt.Sprintf("QUERY FAILURE: %v", err)), 0644)
		return
	}

	os.WriteFile("db_verification_result.txt", []byte(fmt.Sprintf("SUCCESS: Found %d tables in public schema.", count)), 0644)
}
