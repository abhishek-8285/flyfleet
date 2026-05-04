package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/fleet_db")
	if err != nil {
		fmt.Printf("❌ Database Connection Failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// 1. Check Tenants Table
	var taxIdExists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'tax_id')").Scan(&taxIdExists)
	if err != nil {
		fmt.Printf("❌ Schema Check Failed: %v\n", err)
		os.Exit(1)
	}

	if !taxIdExists {
		fmt.Println("❌ ERROR: 'tenants' table is missing the 'tax_id' column.")
		fmt.Println("👉 Rerun the 'Force Reset' command block.")
		os.Exit(1)
	}

	// 2. Check Organizations View
	var isView bool
	err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organizations' AND table_type = 'VIEW')").Scan(&isView)
	
	if !isView {
		fmt.Println("❌ ERROR: 'organizations' is not a VIEW. Legacy compatibility will break.")
		os.Exit(1)
	}

	fmt.Println("✅ DATABASE STABLE: 'tenants' and 'organizations' are correctly configured.")
	fmt.Println("🚀 You can now start the backend with ./restart_backend.sh")
}
