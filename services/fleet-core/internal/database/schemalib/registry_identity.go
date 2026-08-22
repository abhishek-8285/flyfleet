package schemalib

// Registry of canonical schemas for the Identity & Access Management domain.
//
// This init() is the declarative source of truth for every IAM/identity table.
// schemalib.Sync uses it to CREATE tables that are missing and to ALIGN any
// existing table whose columns drifted from these canonical definitions —
// regardless of what a migration file or legacy table declared.

func init() {
	// tenants — canonical identity table (organizations is a legacy alias kept
	// only for foreign keys).
	Register(TableDef{
		Name: "tenants",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "name", Type: Text},
			{Name: "tax_id", Type: Text, Unique: true},
			{Name: "status", Type: Text, Default: "'active'"},
			{Name: "metadata", Type: JSONB, Default: "'{}'::jsonb"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// organizations — legacy alias table kept solely for FK compatibility.
	Register(TableDef{
		Name: "organizations",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "name", Type: Text},
			{Name: "tax_id", Type: Text, Unique: true},
			{Name: "status", Type: Text, Default: "'active'"},
			{Name: "metadata", Type: JSONB, Default: "'{}'::jsonb"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// users — canonical IAM user record.
	Register(TableDef{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "email", Type: Text, Unique: true},
			{Name: "password_hash", Type: Text},
			{Name: "full_name", Type: Text},
			{Name: "status", Type: Text, Default: "'active'"},
			{Name: "email_verified", Type: Boolean, Default: "false"},
			{Name: "mfa_enabled", Type: Boolean, Default: "false"},
			{Name: "must_reset_password", Type: Boolean, Default: "false"},
			{Name: "theme", Type: Text, Default: "'light'"},
			{Name: "language", Type: Text, Default: "'en'"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "deleted_at", Type: Timestamptz},
		},
	})

	// roles — global and per-tenant roles.
	Register(TableDef{
		Name: "roles",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "name", Type: Text},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "description", Type: Text},
			{Name: "is_system", Type: Boolean, Default: "false"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// permissions — resource/action catalog (canonical superset of 058/078/080/099).
	Register(TableDef{
		Name: "permissions",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "resource", Type: Text},
			{Name: "action", Type: Text},
			{Name: "name", Type: Text, Unique: true},
			{Name: "category", Type: Text},
			{Name: "description", Type: Text},
			{Name: "bit_position", Type: Integer},
			{Name: "is_global", Type: Boolean, Default: "true"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// role_permissions — many-to-many role<->permission mapping (id-based PK).
	Register(TableDef{
		Name: "role_permissions",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "role_id", Type: UUID, RefTable: "roles"},
			{Name: "permission_id", Type: UUID, RefTable: "permissions"},
			{Name: "role", Type: Text},
			{Name: "permission_name", Type: Text},
		},
	})

	// user_roles — many-to-many user<->role (composite PK).
	Register(TableDef{
		Name: "user_roles",
		Columns: []ColumnDef{
			{Name: "user_id", Type: UUID, RefTable: "users"},
			{Name: "role_id", Type: UUID, RefTable: "roles"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "granted_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// user_tenants — membership link between users and tenants.
	Register(TableDef{
		Name: "user_tenants",
		Columns: []ColumnDef{
			{Name: "user_id", Type: UUID, RefTable: "users"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "role", Type: Text},
			{Name: "is_active", Type: Boolean, Default: "true"},
		},
	})

	// invitations — pending user invites (referenced by ensureIAMSchema).
	Register(TableDef{
		Name: "invitations",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "email", Type: Text},
			{Name: "token", Type: Text},
			{Name: "expires_at", Type: Timestamptz},
			{Name: "accepted", Type: Boolean, Default: "false"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})
}
