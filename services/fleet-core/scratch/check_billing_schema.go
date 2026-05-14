package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/fleet"
	pool, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	rows, err := pool.Query(context.Background(), `
		SELECT column_name, is_nullable, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'invoice_items'
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Table: invoice_items")
	for rows.Next() {
		var name, nullable, dtype string
		if err := rows.Scan(&name, &nullable, &dtype); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %s: %s (%s)\n", name, dtype, nullable)
	}

	rows2, err := pool.Query(context.Background(), `
		SELECT column_name, is_nullable, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'invoices'
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	fmt.Println("\nTable: invoices")
	for rows2.Next() {
		var name, nullable, dtype string
		if err := rows2.Scan(&name, &nullable, &dtype); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %s: %s (%s)\n", name, dtype, nullable)
	}
}
