package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/fleet_db")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	// Link DL1LAH8783 to device 1234567890
	sql := `UPDATE tracking_v2_vehicles SET gps_device_id = '1234567890', gps_enabled = true WHERE registration_number = 'DL1LAH8783';`
	res, err := conn.Exec(ctx, sql)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Update complete. Rows affected: %d\n", res.RowsAffected())
}
