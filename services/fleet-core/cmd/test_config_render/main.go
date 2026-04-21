package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
	ctx := context.Background()
	pool, _ := pgxpool.New(ctx, dbURL)
	defer pool.Close()

	tenantID := "00000000-0000-0000-0000-000000000001"

	// 1. Define a Config-Level change (Turn off QR code, change brand color)
	config := map[string]interface{}{
		"show_qrcode":    false,
		"primary_color":  "#e11d48", // Rose Red
		"enabled_blocks": []string{"header", "billto", "items", "totals", "signature"},
	}
	configJSON, _ := json.Marshal(config)

	// 2. Save this config to the DB for this organization
	_, err := pool.Exec(ctx, `
		INSERT INTO billing_settings (tenant_id, render_config)
		VALUES ($1, $2)
		ON CONFLICT (tenant_id) DO UPDATE SET render_config = EXCLUDED.render_config, updated_at = NOW()
	`, tenantID, configJSON)

	if err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Println("✅ Config-Level changes saved to Database.")
	fmt.Printf("Org ID: %s\n", tenantID)
	fmt.Printf("New Config: %s\n", string(configJSON))
}
