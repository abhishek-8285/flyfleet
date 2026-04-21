package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	tables := []string{"invoices", "invoice_items", "customers", "vehicles", "drivers", "trips"}
	backupFile := "flyfleet_sync_data.json"
	fmt.Printf("Starting programmatic sync backup to %s...\n", backupFile)

	fullBackup := make(map[string][]map[string]interface{})

	for _, table := range tables {
		fmt.Printf("Backing up table: %s\n", table)
		rows, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			fmt.Printf("Warning: failed to backup table %s: %v\n", table, err)
			continue
		}

		var tableData []map[string]interface{}
		fieldDescriptions := rows.FieldDescriptions()

		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				log.Fatalf("Error reading row from %s: %v", table, err)
			}

			rowMap := make(map[string]interface{})
			for i, fd := range fieldDescriptions {
				rowMap[string(fd.Name)] = values[i]
			}
			tableData = append(tableData, rowMap)
		}
		fullBackup[table] = tableData
	}

	data, err := json.MarshalIndent(fullBackup, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal backup data: %v", err)
	}

	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		log.Fatalf("Failed to write backup file: %v", err)
	}

	fmt.Printf("Sync backup complete! Generated %s\n", backupFile)
}
