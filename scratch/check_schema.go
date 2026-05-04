package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/fleet_db")
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		return
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'tenants'")
	if err != nil {
		fmt.Printf("Query Error: %v\n", err)
		return
	}
	defer rows.Close()

	var output string
	for rows.Next() {
		var name, dtype string
		rows.Scan(&name, &dtype)
		output += fmt.Sprintf("%s (%s)\n", name, dtype)
	}

	os.WriteFile("schema_dump.txt", []byte(output), 0644)
	fmt.Println("Schema dumped to schema_dump.txt")
}
