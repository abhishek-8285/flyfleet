package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
	ctx := context.Background()
	
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE email = 'admin@flyfleet.com'").Scan(&count)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("User 'admin@flyfleet.com' count: %d\n", count)

	var email string
	err = pool.QueryRow(ctx, "SELECT email FROM users WHERE email = 'admin@flyfleet.com'").Scan(&email)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
	} else {
		fmt.Printf("Found user: %s\n", email)
	}
}
