package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/fleetcore")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	var tableName string
	err = conn.QueryRow(ctx, "SELECT tablename FROM pg_catalog.pg_tables WHERE tablename = 'device_registry'").Scan(&tableName)
	if err != nil {
		fmt.Println("Table 'device_registry' does NOT exist")
	} else {
		fmt.Println("Table 'device_registry' exists")
		
		// Check columns
		rows, _ := conn.Query(ctx, "SELECT column_name FROM information_schema.columns WHERE table_name = 'device_registry'")
		fmt.Println("Columns:")
		for rows.Next() {
			var col string
			rows.Scan(&col)
			fmt.Printf("- %s\n", col)
		}
	}
}
