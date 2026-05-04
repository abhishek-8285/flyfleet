package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
)

func main() {
	ctx := context.Background()
	// Connect to default 'postgres' database
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/postgres")
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		return
	}
	defer conn.Close(ctx)

	dbName := "fleet_db_v5"

	// Check if exists
	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		fmt.Printf("Check Error: %v\n", err)
		return
	}

	if !exists {
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			fmt.Printf("Create Error: %v\n", err)
			return
		}
		fmt.Printf("Database %s created successfully.\n", dbName)
	} else {
		fmt.Printf("Database %s already exists.\n", dbName)
	}
}
