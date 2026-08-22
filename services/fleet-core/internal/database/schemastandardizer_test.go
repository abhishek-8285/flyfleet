package database

import (
	"strings"
	"testing"
)

func TestCanonicalColumnType(t *testing.T) {
	tests := []struct {
		column string
		want   string
	}{
		// UUID identity & FK columns
		{"id", "uuid"},
		{"tenant_id", "uuid"},
		{"user_id", "uuid"},
		{"vehicle_id", "uuid"},
		{"driver_id", "uuid"},
		{"role_id", "uuid"},
		{"permission_id", "uuid"},
		{"current_driver_id", "uuid"},
		{"parent_vehicle_id", "uuid"},
		{"branch_id", "uuid"},
		{"cost_center_id", "uuid"},
		{"processed_invoice_id", "uuid"},
		{"from_tenant_id", "uuid"},
		{"to_tenant_id", "uuid"},

		// String-ID exceptions
		{"gps_device_id", "text"},
		{"device_id", ""},
		{"transaction_id", "text"},
		{"external_transaction_id", "text"},

		// Timestamps
		{"created_at", "timestamptz"},
		{"updated_at", "timestamptz"},
		{"deleted_at", "timestamptz"},
		{"processed_at", "timestamptz"},
		{"granted_at", "timestamptz"},
		{"expires_at", "timestamptz"},
		{"timestamp", "timestamptz"},
		{"time", "timestamptz"},
		{"last_seen", "timestamptz"},
		{"pickup_time", "timestamptz"},
		{"drop_time", "timestamptz"},

		// Common string columns
		{"status", "text"},
		{"name", "text"},
		{"email", "text"},
		{"description", "text"},
		{"full_name", "text"},
		{"action", "text"},
		{"resource", "text"},
		{"category", "text"},
		{"currency", "text"},

		// Unruled / should be left untouched
		{"bit_position", ""},
		{"capacity_tonnage", ""},
		{"distance", ""},
		{"latitude", ""},
		{"longitude", ""},
		{"metadata", ""},
	}

	for _, tt := range tests {
		t.Run(tt.column, func(t *testing.T) {
			got := canonicalColumnType(tt.column)
			if got != tt.want {
				t.Errorf("canonicalColumnType(%q) = %q, want %q", tt.column, got, tt.want)
			}
		})
	}
}

func TestCoercionFor(t *testing.T) {
	// id already uuid -> no-op
	if stmt, _, ok := coercionFor("vehicles", "id", "uuid"); ok {
		t.Errorf("expected no coercion for already-canonical uuid, got %q", stmt)
	}

	// id currently text -> coerce to uuid
	stmt, reason, ok := coercionFor("vehicles", "id", "character varying")
	if !ok {
		t.Fatal("expected coercion for varchar id -> uuid")
	}
	if reason != "uuid" {
		t.Errorf("reason = %q, want %q", reason, "uuid")
	}
	if !strings.Contains(stmt, "ALTER TABLE") || !strings.Contains(stmt, "TYPE uuid USING") {
		t.Errorf("unexpected stmt: %s", stmt)
	}

	// created_at currently "timestamp without time zone" -> timestamptz
	stmt, reason, ok = coercionFor("drivers", "created_at", "timestamp without time zone")
	if !ok {
		t.Fatal("expected coercion for naive timestamp -> timestamptz")
	}
	if reason != "timestamptz" {
		t.Errorf("reason = %q, want %q", reason, "timestamptz")
	}
	if !strings.Contains(stmt, "TYPE timestamptz") {
		t.Errorf("unexpected stmt: %s", stmt)
	}

	// created_at already "timestamp with time zone" -> no-op
	if _, _, ok := coercionFor("drivers", "created_at", "timestamp with time zone"); ok {
		t.Error("expected no coercion for already-canonical timestamptz")
	}

	// status currently "character varying(50)" data_type reports as "character varying" -> text
	stmt, reason, ok = coercionFor("drivers", "status", "character varying")
	if !ok {
		t.Fatal("expected coercion for varchar status -> text")
	}
	if reason != "text" {
		t.Errorf("reason = %q, want %q", reason, "text")
	}
	if !strings.Contains(stmt, "TYPE text") {
		t.Errorf("unexpected stmt: %s", stmt)
	}

	// status already "text" -> no-op
	if _, _, ok := coercionFor("drivers", "status", "text"); ok {
		t.Error("expected no coercion for already-canonical text")
	}
}

func TestCoercionForSafeGuards(t *testing.T) {
	// Integer primary key must NEVER be coerced to uuid.
	if _, _, ok := coercionFor("legacy_orders", "id", "integer"); ok {
		t.Error("integer id must NOT be coerced to uuid")
	}
	if _, _, ok := coercionFor("legacy_orders", "id", "bigint"); ok {
		t.Error("bigint id must NOT be coerced to uuid")
	}

	// gps_device_id is a string ID -> coerced to TEXT (broadened from VARCHAR),
	// never to uuid.
	stmt, reason, ok := coercionFor("vehicles", "gps_device_id", "character varying")
	if !ok {
		t.Fatal("expected gps_device_id varchar -> text coercion")
	}
	if reason != "text" {
		t.Errorf("gps_device_id reason = %q, want %q", reason, "text")
	}
	if strings.Contains(stmt, "TYPE uuid") {
		t.Error("gps_device_id must NOT be coerced to uuid")
	}

	// device_id is ambiguous (UUID FK vs VARCHAR serial) -> never coerced globally.
	if _, _, ok := coercionFor("telemetry_points", "device_id", "character varying"); ok {
		t.Error("device_id must not be coerced globally (ambiguous)")
	}
	if _, _, ok := coercionFor("devices", "device_id", "uuid"); ok {
		t.Error("device_id already uuid must not be coerced")
	}

	// A column with no canonical rule is left untouched.
	if _, _, ok := coercionFor("vehicles", "capacity_tonnage", "numeric"); ok {
		t.Error("capacity_tonnage has no canonical rule and must be skipped")
	}

	// status as integer (unusual) -> do not force to text
	if _, _, ok := coercionFor("weird", "status", "integer"); ok {
		t.Error("integer status must not be coerced to text")
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
