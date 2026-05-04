package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/postgres")
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		return
	}
	defer conn.Close(ctx)

	dbName := "fleet_db"

	// Kill connections
	conn.Exec(ctx, fmt.Sprintf("REVOKE CONNECT ON DATABASE %s FROM public", dbName))
	conn.Exec(ctx, fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()", dbName))
	
	// Drop
	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		fmt.Printf("Drop Error: %v\n", err)
	} else {
		fmt.Printf("Database %s dropped successfully.\n", dbName)
	}

	// Verify
	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		fmt.Printf("Verify Error: %v\n", err)
	} else {
		fmt.Printf("Database exists: %v\n", exists)
	}
}
