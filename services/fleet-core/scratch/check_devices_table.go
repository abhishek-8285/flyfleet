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
	err = conn.QueryRow(ctx, "SELECT tablename FROM pg_catalog.pg_tables WHERE tablename = 'devices'").Scan(&tableName)
	if err != nil {
		fmt.Println("Table 'devices' does NOT exist")
	} else {
		fmt.Println("Table 'devices' exists")
	}
}
