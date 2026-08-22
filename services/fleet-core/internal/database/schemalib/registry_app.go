package schemalib

func init() {
	// assignment_operations — idempotent dispatch/assignment work.
	Register(TableDef{
		Name: "assignment_operations",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "idempotency_key", Type: Text, Unique: true},
			{Name: "payload_hash", Type: Text},
		},
	})

	// assignment_audit_logs — immutable assignment event log.
	Register(TableDef{
		Name: "assignment_audit_logs",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "driver_id", Type: UUID, RefTable: "drivers"},
			{Name: "from_state", Type: Text},
			{Name: "to_state", Type: Text},
			{Name: "event", Type: Text},
			{Name: "event_type", Type: Text},
			{Name: "actor_id", Type: UUID},
			{Name: "request_id", Type: UUID},
			{Name: "before_snapshot", Type: JSONB},
			{Name: "is_full_snapshot", Type: Boolean, Default: "false"},
			{Name: "remarks", Type: Text},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// yard_gate_logs — gated entry/exit transactions.
	Register(TableDef{
		Name: "yard_gate_logs",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "gate_name", Type: Text},
			{Name: "gatekeeper_id", Type: UUID},
			{Name: "action_type", Type: Text},
			{Name: "cargo_status", Type: Text},
			{Name: "weight_measured", Type: Numeric, Size: "10,2"},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// fintech_transactions — commission audit ledger.
	Register(TableDef{
		Name: "fintech_transactions",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "service_type", Type: Text},
			{Name: "transaction_amount", Type: Numeric, Size: "10,2"},
			{Name: "commission_earned", Type: Numeric, Size: "10,2"},
			{Name: "payment_status", Type: Text},
			{Name: "created_at", Type: Timestamptz, Default: "CURRENT_TIMESTAMP"},
		},
	})

	// telemetry_analytics / telemetry_anomalies / telemetry_events — analytics views.
	for _, name := range []string{"telemetry_analytics", "telemetry_anomalies", "telemetry_events"} {
		Register(TableDef{
			Name: name,
			Columns: []ColumnDef{
				{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
				{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
				{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			},
		})
	}

	// consumption_events — metered consumption ledger.
	Register(TableDef{
		Name: "consumption_events",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			{Name: "time", Type: Timestamptz},
			{Name: "meter_code", Type: Text},
			{Name: "amount", Type: Numeric, Size: "15,4", Default: "0.0000"},
			{Name: "metadata", Type: JSONB},
			{Name: "processed_invoice_id", Type: UUID},
		},
	})

	// video infrastructure
	for _, name := range []string{"video_recordings", "video_streams"} {
		Register(TableDef{
			Name: name,
			Columns: []ColumnDef{
				{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
				{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
				{Name: "vehicle_id", Type: UUID, RefTable: "vehicles"},
			},
		})
	}
}
