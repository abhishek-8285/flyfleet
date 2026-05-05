package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// RunMigrations applies all pending SQL migrations from the migrations directory.
func RunMigrations(dsn string, logger *zap.Logger) error {
	logger.Info("starting database migrations...")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	// Ensure migration tracking tables exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations_legacy (version TEXT PRIMARY KEY, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return fmt.Errorf("failed to create legacy migration table: %w", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL)")
	if err != nil {
		return fmt.Errorf("failed to create schema migration table: %w", err)
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version := filepath.Base(file)

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations_legacy WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}

		if exists {
			logger.Debug("skipping migration (already applied)", zap.String("version", version))
			continue
		}

		logger.Info("applying migration...", zap.String("version", version))
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", version, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction for %s: %w", version, err)
		}

		// Execute migration content (ignoring Down sections)
		sqlContent := string(content)
		downIndex := regexp.MustCompile(`(?i)--\s*\+goose\s+Down`).FindStringIndex(sqlContent)
		if downIndex != nil {
			sqlContent = sqlContent[:downIndex[0]]
		}

		_, err = tx.Exec(sqlContent)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		// Record migration as applied
		_, err = tx.Exec("INSERT INTO schema_migrations_legacy (version) VALUES ($1) ON CONFLICT (version) DO NOTHING", version)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record legacy migration %s: %w", version, err)
		}

		if vNum, ok := parseVersionNumber(version); ok {
			_, err = tx.Exec("INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE) ON CONFLICT (version) DO UPDATE SET dirty = FALSE", vNum)
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("failed to record version %d in schema_migrations: %w", vNum, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		logger.Info("migration applied successfully", zap.String("version", version))
	}

	logger.Info("all database migrations completed successfully")
	return nil
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
