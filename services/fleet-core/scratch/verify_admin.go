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

	var email string
	err = conn.QueryRow(ctx, "SELECT email FROM users WHERE email = 'admin@delta.io'").Scan(&email)
	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("User NOT found")
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("User FOUND: %s\n", email)
	}
}
