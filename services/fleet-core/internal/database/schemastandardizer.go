package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ============================================================
// Universal Column-Type Standardization
// ============================================================
//
// This file is the SINGLE, AUTHORITATIVE source of truth for column-type
// conventions across the entire FlyFleet schema. Every migration, every
// ensure*Schema helper, and every legacy table is funneled through
// StandardizeColumnTypes at startup so that no matter what type an individual
// migration declared, the running database always ends up canonical.
//
// The supported canonical conventions are:
//
//   - id, *_id columns            -> uuid   (the primary source of disparity)
//   - *_at, timestamp, time, ...   -> timestamptz
//   - status, name, email, ...     -> text   (VARCHAR(n) coerced to TEXT)
//   - uuid defaults                  -> gen_random_uuid()  (not uuid_generate_v4())
//
// Coercion is ALWAYS guarded:
//   - Only columns whose data_type is a safe cast-source are touched (e.g. an
//     integer primary key is NEVER forced to uuid).
//   - Each ALTER runs in its own error-guarded step so one bad column never
//     aborts the whole pass.
//   - The function is idempotent: re-running it is a no-op when the schema is
//     already canonical.

// stringIDTypes maps *_id-suffixed columns that are intentionally STRING
// identifiers to their canonical type. Columns NOT listed here fall through to
// the generic "ends in _id => uuid" rule. An empty value means "skip entirely"
// (the column is too ambiguous to coerce globally — e.g. device_id is a UUID FK
// in some tables and a VARCHAR serial in others, so it is left to the
// per-table schemalib registry).
var stringIDTypes = map[string]string{
	"gps_device_id":           "text", // device serial code -> broaden VARCHAR to text
	"device_id":               "",     // ambiguous (UUID FK vs VARCHAR serial) — skip globally
	"transaction_id":          "text", // external bank / FASTag transaction refs
	"external_transaction_id": "text", // external gateway transaction refs
}

// extraTimestampColumns are timestamp columns that do NOT use the trailing "_at"
// suffix but must still be timestamptz.
var extraTimestampColumns = map[string]bool{
	"timestamp":   true,
	"time":        true,
	"last_seen":   true,
	"pickup_time": true,
	"drop_time":   true,
}

// textColumns lists common free-form string columns that must always be TEXT
// (coerces VARCHAR(n) / CHARACTER(n) -> TEXT to remove length-cap disparity).
var textColumns = map[string]bool{
	"status":      true,
	"name":        true,
	"email":       true,
	"description": true,
	"full_name":   true,
	"action":      true,
	"resource":    true,
	"category":    true,
	"color":       true,
	"make":        true,
	"model":       true,
	"currency":    true,
}

// skippedTables are tracking tables that are never coerced.
var skippedTables = map[string]bool{
	"schema_migrations":        true,
	"schema_migrations_legacy": true,
}

// canonicalColumnType returns the canonical PostgreSQL type for a column name,
// or "" when no universal rule applies (column is left untouched).
func canonicalColumnType(col string) string {
	// UUID identity & foreign-key columns.
	if col == "id" || strings.HasSuffix(col, "_id") {
		if t, ok := stringIDTypes[col]; ok {
			return t // "" => no rule, skip
		}
		return "uuid"
	}
	// Timestamps.
	if strings.HasSuffix(col, "_at") || extraTimestampColumns[col] {
		return "timestamptz"
	}
	// Common string columns -> text.
	if textColumns[col] {
		return "text"
	}
	return ""
}

// isCanonicalType reports whether the current information_schema data_type is
// already the canonical form for the target type.
func isCanonicalType(current, target string) bool {
	switch target {
	case "uuid":
		return current == "uuid"
	case "timestamptz":
		// TIMESTAMPTZ and "TIMESTAMP WITH TIME ZONE" share one internal type;
		// information_schema always reports it as "timestamp with time zone".
		return current == "timestamp with time zone"
	case "text":
		return current == "text"
	default:
		return strings.EqualFold(current, target)
	}
}

// isStringType reports whether a data_type can hold character data that is
// safely castable to uuid / text.
func isStringType(dt string) bool {
	return dt == "character varying" ||
		dt == "character" ||
		dt == "text" ||
		strings.HasPrefix(dt, "varchar") ||
		strings.HasPrefix(dt, "char(")
}

