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
	WITH new_device AS (
		INSERT INTO billing_device_assignments (
			id, 
			tenant_id, 
			device_id, 
			product_id, 
			billing_type, 
			status
		) VALUES (
			gen_random_uuid(), 
			'550e8400-e29b-41d4-a716-446655440002', 
			'1234567890',  
			(SELECT id FROM billing_hardware_products WHERE model_code = 'FF-GPS-100' LIMIT 1), 
			'ONE_TIME', 
			'ACTIVE'
		)
		ON CONFLICT (device_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
		RETURNING device_id
	)
	INSERT INTO device_registry (
		device_id, 
		status, 
		created_at, 
		updated_at
	)
	SELECT 
		device_id, 
		'active', 
		NOW(), 
		NOW()
	FROM new_device
	ON CONFLICT (device_id) DO NOTHING;
	`

	_, err = conn.Exec(ctx, sql)
	if err != nil {
		log.Fatalf("Failed to execute SQL: %v", err)
	}

	fmt.Println("Successfully onboarded device 1234567890 to both billing and tracking tables.")
}
