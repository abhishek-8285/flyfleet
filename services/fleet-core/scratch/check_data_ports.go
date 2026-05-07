package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ports := []int{5432, 5433}
	user := "fleet_user"
	password := "fleet_password"
	dbName := "fleet_db"

	for _, port := range ports {
		url := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, password, port, dbName)
		conn, err := pgx.Connect(context.Background(), url)
		if err != nil {
			fmt.Printf("Port %d: Connection failed: %v\n", port, err)
			continue
		}
		defer conn.Close(context.Background())

		var count int
		err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM vehicles").Scan(&count)
		if err != nil {
			fmt.Printf("Port %d: Query failed: %v\n", port, err)
			continue
		}

		fmt.Printf("Port %d: %d vehicles found.\n", port, count)
	}
}
