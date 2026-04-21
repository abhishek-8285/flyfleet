package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sql := `
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'registration_number') THEN
			ALTER TABLE vehicles ADD COLUMN registration_number VARCHAR(50);
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'license_plate') THEN
				UPDATE vehicles SET registration_number = license_plate;
			END IF;
		END IF;
		UPDATE vehicles SET registration_number = 'FIXED-' || id::text WHERE registration_number IS NULL OR registration_number = '';
		ALTER TABLE vehicles ALTER COLUMN registration_number SET NOT NULL;
	END $$;`

	fmt.Println("Applying schema fix...")
	_, err = db.Exec(sql)
	if err != nil {
		log.Fatalf("FAILED: %v", err)
	}
	fmt.Println("Schema fix applied successfully!")
}
