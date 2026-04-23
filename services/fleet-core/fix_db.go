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
		-- Fix Vehicle redundancy and NOT NULL constraints
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'license_plate') THEN
			ALTER TABLE vehicles ALTER COLUMN license_plate DROP NOT NULL;
		END IF;

		-- Relax Vehicle status constraint (allow case-insensitive check)
		IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_vehicles_status') THEN
			ALTER TABLE vehicles DROP CONSTRAINT chk_vehicles_status;
		END IF;
		ALTER TABLE vehicles ADD CONSTRAINT chk_vehicles_status CHECK (LOWER(status) IN ('active', 'inactive', 'maintenance', 'retired', 'out_of_service'));

		-- Relax Vehicle fuel_type constraint
		IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_vehicles_fuel_type') THEN
			ALTER TABLE vehicles DROP CONSTRAINT chk_vehicles_fuel_type;
		END IF;
		ALTER TABLE vehicles ADD CONSTRAINT chk_vehicles_fuel_type CHECK (LOWER(fuel_type) IN ('petrol', 'diesel', 'electric', 'hybrid', 'cng', 'lpg'));

		-- Relax Driver status constraint
		IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_drivers_status') THEN
			ALTER TABLE drivers DROP CONSTRAINT chk_drivers_status;
		END IF;
		ALTER TABLE drivers ADD CONSTRAINT chk_drivers_status CHECK (LOWER(status) IN ('active', 'inactive', 'suspended', 'resigned', 'on_leave', 'terminated'));

		-- Sync registration_number if needed
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'registration_number') THEN
			UPDATE vehicles SET registration_number = license_plate WHERE registration_number IS NULL AND license_plate IS NOT NULL;
		END IF;

		-- Add missing columns to drivers table
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'employee_code') THEN
			ALTER TABLE drivers ADD COLUMN employee_code VARCHAR(50);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'aadhaar_number') THEN
			ALTER TABLE drivers ADD COLUMN aadhaar_number VARCHAR(12);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'pan_number') THEN
			ALTER TABLE drivers ADD COLUMN pan_number VARCHAR(10);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'license_issue_date') THEN
			ALTER TABLE drivers ADD COLUMN license_issue_date TIMESTAMPTZ;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'employment_type') THEN
			ALTER TABLE drivers ADD COLUMN employment_type VARCHAR(20) DEFAULT 'FULL_TIME';
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'blood_group') THEN
			ALTER TABLE drivers ADD COLUMN blood_group VARCHAR(5);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'shift_type') THEN
			ALTER TABLE drivers ADD COLUMN shift_type VARCHAR(20);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'behavior_monitoring') THEN
			ALTER TABLE drivers ADD COLUMN behavior_monitoring BOOLEAN DEFAULT TRUE;
		END IF;
	END $$;`



	fmt.Println("Applying schema fix...")
	_, err = db.Exec(sql)
	if err != nil {
		log.Fatalf("FAILED: %v", err)
	}
	fmt.Println("Schema fix applied successfully!")
}
