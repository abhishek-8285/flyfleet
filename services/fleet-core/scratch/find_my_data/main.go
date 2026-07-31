package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	// Try to list all databases first
	masterConnStr := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	masterDB, err := sql.Open("postgres", masterConnStr)
	if err != nil {
		// Fallback to flyfleet_admin
		masterConnStr = "postgres://flyfleet_admin:flyfleet_secret@localhost:5432/postgres?sslmode=disable"
		masterDB, err = sql.Open("postgres", masterConnStr)
	}

	dbsToCheck := []string{"flyfleet", "fleet_db", "postgres"}
	if err == nil {
		rows, dbErr := masterDB.Query("SELECT datname FROM pg_database WHERE datistemplate = false")
		if dbErr == nil {
			var dbs []string
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err == nil {
					dbs = append(dbs, name)
				}
			}
			rows.Close()
			if len(dbs) > 0 {
				dbsToCheck = dbs
			}
		}
		masterDB.Close()
	}

	fmt.Println("🔍 Scanning databases for vehicles and drivers...")

	for _, dbName := range dbsToCheck {
		// Try connecting with both users
		connStrs := []string{
			fmt.Sprintf("postgres://flyfleet_admin:flyfleet_secret@localhost:5432/%s?sslmode=disable", dbName),
			fmt.Sprintf("postgres://fleet_user:fleet_password@localhost:5432/%s?sslmode=disable", dbName),
			fmt.Sprintf("postgres://postgres:postgres@localhost:5432/%s?sslmode=disable", dbName),
		}

		var db *sql.DB
		var connected bool
		for _, connStr := range connStrs {
			var openErr error
			db, openErr = sql.Open("postgres", connStr)
			if openErr == nil {
				// test ping
				var temp int
				if pingErr := db.QueryRow("SELECT 1").Scan(&temp); pingErr == nil {
					connected = true
					break
				}
				db.Close()
			}
		}

		if !connected {
			continue
		}
		defer db.Close()

		fmt.Printf("\n📂 Database: %s\n", dbName)

		// Check vehicles
		var vehicleCount int
		err := db.QueryRow("SELECT COUNT(*) FROM vehicles").Scan(&vehicleCount)
		if err != nil {
			fmt.Printf("   ❌ vehicles table: %v\n", err.Error())
		} else {
			fmt.Printf("   ✅ vehicles table: %d records\n", vehicleCount)
		}

		// Check drivers
		var driverCount int
		err = db.QueryRow("SELECT COUNT(*) FROM drivers").Scan(&driverCount)
		if err != nil {
			fmt.Printf("   ❌ drivers table: %v\n", err.Error())
		} else {
			fmt.Printf("   ✅ drivers table: %d records\n", driverCount)
		}
	}
}
