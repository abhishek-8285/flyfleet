package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), `
		SELECT name, default_version, installed_version, comment 
		FROM pg_available_extensions 
		WHERE name LIKE 'postgis%' OR name = 'uuid-ossp';
	`)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	fmt.Println("Available Extensions:")
	for rows.Next() {
		var name, defVer, instVer, comment *string
		if err := rows.Scan(&name, &defVer, &instVer, &comment); err != nil {
			log.Fatal(err)
		}
		inst := "not installed"
		if instVer != nil {
			inst = *instVer
		}
		fmt.Printf("- %s: %s (default: %s) - %s\n", *name, inst, *defVer, *comment)
	}
}
