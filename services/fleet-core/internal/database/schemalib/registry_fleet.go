package schemalib

func init() {
	// vehicles — the enterprise aggregate root.
	Register(TableDef{
		Name: "vehicles",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "registration_number", Type: Text},
			{Name: "registration_number_legacy", Type: Text},
			{Name: "status", Type: Text, Default: "'active'"},
			{Name: "deleted_at", Type: Timestamptz},
			{Name: "model", Type: Text},
			{Name: "make", Type: Text},
			{Name: "year", Type: Integer},
			{Name: "current_driver_id", Type: UUID, RefTable: "drivers"},
			{Name: "chassis_number", Type: Text},
			{Name: "engine_number", Type: Text},
			{Name: "vin_number", Type: Text},
			{Name: "vehicle_type", Type: Text},
			{Name: "capacity_tonnage", Type: Double},
			{Name: "fuel_type", Type: Text},
			{Name: "fuel_tank_capacity", Type: Double},
			{Name: "ownership_type", Type: Text},
			{Name: "parent_vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "branch_id", Type: UUID},
			{Name: "cost_center_id", Type: UUID},
			{Name: "gps_device_id", Type: Text},
			{Name: "gps_enabled", Type: Boolean, Default: "false"},
			{Name: "gps_device_type", Type: Text},
			{Name: "immobilizer_enabled", Type: Boolean, Default: "false"},
			{Name: "ais140_compliant", Type: Boolean, Default: "false"},
			{Name: "color", Type: Text},
			{Name: "odometer", Type: Double, Default: "0"},
			{Name: "speed_limit", Type: Double, Default: "0"},
			{Name: "is_blacklisted", Type: Boolean, Default: "false"},
			{Name: "is_hard_blocked", Type: Boolean, Default: "false"},
			{Name: "risk_score", Type: Double, Default: "0"},
			{Name: "compliance_status", Type: Text, Default: "'COMPLIANT'"},
			{Name: "version", Type: Integer, Default: "1"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// drivers — enterprise driver aggregate.
	Register(TableDef{
		Name: "drivers",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "status", Type: Text, Default: "'active'"},
			{Name: "phone", Type: Text},
			{Name: "email", Type: Text},
			{Name: "license_number", Type: Text},
			{Name: "license_expiry", Type: Date},
			{Name: "license_authority", Type: Text},
			{Name: "date_of_birth", Type: Date},
			{Name: "address", Type: JSONB, Default: "'{}'::jsonb"},
			{Name: "emergency_contact", Type: JSONB, Default: "'{}'::jsonb"},
			{Name: "joining_date", Type: Date},
			{Name: "resignation_date", Type: Date},
			{Name: "first_name", Type: Text},
			{Name: "last_name", Type: Text},
			{Name: "license_issue_date", Type: Date},
			{Name: "license_type", Type: Text},
			{Name: "experience_years", Type: Integer, Default: "0"},
			{Name: "rating", Type: Double, Default: "0"},
			{Name: "total_trips", Type: Integer, Default: "0"},
			{Name: "total_distance_km", Type: Double, Default: "0"},
			{Name: "total_hours_driven", Type: Double, Default: "0"},
			{Name: "behavior_score", Type: Integer, Default: "100"},
			{Name: "employment_type", Type: Text},
			{Name: "employee_code", Type: Text},
			{Name: "aadhaar_number", Type: Text},
			{Name: "pan_number", Type: Text},
			{Name: "documents", Type: JSONB, Default: "'[]'::jsonb"},
			{Name: "metadata", Type: JSONB, Default: "'{}'::jsonb"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "deleted_at", Type: Timestamptz},
			{Name: "org_id", Type: UUID, RefTable: "tenants"},
			{Name: "current_vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "current_trip_id", Type: UUID, RefTable: "trips"},
			{Name: "full_name", Type: Text},
			{Name: "compliance_status", Type: Text, Default: "'COMPLIANT'"},
			{Name: "version", Type: Integer, Default: "1"},
		},
	})

	// trips — vehicle trip lifecycle.
	Register(TableDef{
		Name: "trips",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "status", Type: Text, Default: "'CREATED'"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "driver_id", Type: UUID, RefTable: "drivers"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// customers — enterprise customer registry.
	Register(TableDef{
		Name: "customers",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "name", Type: Text},
			{Name: "email", Type: Text},
			{Name: "phone", Type: Text},
			{Name: "address", Type: Text},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})
}
