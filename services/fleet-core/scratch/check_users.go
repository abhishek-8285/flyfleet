package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT email FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Existing Users:")
	for rows.Next() {
		var email string
		rows.Scan(&email)
		fmt.Println("-", email)
	}
}
