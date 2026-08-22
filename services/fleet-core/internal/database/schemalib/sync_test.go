package schemalib

import (
	"strings"
	"testing"
)

func TestCanonicalColumnType(t *testing.T) {
	// uuid column
	if !liveTypeMatchesCanonical("uuid", UUID) {
		t.Error("uuid should match uuid")
	}
	// ID columns that are uuid are canonical
	if !liveTypeMatchesCanonical("uuid", UUID) {
		t.Error("uuid is canonical for UUID type")
	}
	// TIMESTAMPTZ reports as "timestamp with time zone" in information_schema
	if !liveTypeMatchesCanonical("timestamp with time zone", Timestamptz) {
		t.Error("timestamp with time zone should match timestamptz canonical")
	}
	if !liveTypeMatchesCanonical("numeric", Numeric) {
		t.Error("numeric should match numeric")
	}
	if !liveTypeMatchesCanonical("character varying", Varchar) {
		t.Error("character varying should match varchar")
	}
	// mismatches
	if liveTypeMatchesCanonical("character varying", UUID) {
		t.Error("varchar should NOT match uuid canonical")
	}
	if liveTypeMatchesCanonical("timestamp without time zone", Timestamptz) {
		t.Error("timestamp without tz should NOT match timestamptz canonical")
	}
}

func TestCoerceColumnStmt(t *testing.T) {
	// varchar id -> uuid (safe cast)
	stmt, ok := coerceColumnStmt("vehicles", ColumnDef{Name: "id", Type: UUID}, "character varying")
	if !ok {
		t.Fatal("expected coercion varchar -> uuid")
	}
	if !strings.Contains(stmt, "TYPE uuid USING") {
		t.Errorf("expected USING clause, got: %s", stmt)
	}
	if !strings.Contains(stmt, "\"vehicles\"") {
		t.Errorf("expected quoted table, got: %s", stmt)
	}

	// already-uuid -> no coercion
	if _, ok := coerceColumnStmt("vehicles", ColumnDef{Name: "id", Type: UUID}, "uuid"); ok {
		t.Error("already-uuid should not be coerced")
	}

	// naive timestamp -> timestamptz
	stmt, ok = coerceColumnStmt("drivers", ColumnDef{Name: "created_at", Type: Timestamptz}, "timestamp without time zone")
	if !ok {
		t.Fatal("expected coercion timestamp-without-tz -> timestamptz")
	}
	if !strings.Contains(stmt, "TYPE timestamptz") {
		t.Errorf("expected timestamptz, got: %s", stmt)
	}

	// integer PK -> uuid: must NOT coerce (safe guard)
	if _, ok := coerceColumnStmt("legacy", ColumnDef{Name: "id", Type: UUID}, "integer"); ok {
		t.Error("integer id must NOT be coerced to uuid")
	}

	// varchar status -> text (lossless broadening)
	stmt, ok = coerceColumnStmt("drivers", ColumnDef{Name: "status", Type: Text}, "character varying")
	if !ok {
		t.Fatal("expected coercion varchar -> text")
	}
	if !strings.Contains(stmt, "TYPE text") {
		t.Errorf("expected text, got: %s", stmt)
	}

	// jsonb column registered as jsonb but live as text -> should NOT coerce
	// (jsonb is not in the safe-coerce set)
	if _, ok := coerceColumnStmt("x", ColumnDef{Name: "metadata", Type: JSONB}, "text"); ok {
		t.Error("should not coerce arbitrary text -> jsonb (not a safe rule)")
	}
}

func TestBuildCreateTable(t *testing.T) {
	td := TableDef{
		Name: "user_tenants",
		Columns: []ColumnDef{
			{Name: "user_id", Type: UUID, RefTable: "users"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "role", Type: Text},
			{Name: "is_active", Type: Boolean, Default: "true"},
		},
	}
	stmt := buildCreateTable(td)
	if !strings.HasPrefix(stmt, "CREATE TABLE IF NOT EXISTS") {
		t.Errorf("expected CREATE TABLE, got: %s", stmt)
	}
	// No column has PrimaryKey=true, so no PRIMARY KEY clause should be emitted.
	if strings.Contains(stmt, "PRIMARY KEY") {
		t.Errorf("did not expect a PK clause, got: %s", stmt)
	}
	if !strings.Contains(stmt, "\"user_tenants\"") {
		t.Errorf("expected quoted table name, got: %s", stmt)
	}
	if !strings.Contains(stmt, "REFERENCES \"users\"(\"id\")") {
		t.Errorf("expected FK to users(id), got: %s", stmt)
	}
}

func TestBuildCreateTableWithPK(t *testing.T) {
	td := TableDef{
		Name: "vehicles",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
		},
	}
	stmt := buildCreateTable(td)
	if !strings.Contains(stmt, "PRIMARY KEY (\"id\")") {
		t.Errorf("expected PRIMARY KEY clause, got: %s", stmt)
	}
}

func TestRegistryNonEmpty(t *testing.T) {
	tables := Tables()
	if len(tables) < 20 {
		t.Fatalf("expected at least 20 registered tables, got %d", len(tables))
	}
	// Ensure a few critical tables are registered.
	want := map[string]bool{"tenants": false, "users": false, "vehicles": false,
		"drivers": false, "trips": false, "permissions": false,
		"role_permissions": false, "user_roles": false, "invoices": false}
	for _, td := range tables {
		if _, ok := want[td.Name]; ok {
			want[td.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("critical table %q not registered", name)
		}
	}
}

func TestPgIdentQuoting(t *testing.T) {
	got := pgIdent("simple")
	if got != `"simple"` {
		t.Errorf("pgIdent(simple) = %q", got)
	}
	got = pgIdent(`has"quote`)
	if got != `"has""quote"` {
		t.Errorf("pgIdent(has\"quote) = %q", got)
	}
}
