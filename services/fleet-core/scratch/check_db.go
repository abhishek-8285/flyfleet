package main

import (
	"context"
	"fmt"
	"log"

	"github.com/flyfleet/fleet-core/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_tenants')").Scan(&exists)
	if err != nil {
		log.Fatalf("Check failed: %v\n", err)
	}

	fmt.Printf("Table user_tenants exists: %v\n", exists)

	if exists {
		var constraintCount int
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_name = 'user_tenants'").Scan(&constraintCount)
		fmt.Printf("Constraints on user_tenants: %d\n", constraintCount)
	}
}
