package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	dbURL := "postgres://flyfleet:flyfleet@localhost:5432/flyfleet"
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	rows, err := pool.Query(context.Background(), "SELECT column_name, is_nullable FROM information_schema.columns WHERE table_name = 'telemetry_points'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Columns in telemetry_points:")
	for rows.Next() {
		var name, nullable string
		if err := rows.Scan(&name, &nullable); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %s (nullable: %s)\n", name, nullable)
	}
}
