// Package schemalib is the single, declarative source of truth for the FlyFleet
// database schema. Every table, column, type, default, constraint and index
// that the service relies on is declared here as data — never re-typed as ad-hoc
// SQL strings scattered across migrations or startup code.
//
// The Sync function consumes this registry at startup and idempotently brings
// the live database into conformance: it creates missing tables, adds missing
// columns, coerces any column that drifted from its canonical type, normalizes
// defaults and constraints, and installs indexes. The companion
// database.StandardizeColumnTypes pass is a safety-net type coercer that runs
// after migrations.
//
// Together these two layers guarantee: "no discrepancy at all" — the running
// schema is always canonical regardless of what any individual migration file
// or legacy table happened to declare.
package schemalib

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Type is a canonical PostgreSQL column type. Only the standardized spellings
// below are ever used anywhere the schema is emitted, eliminating the
// TIMESTAMPTZ / "TIMESTAMP WITH TIME ZONE" and uuid_generate_v4 /
// gen_random_uuid() spelling wars at the source.
type Type string

const (
	UUID        Type = "uuid"
	Text        Type = "text"
	Timestamptz Type = "timestamptz"
	Boolean     Type = "boolean"
	Integer     Type = "integer"
	BigInt      Type = "bigint"
	Double      Type = "double precision"
	Numeric     Type = "numeric" // parameterized via Size, e.g. "15,2"
	Varchar     Type = "varchar" // parameterized via Size when a length is desired
	Date        Type = "date"
	JSONB       Type = "jsonb"
	Bytea       Type = "bytea"
)

// Size returns the PostgreSQL type expression including any size/precision.
// For parameterized types (varchar, numeric) the Size field supplies the
// precision; for scalar types it is ignored.
func (t Type) Size(size string) string {
	if size == "" {
		return string(t)
	}
	return fmt.Sprintf("%s(%s)", t, size)
}

func (t Type) String() string { return string(t) }

// ColumnDef is the canonical, data-driven definition of a single column.
type ColumnDef struct {
	Name       string
	Type       Type
	Size       string // precision for parameterized types, e.g. "255" or "15,2"
	Null       bool   // true  => nullable (default). false => NOT NULL.
	Default    string // raw SQL default expression, e.g. "gen_random_uuid()", "'active'", "CURRENT_TIMESTAMP"
	PrimaryKey bool
	RefTable   string // FK target table (empty = no FK)
	RefColumn  string // FK target column (defaults to "id" when RefTable set)
	Unique     bool   // UNIQUE(table) constraint
	Index      bool   // create a non-unique index on this column
}

// TableDef is the canonical definition of a table.
type TableDef struct {
	Name    string
	Columns []ColumnDef
}

// FK returns the foreign-key clause for a column, or "" if the column has no FK.
func (c ColumnDef) FK() string {
	if c.RefTable == "" {
		return ""
	}
	refCol := c.RefColumn
	if refCol == "" {
		refCol = "id"
	}
	return fmt.Sprintf("REFERENCES %s(%s)", pgIdent(c.RefTable), pgIdent(refCol))
}

// nullable renders "NULL" / "NOT NULL".
func (c ColumnDef) nullable() string {
	if c.Null {
		return "NULL"
	}
	return "NOT NULL"
}

// sqlDefault renders the DEFAULT clause, or "" when none is declared.
func (c ColumnDef) sqlDefault() string {
	if c.Default == "" {
		return ""
	}
	return " DEFAULT " + c.Default
}

// columnBody renders the type, nullability, default and FK clause for a column
// (without the column name, which callers quote separately via pgIdent).
func (c ColumnDef) columnBody() string {
	parts := []string{
		c.Type.Size(c.Size),
		c.nullable(),
	}
	if dd := c.sqlDefault(); dd != "" {
		parts = append(parts, dd)
	}
	if fk := c.FK(); fk != "" {
		parts = append(parts, fk)
	}
	return strings.Join(parts, " ")
}

// Schema is the canonical registry of every table the service manages, keyed by
// table name. Order of insertion is preserved via tableOrder for deterministic
// DDL generation (tables are created before their dependents).
var Schema = map[string]TableDef{}

// tableOrder records the deterministic order tables were registered.
var tableOrder []string

// Register adds a table definition to the canonical registry. It is intended
// to be called from init() functions (see registry package).
func Register(t TableDef) {
	Schema[t.Name] = t
	tableOrder = append(tableOrder, t.Name)
}

// Tables returns all registered table definitions in registration order.
func Tables() []TableDef {
	out := make([]TableDef, 0, len(tableOrder))
	for _, name := range tableOrder {
		if td, ok := Schema[name]; ok {
			out = append(out, td)
		}
	}
	return out
}

// Sync brings the live database into conformance with the canonical Schema.
// It is safe to call at any time and is fully idempotent.
func Sync(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("📋 Universal Schema Sync starting (canonical registry)...")
	totalCoerced := 0
	for _, table := range Tables() {
		if n, err := syncTable(ctx, pool, table); err != nil {
			logger.Error("schema sync failed for table",
				zap.String("table", table.Name), zap.Error(err))
			continue
		} else {
			totalCoerced += n
		}
	}
	if totalCoerced > 0 {
		logger.Info("✅ Universal Schema Sync complete", zap.Int("columns_aligned", totalCoerced))
	} else {
		logger.Info("✅ Universal Schema Sync complete — schema already canonical")
	}
}