// pgIdent quotes a SQL identifier safely (double-quotes with escaping).
func pgIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// coercionFor returns the ALTER statement (and a short rationale) needed to
// bring a column to its canonical type, or ("", "", false) when no safe
// coercion applies.
func coercionFor(table, column, currentType string) (stmt, rationale string, ok bool) {
	target := canonicalColumnType(column)
	if target == "" {
		return "", "", false
	}
	if isCanonicalType(currentType, target) {
		return "", "", false
	}

	switch target {
	case "uuid":
		// Only coerce from character types. Never touch integer/bigint/numeric PKs.
		if !isStringType(currentType) {
			return "", "", false
		}
		return fmt.Sprintf(
			`ALTER TABLE %s ALTER COLUMN %s TYPE uuid USING %s::uuid`,
			pgIdent(table), pgIdent(column), pgIdent(column)), "uuid", true

	case "timestamptz":
		// Only coerce naive timestamps (the real discrepancy). Columns already
		// holding "timestamp with time zone" are canonical and skipped above.
		if currentType != "timestamp without time zone" && currentType != "time without time zone" {
			return "", "", false
		}
		return fmt.Sprintf(
			`ALTER TABLE %s ALTER COLUMN %s TYPE timestamptz`,
			pgIdent(table), pgIdent(column)), "timestamptz", true

	case "text":
		// Only broaden varchar(n)/character(n) -> text (lossless). Leave other
		// types (jsonb, enum, etc.) untouched.
		if currentType != "character varying" && currentType != "character" {
			return "", "", false
		}
		return fmt.Sprintf(
			`ALTER TABLE %s ALTER COLUMN %s TYPE text`,
			pgIdent(table), pgIdent(column)), "text", true
	}
	return "", "", false
}

// StandardizeColumnTypes is the universal enforcer. It scans every base table
// in the 'public' schema and coerces any column whose type deviates from the
// platform canonical conventions. It is safe to run at any time — before or
// after migrations — and is fully idempotent.
func StandardizeColumnTypes(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("🛡️  Universal Column-Type Standardization starting...")

	// gen_random_uuid() is built-in on PostgreSQL 16+ and provided by the
	// pgcrypto extension on older versions. Ensure the extension exists so
	// default standardization succeeds everywhere.
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`); err != nil {
		logger.Warn("could not enable pgcrypto extension; gen_random_uuid defaults may be skipped",
			zap.Error(err))
	}

	coerced := 0

	// --- Pass 1: type coercion ---
	rows, err := pool.Query(ctx, `
		SELECT c.table_name, c.column_name, c.data_type
		FROM information_schema.columns c
		JOIN information_schema.tables t
		  ON t.table_schema = c.table_schema
		 AND t.table_name = c.table_name
		WHERE c.table_schema = 'public'
		  AND t.table_type = 'BASE TABLE'
		  AND t.table_name NOT IN ('schema_migrations', 'schema_migrations_legacy')
		ORDER BY c.table_name, c.ordinal_position`)
	if err != nil {
		logger.Error("universal standardization: failed to query columns", zap.Error(err))
	} else {
		for rows.Next() {
			var table, column, dataType string
			if err := rows.Scan(&table, &column, &dataType); err != nil {
				continue
			}
			stmt, rationale, ok := coercionFor(table, column, dataType)
			if !ok {
				continue
			}
			if _, err := pool.Exec(ctx, stmt); err != nil {
				logger.Debug("universal standardization: coercion skipped",
					zap.String("table", table), zap.String("column", column),
					zap.String("from", dataType), zap.String("to", rationale),
					zap.Error(err))
				continue
			}
			logger.Debug("universal standardization: coerced column",
				zap.String("table", table), zap.String("column", column),
				zap.String("from", dataType), zap.String("to", rationale))
			coerced++
		}
		rows.Close()
	}

	// --- Pass 2: default standardization (uuid_generate_v4 -> gen_random_uuid) ---
	defRows, err := pool.Query(ctx, `
		SELECT table_name, column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND column_default IS NOT NULL
		  AND column_default LIKE '%uuid_generate_v4%'
		  AND table_name NOT IN ('schema_migrations', 'schema_migrations_legacy')`)
	if err != nil {
		logger.Debug("universal standardization: could not enumerate uuid defaults",
			zap.Error(err))
	} else {
		for defRows.Next() {
			var table, column string
			if err := defRows.Scan(&table, &column); err != nil {
				continue
			}
			stmt := fmt.Sprintf(
				`ALTER TABLE %s ALTER COLUMN %s SET DEFAULT gen_random_uuid()`,
				pgIdent(table), pgIdent(column))
			if _, err := pool.Exec(ctx, stmt); err != nil {
				logger.Debug("universal standardization: default not updated",
					zap.String("table", table), zap.String("column", column),
					zap.Error(err))
				continue
			}
			logger.Debug("universal standardization: default uuid_generate_v4() -> gen_random_uuid()",
				zap.String("table", table), zap.String("column", column))
			coerced++
		}
		defRows.Close()
	}

	if coerced > 0 {
		logger.Info("✅ Universal column standardization complete", zap.Int("coerced", coerced))
	} else {
		logger.Info("✅ Universal column standardization complete — no discrepancies found")
	}
}
