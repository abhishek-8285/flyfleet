package schemalib

func init() {
	// vehicle_documents
	Register(TableDef{
		Name: "vehicle_documents",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "document_type", Type: Text},
			{Name: "type", Type: Text},
			{Name: "file_url", Type: Text},
			{Name: "status", Type: Text, Default: "'active'"},
			{Name: "expiry_date", Type: Date},
		},
	})

	// driver_documents
	Register(TableDef{
		Name: "driver_documents",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "driver_id", Type: UUID, RefTable: "drivers"},
			{Name: "document_type", Type: Text},
			{Name: "type", Type: Text},
			{Name: "file_url", Type: Text},
			{Name: "status", Type: Text, Default: "'active'"},
			{Name: "expiry_date", Type: Date},
		},
	})

	// global_documents
	Register(TableDef{
		Name: "global_documents",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "owner_id", Type: UUID},
			{Name: "owner_type", Type: Text},
			{Name: "metadata", Type: JSONB, Default: "'{}'::jsonb"},
		},
	})

	// maintenance_logs
	Register(TableDef{
		Name: "maintenance_logs",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "actual_cost", Type: Numeric, Size: "15,2", Default: "0.00"},
			{Name: "total_cost", Type: Numeric, Size: "15,2", Default: "0.00"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// compliance tables (HOS/ELD)
	Register(TableDef{
		Name: "duty_status_logs",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "driver_id", Type: UUID, RefTable: "drivers"},
			{Name: "status", Type: Text},
		},
	})
	Register(TableDef{
		Name: "hos_daily_limits",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "driver_id", Type: UUID, RefTable: "drivers"},
			{Name: "date", Type: Date},
		},
	})
	Register(TableDef{
		Name: "hos_violations",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "driver_id", Type: UUID, RefTable: "drivers"},
		},
	})
	Register(TableDef{
		Name: "eld_events",
		Columns: []ColumnDef{
			{Name: "event_id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
		},
	})
	Register(TableDef{
		Name: "dvir_reports",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "driver_id", Type: UUID, RefTable: "drivers"},
		},
	})

	// India regional compliance tables
	Register(TableDef{
		Name: "local_permits",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "valid_to", Type: Date},
			{Name: "status", Type: Text},
		},
	})
	Register(TableDef{
		Name: "driver_verifications",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "validity_to", Type: Date},
			{Name: "status", Type: Text},
		},
	})
	Register(TableDef{
		Name: "vehicle_verifications",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "insurance_expiry", Type: Date},
			{Name: "puc_expiry", Type: Date},
			{Name: "status", Type: Text},
		},
	})
	Register(TableDef{
		Name: "gst_compliances",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "status", Type: Text},
		},
	})
}
