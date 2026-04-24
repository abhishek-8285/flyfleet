package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/fleetcore")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	sql := `
		CREATE TABLE IF NOT EXISTS device_registry (
			device_id         VARCHAR(100) PRIMARY KEY,
			vehicle_id        UUID,
			firmware_version  VARCHAR(50),
			sim_iccid         VARCHAR(30),
			carrier           VARCHAR(50),
			last_heartbeat    TIMESTAMPTZ,
			status            VARCHAR(20) NOT NULL DEFAULT 'active',
			certificate_hash  VARCHAR(64),
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	_, err = conn.Exec(ctx, sql)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Println("Successfully ensured device_registry table exists.")
}
