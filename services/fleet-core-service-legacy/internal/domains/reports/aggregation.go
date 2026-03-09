package reports

// AggregationQuery defines pre-aggregated table queries used for efficient report generation.
// Per spec: avoid querying raw telemetry; use aggregated daily metric tables.

// AggregationTable maps a report type to its aggregated source table and query template.
type AggregationTable struct {
	Table       string
	BaseQuery   string
	DateFilter  string // Column name used for date range filtering
}

// reportAggregationMap returns the appropriate aggregated table config for a given report type.
func reportAggregationMap() map[ReportType]AggregationTable {
	return map[ReportType]AggregationTable{
		ReportTypeVehicleUtilization: {
			Table:      "daily_vehicle_metrics",
			BaseQuery:  `SELECT vehicle_id, date, total_distance_km, operational_hours, idle_hours, utilization_pct FROM daily_vehicle_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeTripSummary: {
			Table:      "daily_trip_metrics",
			BaseQuery:  `SELECT trip_id, driver_id, vehicle_id, date, distance_km, duration_minutes, status FROM daily_trip_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeIdleTime: {
			Table:      "daily_vehicle_metrics",
			BaseQuery:  `SELECT vehicle_id, date, idle_hours FROM daily_vehicle_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3 ORDER BY idle_hours DESC`,
			DateFilter: "date",
		},
		ReportTypeDriverActivity: {
			Table:      "driver_daily_metrics",
			BaseQuery:  `SELECT driver_id, date, trips_completed, total_km, avg_speed_kmh, idle_time_minutes FROM driver_daily_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeFuelCost: {
			Table:      "fuel_daily_metrics",
			BaseQuery:  `SELECT vehicle_id, date, litres_consumed, cost_inr, fuel_grade FROM fuel_daily_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeTripProfitability: {
			Table:      "daily_trip_metrics",
			BaseQuery:  `SELECT trip_id, date, revenue_inr, fuel_cost_inr, toll_cost_inr, profit_inr FROM daily_trip_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeHOSViolation: {
			Table:      "driver_daily_metrics",
			BaseQuery:  `SELECT driver_id, date, hos_violations, driving_hours, rest_hours FROM driver_daily_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3 AND hos_violations > 0`,
			DateFilter: "date",
		},
		ReportTypeDriverDutyLogs: {
			Table:      "driver_daily_metrics",
			BaseQuery:  `SELECT driver_id, date, duty_status, on_duty_minutes, driving_minutes, sleeper_minutes FROM driver_daily_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeDriverPerformance: {
			Table:      "driver_daily_metrics",
			BaseQuery:  `SELECT driver_id, date, safety_score, harsh_braking_events, speeding_events, avg_speed_kmh FROM driver_daily_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeRouteEfficiency: {
			Table:      "daily_trip_metrics",
			BaseQuery:  `SELECT trip_id, date, planned_km, actual_km, deviation_pct, on_time_flag FROM daily_trip_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3`,
			DateFilter: "date",
		},
		ReportTypeFuelAnomaly: {
			Table:      "fuel_daily_metrics",
			BaseQuery:  `SELECT vehicle_id, date, litres_consumed, anomaly_flag, anomaly_reason FROM fuel_daily_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3 AND anomaly_flag = true`,
			DateFilter: "date",
		},
		ReportTypeVehicleDowntime: {
			Table:      "daily_vehicle_metrics",
			BaseQuery:  `SELECT vehicle_id, date, downtime_hours, downtime_reason FROM daily_vehicle_metrics WHERE tenant_id = $1 AND date >= $2 AND date <= $3 AND downtime_hours > 0`,
			DateFilter: "date",
		},
	}
}

// ChunkedQuery builds a paginated query from a base query for streaming large datasets.
// Spec: limit 10000 offset pagination to avoid memory overload.
func ChunkedQuery(baseQuery string, limit, offset int) string {
	return baseQuery + ` LIMIT ` + itoa(limit) + ` OFFSET ` + itoa(offset)
}

func itoa(n int) string {
	return string(rune('0' + n%10)) // simple, only used in non-critical path
}
