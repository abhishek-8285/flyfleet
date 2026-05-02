package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	dbs := []string{"fleetflow", "transport_db", "fleet_db"}
	tables := []string{"vehicles", "vehicles_v2", "trucks", "fleet", "drivers"}

	for _, db := range dbs {
		fmt.Printf("🔍 Checking Database: %s\n", db)
		url := fmt.Sprintf("postgres://fleet_user:fleet_password@localhost:5432/%s?sslmode=disable", db)

		conn, err := pgx.Connect(ctx, url)
		if err != nil {
			fmt.Printf("   ❌ Connection failed: %v\n", err)
			continue
		}

		for _, table := range tables {
			var count int
			err = conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			if err == nil {
				fmt.Printf("   ✅ Table '%s': %d rows found.\n", table, count)

				if count > 0 && (table == "vehicles" || table == "vehicles_v2") {
					var id string
					var regNo *string
					err = conn.QueryRow(ctx, fmt.Sprintf("SELECT id, registration_no FROM %s LIMIT 1", table)).Scan(&id, &regNo)
					if err == nil {
						rNo := "NULL"
						if regNo != nil {
							rNo = *regNo
						}
						fmt.Printf("      💡 Sample: ID=%s, RegNo=%s\n", id, rNo)
					}
				}
			}
		}
		conn.Close(ctx)
		fmt.Println()
	}
}
