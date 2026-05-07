package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func main() {
	ports := []int{5432, 5433}
	user := "fleet_user"
	password := "fleet_password"
	dbName := "fleet_db"

	for _, port := range ports {
		fmt.Printf("\n--- Port %d ---\n", port)
		url := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, password, port, dbName)
		conn, err := pgx.Connect(context.Background(), url)
		if err != nil {
			fmt.Printf("Connection failed: %v\n", err)
			continue
		}
		defer conn.Close(context.Background())

		rows, err := conn.Query(context.Background(), "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
		if err != nil {
			fmt.Printf("Query tables failed: %v\n", err)
			continue
		}

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				continue
			}

			var count int
			err = conn.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
			if err != nil {
				fmt.Printf("Table %s: %v\n", tableName, err)
				continue
			}

			if count > 0 {
				fmt.Printf("Table %s: %d rows\n", tableName, count)
			}
		}
	}
}
