package schemalib

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================
// DDL generation + idempotent sync engine
// ============================================================

// tableExists reports whether a base table exists in the public schema.
func tableExists(ctx context.Context, pool *pgxpool.Pool, table string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		         WHERE table_schema='public' AND table_name=$1 AND table_type='BASE TABLE')`,
		table).Scan(&exists)
	return exists, err
}

// columnInfo holds the live metadata for one column.
type columnInfo struct {
	dataType string
	def      string
	nullable string
}

// columnsFor returns a map of column name -> live metadata for a table.
func columnsFor(ctx context.Context, pool *pgxpool.Pool, table string) (map[string]columnInfo, error) {
	rows, err := pool.Query(ctx, `
		SELECT column_name, data_type, column_default, is_nullable
		FROM information_schema.columns
		WHERE table_schema='public' AND table_name=$1
		ORDER BY ordinal_position`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]columnInfo{}
	for rows.Next() {
		var name, dt, nullable string
		var def *string
		if err := rows.Scan(&name, &dt, &def, &nullable); err != nil {
			return out, err
		}
		d := ""
		if def != nil {
			d = *def
		}
		out[name] = columnInfo{dataType: dt, def: d, nullable: nullable}
	}
	return out, rows.Err()
}

// pgIdent quotes a SQL identifier safely.
func pgIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// liveTypeMatchesCanonical reports whether the live data_type already matches
// the declared canonical type from the registry.
func liveTypeMatchesCanonical(live string, declared Type) bool {
	want := string(declared)
	switch declared {
	case Timestamptz:
		return live == "timestamp with time zone"
	case Numeric:
		return live == "numeric"
	case Varchar:
		return strings.HasPrefix(live, "character varying")
	default:
		return strings.EqualFold(live, want)
	}
}

// declaredTypeSQL renders the full declared type (with size) for DDL.
func declaredTypeSQL(c ColumnDef) string {
	return c.Type.Size(c.Size)
}

// buildCreateTable renders a CREATE TABLE IF NOT EXISTS statement.
func buildCreateTable(t TableDef) string {
	var b strings.Builder
	b.WriteString("CREATE TABLE IF NOT EXISTS ")
	b.WriteString(pgIdent(t.Name))
	b.WriteByte('(')

	var cols []string
	var pk []string
	for _, c := range t.Columns {
		cols = append(cols, pgIdent(c.Name)+" "+c.columnBody())
		if c.PrimaryKey {
			pk = append(pk, pgIdent(c.Name))
		}
	}
	b.WriteString(strings.Join(cols, ", "))

	// Primary key
	if len(pk) > 0 {
		b.WriteString(", PRIMARY KEY (")
		b.WriteString(strings.Join(pk, ", "))
		b.WriteString(")")
	}

	// Unique constraints declared per-column
	for _, c := range t.Columns {
		if c.Unique && !c.PrimaryKey {
			b.WriteString(", UNIQUE (")
			b.WriteString(pgIdent(c.Name))
			b.WriteString(")")
		}
	}
	b.WriteByte(')')
	return b.String()
}

// buildAddColumn renders an ADD COLUMN IF NOT EXISTS for a missing column.
func buildAddColumn(t TableDef, c ColumnDef) string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s",
		pgIdent(t.Name), pgIdent(c.Name), c.columnBody())
}

// coerceColumnStmt renders an ALTER COLUMN TYPE to the canonical declared type
// for a column whose live type has drifted. Returns ("", false) when the live
// type already matches or a safe cast is not possible.
func coerceColumnStmt(table string, c ColumnDef, live string) (string, bool) {
	if liveTypeMatchesCanonical(live, c.Type) {
		return "", false
	}
	// Only coerce from safe, castable source types.
	switch c.Type {
	case UUID:
		if !isCharType(live) {
			return "", false // never coerce integer/numeric PKs to uuid
		}
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s",
			pgIdent(table), pgIdent(c.Name), c.Type, pgIdent(c.Name), c.Type), true
	case Timestamptz:
		if live != "timestamp without time zone" && live != "time without time zone" {
			return "", false
		}
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			pgIdent(table), pgIdent(c.Name), c.Type), true
	case Text:
		if live != "character varying" && live != "character" {
			return "", false
		}
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			pgIdent(table), pgIdent(c.Name), c.Type), true
	}
	return "", false
}

func isCharType(dt string) bool {
	return dt == "character varying" || dt == "character" || dt == "text" ||
		strings.HasPrefix(dt, "varchar") || strings.HasPrefix(dt, "char(")
}

// defaultStmt renders the SET DEFAULT / DROP DEFAULT needed to match the
// declared default, or "" when already correct.
func defaultStmt(table string, c ColumnDef, live columnInfo) string {
	want := c.Default
	if want == "" {
		if live.def == "" {
			return ""
		}
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT",
			pgIdent(table), pgIdent(c.Name))
	}
	if live.def == "" || live.def != want {
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
			pgIdent(table), pgIdent(c.Name), want)
	}
	return ""
}

// nullabilityStmt renders SET/DROP NOT NULL to match the declared nullability.
func nullabilityStmt(table string, c ColumnDef, live columnInfo) string {
	wantNullable := "NULL"
	if !c.Null {
		wantNullable = "NOT NULL"
	}
	if live.nullable == wantNullable {
		return ""
	}
	if c.Null {
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL",
			pgIdent(table), pgIdent(c.Name))
	}
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL",
		pgIdent(table), pgIdent(c.Name))
}

// indexStmt renders a CREATE INDEX IF NOT EXISTS for an indexed column.
func indexStmt(table string, c ColumnDef) string {
	if !c.Index {
		return ""
	}
	idx := fmt.Sprintf("idx_%s_%s", table, c.Name)
	if c.Unique {
		return fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s)",
			pgIdent(idx), pgIdent(table), pgIdent(c.Name))
	}
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
		pgIdent(idx), pgIdent(table), pgIdent(c.Name))
}

// syncTable brings a single registered table into conformance with its
// canonical TableDef. It is idempotent and safe to re-run.
func syncTable(ctx context.Context, pool *pgxpool.Pool, table TableDef) (int, error) {
	exists, err := tableExists(ctx, pool, table.Name)
	if err != nil {
		return 0, fmt.Errorf("check table existence: %w", err)
	}

	coerced := 0

	if !exists {
		// Brand-new table: create with full canonical definition.
		stmt := buildCreateTable(table)
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return 0, fmt.Errorf("create table %s: %w", table.Name, err)
		}
		for _, c := range table.Columns {
			if idx := indexStmt(table.Name, c); idx != "" {
				if _, err := pool.Exec(ctx, idx); err != nil {
					return coerced, fmt.Errorf("create index on %s.%s: %w", table.Name, c.Name, err)
				}
			}
		}
		return 0, nil
	}

	// Existing table: ensure each column exists with canonical type/default/nullability.
	liveCols, err := columnsFor(ctx, pool, table.Name)
	if err != nil {
		return 0, fmt.Errorf("inspect columns for %s: %w", table.Name, err)
	}

	for _, c := range table.Columns {
		live, found := liveCols[c.Name]

		// 1. Add missing column.
		if !found {
			if _, err := pool.Exec(ctx, buildAddColumn(table, c)); err != nil {
				// A pre-existing NOT-NULL-less-add could fail if a migration already
				// created the column; treat as non-fatal and continue to coercion below.
				continue
			}
		} else {
			live = liveCols[c.Name]
		}

		// 2. Coerce type if it drifted from canonical.
		if stmt, ok := coerceColumnStmt(table.Name, c, live.dataType); ok {
			if _, err := pool.Exec(ctx, stmt); err != nil {
				// Skip — a bad value would block the cast; the standardizer logs at debug.
				continue
			}
			coerced++
		}

		// 3. Normalize the default expression (e.g. uuid_generate_v4 -> gen_random_uuid).
		if d := defaultStmt(table.Name, c, live); d != "" {
			if _, err := pool.Exec(ctx, d); err != nil {
				continue
			}
		}

		// 4. Normalize nullability.
		if n := nullabilityStmt(table.Name, c, live); n != "" {
			if _, err := pool.Exec(ctx, n); err != nil {
				continue
			}
		}

		// 5. Ensure index/unique.
		if idx := indexStmt(table.Name, c); idx != "" {
			if _, err := pool.Exec(ctx, idx); err != nil {
				continue
			}
		}
	}

	return coerced, nil
}
