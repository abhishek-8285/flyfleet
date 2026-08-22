package schemalib

func init() {
	// telemetry_points — the core telemetry store.
	Register(TableDef{
		Name: "telemetry_points",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "device_id", Type: Text}, // device serial (string ID)
			{Name: "event_id", Type: UUID},
			{Name: "recorded_at", Type: Timestamptz},
			{Name: "latitude", Type: Double},
			{Name: "longitude", Type: Double},
			{Name: "speed", Type: Double},
			{Name: "heading", Type: Double},
			{Name: "ignition", Type: Boolean, Default: "false"},
			{Name: "fuel_level", Type: Double},
			{Name: "trip_id", Type: UUID, RefTable: "trips"},
			{Name: "ignition_status", Type: Text, Default: "'UNKNOWN'"},
		},
	})

	// geofences — polygon/circle geofence definitions.
	Register(TableDef{
		Name: "geofences",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "name", Type: Text, Index: true},
			{Name: "description", Type: Text},
			{Name: "type", Type: Text},
			{Name: "enabled", Type: Boolean, Default: "true"},
			{Name: "is_deleted", Type: Boolean, Default: "false"},
			{Name: "coordinates", Type: JSONB},
			{Name: "center_lat", Type: Double},
			{Name: "center_lng", Type: Double},
			{Name: "radius_m", Type: Double},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// geofence_events — recorded boundary crossings.
	Register(TableDef{
		Name: "geofence_events",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "geofence_id", Type: UUID, RefTable: "geofences"},
			{Name: "time", Type: Timestamptz},
		},
	})

	// tracking_sessions — active live-map sessions.
	Register(TableDef{
		Name: "tracking_sessions",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "started_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "ended_at", Type: Timestamptz},
		},
	})

	// alerts — operational alerts.
	Register(TableDef{
		Name: "alerts",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "type", Type: Text},
			{Name: "status", Type: Text, Default: "'ACTIVE'"},
			{Name: "severity", Type: Text, Default: "'warning'"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
			{Name: "updated_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// hardware_devices — per-tenant hardware inventory.
	Register(TableDef{
		Name: "hardware_devices",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "name", Type: Text},
			{Name: "device_id", Type: UUID},
			{Name: "type", Type: Text},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})
}
