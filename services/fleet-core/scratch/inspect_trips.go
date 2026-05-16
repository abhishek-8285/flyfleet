package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	connStr := "postgres://postgres:postgres@localhost:5432/fleet?sslmode=disable"
	if envConn := os.Getenv("DATABASE_URL"); envConn != "" {
		connStr = envConn
	}

	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("Columns in 'trips' table:")
	rows, err := pool.Query(ctx, "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'trips'")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, dtype, nullable string
		if err := rows.Scan(&name, &dtype, &nullable); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %s (%s, nullable: %s)\n", name, dtype, nullable)
	}

	fmt.Println("\nRow count in 'trips' table:")
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM trips").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total trips: %d\n", count)

	fmt.Println("\nTrips for tenant '550e8400-e29b-41d4-a716-446655440000':")
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM trips WHERE tenant_id = '550e8400-e29b-41d4-a716-446655440000'").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Trips for tenant: %d\n", count)
}
