package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env to get the correct DATABASE_URL
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Warning: Could not load .env file, continuing with environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable is not set. Please ensure it is loaded.")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("🚀 Connected to DB. Applying vehicle schema fixes (org_id -> tenant_id)...")

	// 1. Ensure tenant_id exists
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'vehicles' AND column_name = 'tenant_id'
			) THEN
				ALTER TABLE vehicles ADD COLUMN tenant_id UUID;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Fatalf("❌ Failed to add tenant_id: %v", err)
	}
	fmt.Println("✅ Verified/Added tenant_id column.")

	// 2. Backfill tenant_id if it's null
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'vehicles' AND column_name = 'org_id'
			) THEN
				UPDATE vehicles SET tenant_id = org_id WHERE tenant_id IS NULL;
			END IF;
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'vehicles' AND column_name = 'organization_id'
			) THEN
				UPDATE vehicles SET tenant_id = organization_id WHERE tenant_id IS NULL;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Fatalf("❌ Failed to backfill tenant_id: %v", err)
	}
	fmt.Println("✅ Backfilled tenant_id where possible.")

	// 3. Drop NOT NULL constraint on org_id, then drop the column safely
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			-- Drop org_id
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'vehicles' AND column_name = 'org_id'
			) THEN
				ALTER TABLE vehicles ALTER COLUMN org_id DROP NOT NULL;
				ALTER TABLE vehicles DROP COLUMN org_id CASCADE;
			END IF;

			-- Drop organization_id if present
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'vehicles' AND column_name = 'organization_id'
			) THEN
				ALTER TABLE vehicles ALTER COLUMN organization_id DROP NOT NULL;
				ALTER TABLE vehicles DROP COLUMN organization_id CASCADE;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Fatalf("❌ Failed to drop org_id/organization_id: %v", err)
	}
	fmt.Println("✅ Dropped legacy org_id columns safely.")

	// 4. Ensure tenant_id is NOT NULL (only if all are filled, else leave it for now)
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM vehicles WHERE tenant_id IS NULL) THEN
				ALTER TABLE vehicles ALTER COLUMN tenant_id SET NOT NULL;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Fatalf("❌ Failed to set tenant_id NOT NULL: %v", err)
	}
	fmt.Println("✅ Applied tenant_id constraints.")

	fmt.Println("🎉 Database schema fixed successfully! You can now create vehicles.")
}
