package schemalib

func init() {
	// analytics tables — all share a canonical id/tenant_id skeleton.
	analyticsTables := []string{
		"analytics_reports",
		"predictive_models",
		"prediction_results",
		"analytics_jobs",
		"fleet_analytics",
		"driver_behavior_analytics",
		"predictive_maintenance_analytics",
		"analytics_query_logs",
		"ml_feature_store",
		"analytics_quotas",
	}
	for _, name := range analyticsTables {
		Register(TableDef{
			Name: name,
			Columns: []ColumnDef{
				{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
				{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
				{Name: "status", Type: Text, Default: "'draft'"},
			},
		})
	}

	// vehicle_daily_stats
	Register(TableDef{
		Name: "vehicle_daily_stats",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "date", Type: Date},
		},
	})

	// driver_daily_stats
	Register(TableDef{
		Name: "driver_daily_stats",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "date", Type: Date},
		},
	})

	// fleet_hourly_stats
	Register(TableDef{
		Name: "fleet_hourly_stats",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
			{Name: "hour", Type: Timestamptz},
		},
	})

	// bulk_jobs
	Register(TableDef{
		Name: "bulk_jobs",
		Columns: []ColumnDef{
			{Name: "id", Type: UUID, PrimaryKey: true, Default: "gen_random_uuid()"},
			{Name: "tenant_id", Type: UUID, RefTable: "tenants"},
		},
	})
}
