package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	var orphanedCount int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM trips t LEFT JOIN vehicles v ON t.vehicle_id = v.id WHERE v.id IS NULL").Scan(&orphanedCount)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Orphaned trips: %d\n", orphanedCount)

	if orphanedCount > 0 {
		rows, _ := pool.Query(context.Background(), "SELECT t.id, t.vehicle_id FROM trips t LEFT JOIN vehicles v ON t.vehicle_id = v.id WHERE v.id IS NULL LIMIT 10")
		for rows.Next() {
			var id, vid string
			rows.Scan(&id, &vid)
			fmt.Printf("Trip ID: %s, Missing Vehicle ID: %s\n", id, vid)
		}
	}
}
