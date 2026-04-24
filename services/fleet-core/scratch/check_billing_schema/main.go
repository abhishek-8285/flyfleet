package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connStr := "postgres://flyfleet_admin:flyfleet_secret@localhost:5432/flyfleet?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	var exists bool
	err = pool.QueryRow(context.Background(), "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'billing_device_assignments')").Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		fmt.Println("Table 'billing_device_assignments' does NOT exist.")
	} else {
		fmt.Println("Table 'billing_device_assignments' exists. Checking columns...")
		rows, err := pool.Query(context.Background(), "SELECT column_name FROM information_schema.columns WHERE table_name = 'billing_device_assignments'")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var col string
			if err := rows.Scan(&col); err != nil {
				log.Fatal(err)
			}
			fmt.Println("-", col)
		}
	}

	// Check billing_invoices too
	err = pool.QueryRow(context.Background(), "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'billing_invoices')").Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Table 'billing_invoices' exists: %v\n", exists)

	// Check invoices too
	err = pool.QueryRow(context.Background(), "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'invoices')").Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Table 'invoices' exists: %v\n", exists)
}
