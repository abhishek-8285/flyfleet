package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	ddl := `
	CREATE TABLE IF NOT EXISTS telemetry_points (
		id            BIGSERIAL,
		event_id      UUID         NOT NULL,
		vehicle_id    UUID         NOT NULL,
		device_id     VARCHAR(100) NOT NULL,
		tenant_id     UUID,
		lat           DOUBLE PRECISION NOT NULL,
		lng           DOUBLE PRECISION NOT NULL,
		speed         REAL         NOT NULL DEFAULT 0,
		heading       SMALLINT     NOT NULL DEFAULT 0,
		ignition      BOOLEAN      NOT NULL DEFAULT FALSE,
		fuel_level    REAL,
		odometer      DOUBLE PRECISION,
		event_type    VARCHAR(20)  NOT NULL DEFAULT 'GPS',
		sequence_no   BIGINT       NOT NULL DEFAULT 0,
		recorded_at   TIMESTAMPTZ  NOT NULL,
		received_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	);
	`
	_, err = conn.Exec(ctx, ddl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("telemetry_points table created successfully!")
}
