package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AutoSyncSchema ensures that a table exists and contains all columns defined in the struct tags.
// It supports 'db', 'json', 'unique', 'references', and 'index' tags.
func AutoSyncSchema(ctx context.Context, pool *pgxpool.Pool, tableName string, model interface{}, logger *zap.Logger) error {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct or pointer to struct, got %s", t.Kind())
	}

	logger.Info("Starting Premium Auto-Sync", zap.String("table", tableName))

	// 1. Ensure table exists
	_, err := pool.Exec(ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id UUID PRIMARY KEY DEFAULT gen_random_uuid())", tableName))
	if err != nil {
		return fmt.Errorf("failed to ensure table %s: %w", tableName, err)
	}

	// 2. Introspect struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columnName := getColumnName(field)
		if columnName == "" || columnName == "-" || columnName == "id" {
			continue
		}

		columnType := getPostgresType(field.Type)
		if columnType == "" {
			continue
		}

		// Check if column exists
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = $1 AND column_name = $2
			)`
		err = pool.QueryRow(ctx, query, tableName, columnName).Scan(&exists)
		if err != nil {
			logger.Error("column check failed", zap.String("column", columnName), zap.Error(err))
			continue
		}

		if !exists {
			logger.Info("Auto-creating column", zap.String("table", tableName), zap.String("column", columnName))
			alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
			_, err = pool.Exec(ctx, alterQuery)
			if err != nil {
				logger.Error("column creation failed", zap.String("column", columnName), zap.Error(err))
			}
		}

		// 3. Handle Constraints & Indexes
		handleConstraintsAndIndexes(ctx, pool, tableName, columnName, field, logger)
	}

	return nil
}

func handleConstraintsAndIndexes(ctx context.Context, pool *pgxpool.Pool, tableName, columnName string, field reflect.StructField, logger *zap.Logger) {
	// Unique Constraint
	if field.Tag.Get("unique") == "true" {
		constraintName := fmt.Sprintf("uq_%s_%s", tableName, columnName)
		query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)", tableName, constraintName, columnName)
		_, _ = pool.Exec(ctx, query) // Ignore if exists
	}

	// Foreign Key (references)
	ref := field.Tag.Get("references")
	if ref != "" {
		constraintName := fmt.Sprintf("fk_%s_%s", tableName, columnName)
		query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(id)", tableName, constraintName, columnName, ref)
		_, _ = pool.Exec(ctx, query) // Ignore if exists
	}

	// Indexing support
	indexTag := field.Tag.Get("index")
	if indexTag != "" {
		indexName := fmt.Sprintf("idx_%s_%s", tableName, columnName)
		unique := ""
		if indexTag == "unique" {
			unique = "UNIQUE"
		}
		indexQuery := fmt.Sprintf("CREATE %s INDEX IF NOT EXISTS %s ON %s(%s)", unique, indexName, tableName, columnName)
		_, err := pool.Exec(ctx, indexQuery)
		if err != nil {
			logger.Warn("Failed to create index", zap.String("index", indexName), zap.Error(err))
		} else {
			logger.Debug("Index ensured", zap.String("index", indexName))
		}
	}
}

func getColumnName(field reflect.StructField) string {
	tag := field.Tag.Get("db")
	if tag != "" {
		return strings.Split(tag, ",")[0]
	}
	tag = field.Tag.Get("json")
	if tag != "" {
		return strings.Split(tag, ",")[0]
	}
	return strings.ToLower(field.Name)
}

func getPostgresType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return "TEXT"
	case reflect.Int, reflect.Int32:
		return "INTEGER"
	case reflect.Int64:
		return "BIGINT"
	case reflect.Float32, reflect.Float64:
		return "DOUBLE PRECISION"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Struct:
		if t.Name() == "Time" {
			return "TIMESTAMP WITH TIME ZONE"
		}
		if strings.Contains(t.Name(), "UUID") {
			return "UUID"
		}
		return "JSONB"
	case reflect.Slice, reflect.Map:
		return "JSONB"
	default:
		return "TEXT"
	}
}

// SeedData reads the dev seed SQL file and executes it if the database is fresh.
func SeedData(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) error {
	logger.Info("🌱 Checking for Seed Data requirements...")

	// Defensive check: only seed if 'admin@delta.io' is missing
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = 'admin@delta.io')").Scan(&exists)
	if err != nil {
		logger.Error("Failed to check admin user existence for seeding", zap.Error(err))
		return err
	}
	
	if exists {
		logger.Info("✅ Admin user already exists. Skipping auto-seed.")
		return nil
	}

	logger.Info("🚀 Fresh database detected. Executing Hyper-Defensive Seed Script...")

	// Try multiple possible paths for the seed file
	paths := []string{
		"seeds/seed_complete_dev_data.sql",
		"services/fleet-core/seeds/seed_complete_dev_data.sql",
		"../seeds/seed_complete_dev_data.sql",
	}

	var content []byte
	var finalPath string
	for _, p := range paths {
		absPath, _ := filepath.Abs(p)
		data, err := os.ReadFile(p)
		if err == nil {
			content = data
			finalPath = absPath
			break
		}
	}

	if content == nil {
		logger.Warn("⚠️ Seed file not found in any standard locations. Skipping seeding.")
		return nil
	}

	logger.Info("Found seed file", zap.String("path", finalPath))

	// Execute the full script
	_, err = pool.Exec(ctx, string(content))
	if err != nil {
		logger.Error("❌ Critical Failure during database seeding", zap.Error(err))
		return err
	}

	logger.Info("✨ Database successfully seeded with Premium Dev Data.")
	return nil
}
