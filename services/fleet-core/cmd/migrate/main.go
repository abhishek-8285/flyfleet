package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/flyfleet/fleet-core/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	flag.Parse()
	command := flag.Arg(0)

	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	switch command {
	case "up":
		runMigrations(db)
	case "down":
		fmt.Println("Down migration not supported for custom legacy tool. Please use SQL manually if needed.")
	default:
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down]")
		os.Exit(1)
	}
}

func runMigrations(db *sql.DB) {
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("Failed to list migrations: %v", err)
	}
	sort.Strings(files)

	// Create a simple table to track migrations if it doesn't exist
	_, _ = db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations_legacy (version TEXT PRIMARY KEY, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	_, _ = db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL)")

	for _, file := range files {
		version := filepath.Base(file)

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations_legacy WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check migration status for %s: %v", version, err)
		}

		if exists {
			fmt.Printf("Skipping %s (already applied)\n", version)
			continue
		}

		fmt.Printf("Applying %s... ", version)
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", version, err)
		}

		tx, txErr := db.Begin()
		if txErr != nil {
			log.Fatalf("Failed to start transaction for %s: %v", version, txErr)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			_ = tx.Rollback()
			log.Fatalf("failed to apply %s: %v", version, err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations_legacy (version) VALUES ($1) ON CONFLICT (version) DO NOTHING", version)
		if err != nil {
			_ = tx.Rollback()
			log.Fatalf("failed to record legacy migration %s: %v", version, err)
		}

		if vNum, ok := parseVersionNumber(version); ok {
			_, err = tx.Exec("INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE) ON CONFLICT (version) DO UPDATE SET dirty = FALSE", vNum)
			if err != nil {
				_ = tx.Rollback()
				log.Fatalf("failed to record migration %s in schema_migrations: %v", version, err)
			}
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("failed to commit migration %s: %v", version, err)
		}
		fmt.Println("Done.")
	}
	fmt.Println("All migrations applied successfully.")
}

func parseVersionNumber(name string) (int64, bool) {
	re := regexp.MustCompile(`^(\d+)`)
	m := re.FindStringSubmatch(name)
	if len(m) != 2 {
		return 0, false
	}
	v, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
