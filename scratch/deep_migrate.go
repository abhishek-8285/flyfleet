package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v4"
)

func main() {
	logFile := "/home/abhishek/Desktop/flyfleet/migration_error.txt"
	ctx := context.Background()
	
	conn, err := pgx.Connect(ctx, "postgres://fleet_user:fleet_password@localhost:5432/fleet_db")
	if err != nil {
		os.WriteFile(logFile, []byte(fmt.Sprintf("Connection Error: %v", err)), 0644)
		return
	}
	defer conn.Close(ctx)

	// Get all migration files
	files, err := os.ReadDir("/home/abhishek/Desktop/flyfleet/services/fleet-core/migrations")
	if err != nil {
		os.WriteFile(logFile, []byte(fmt.Sprintf("ReadDir Error: %v", err)), 0644)
		return
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Ensure migration table exists
	_, err = conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations_legacy (version TEXT PRIMARY KEY, applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW())")
	if err != nil {
		os.WriteFile(logFile, []byte(fmt.Sprintf("Table Init Error: %v", err)), 0644)
		return
	}

	for _, filename := range migrationFiles {
		var exists bool
		err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations_legacy WHERE version=$1)", filename).Scan(&exists)
		if err != nil {
			os.WriteFile(logFile, []byte(fmt.Sprintf("Check Error on %s: %v", filename, err)), 0644)
			return
		}
		if exists {
			continue
		}

		fmt.Printf("Applying %s...\n", filename)
		content, err := os.ReadFile(filepath.Join("/home/abhishek/Desktop/flyfleet/services/fleet-core/migrations", filename))
		if err != nil {
			os.WriteFile(logFile, []byte(fmt.Sprintf("ReadFile Error on %s: %v", filename, err)), 0644)
			return
		}

		_, err = conn.Exec(ctx, string(content))
		if err != nil {
			errorMsg := fmt.Sprintf("MIGRATION FAILED: %s\nERROR: %v", filename, err)
			os.WriteFile(logFile, []byte(errorMsg), 0644)
			fmt.Println(errorMsg)
			return
		}

		_, err = conn.Exec(ctx, "INSERT INTO schema_migrations_legacy (version) VALUES ($1)", filename)
		if err != nil {
			os.WriteFile(logFile, []byte(fmt.Sprintf("Log Record Error on %s: %v", filename, err)), 0644)
			return
		}
	}

	os.WriteFile(logFile, []byte("SUCCESS: All migrations applied perfectly."), 0644)
}
