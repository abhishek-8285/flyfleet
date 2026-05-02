package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	// Using the credentials and DB name from start_native.sh
	url := "postgres://flyfleet_admin:flyfleet_secret@localhost:5432/flyfleet?sslmode=disable"

	fmt.Println("🔍 Connecting to 'flyfleet' database...")
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("❌ Failed to connect to 'flyfleet' DB: %v", err)
	}
	defer conn.Close(ctx)

	tables := []string{"vehicles", "drivers", "assets", "devices", "gps_assets"}

	for _, table := range tables {
		var count int
		err = conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err == nil {
			fmt.Printf("   ✅ Table '%s': %d rows found.\n", table, count)

			if count > 0 && table == "vehicles" {
				var id string
				var regNo *string
				err = conn.QueryRow(ctx, "SELECT id, registration_no FROM vehicles LIMIT 1").Scan(&id, &regNo)
				if err == nil {
					rNo := "NULL"
					if regNo != nil {
						rNo = *regNo
					}
					fmt.Printf("      💡 Sample Vehicle: ID=%s, RegNo=%s\n", id, rNo)
				}
			}
		} else {
			fmt.Printf("   ⚠️  Table '%s' check failed or table missing.\n", table)
		}
	}
}
