package cost

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Infrastructure cost metrics (track resource costs)
	CloudCostTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_cloud_total_usd",
		Help: "Total cloud infrastructure cost per day in USD",
	}, []string{"cloud_provider", "service", "region"})

	ComputeCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_compute_usd",
		Help: "Compute costs (VMs, containers, serverless)",
	}, []string{"service", "instance_type", "region"})

	StorageCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_storage_usd",
		Help: "Storage costs (databases, object storage, block volumes)",
	}, []string{"storage_type", "region", "tier"})

	NetworkCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_network_usd",
		Help: "Network costs (data transfer, load balancers)",
	}, []string{"service", "direction"})

	ManagedServiceCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_managed_service_usd",
		Help: "Managed service costs (RDS, ElastiCache, etc.)",
	}, []string{"service", "tier"})

	// Cost per business unit/organization
	CostPerOrganization = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_per_organization_usd",
		Help: "Daily cost attributed to each customer org",
	}, []string{"tenant_id", "cost_category"})

	CostPerTrip = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_per_trip_usd",
		Help: "Cost per completed trip",
	}, []string{"tenant_id", "vehicle_type"})

	CostPerDriver = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_per_driver_usd",
		Help: "Daily cost per active driver",
	}, []string{"tenant_id"})

	CostPerVehicle = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_per_vehicle_usd",
		Help: "Daily cost per active vehicle",
	}, []string{"tenant_id", "vehicle_type"})

	// Cost efficiency metrics
	ResourceUtilizationEfficiency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_resource_utilization_efficiency_percentage",
		Help: "Resource utilization efficiency (actual vs provisioned)",
	}, []string{"tenant_id", "resource_type"})

	CostUnderutilizationWaste = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_underutilization_waste_usd",
		Help: "Daily cost wasted due to underutilization",
	}, []string{"tenant_id", "resource_type"})

	IdleResourceCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_idle_resource_usd",
		Help: "Cost of idle/unused resources",
	}, []string{"resource_type"})

	// Observability cost
	ObservabilityCostTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_observability_total_usd",
		Help: "Total observability stack costs daily",
	}, []string{"tool", "usage_type"})

	MetricCardinalityCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_metric_cardinality_usd",
		Help: "Cost impact of high cardinality metrics",
	}, []string{"metric_name"})

	LogIngestionCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_log_ingestion_usd",
		Help: "Daily log ingestion cost",
	}, []string{"log_source", "volume_bucket"})

	TraceStorageCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_trace_storage_usd",
		Help: "Daily distributed trace storage cost",
	}, []string{"sampling_rate_bucket"})

	// Capacity planning metrics
	CostForecast30Days = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cost_forecast_30d_usd",
		Help: "Predicted total cloud cost for next 30 days",
	})

	CostForecastConfidence = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cost_forecast_confidence_percentage",
		Help: "Forecast confidence interval",
	})

	CostGrowthRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_growth_rate_percent",
		Help: "Weekly cost growth rate",
	}, []string{"service"})

	// Anomaly detection
	CostAnomalyScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_anomaly_score",
		Help: "Cost anomaly detection score (0-1, higher = more anomalous)",
	}, []string{"service", "cost_category"})

	CostAnomalyValue = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_anomaly_value_usd",
		Help: "Anomalous cost amount vs baseline",
	}, []string{"service", "anomaly_type"})

	// Budget tracking
	BudgetSetTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_budget_set_usd",
		Help: "Monthly budget allocated per org/service",
	}, []string{"tenant_id", "budget_type"})

	BudgetConsumptionPercentage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_budget_consumption_percentage",
		Help: "Budget consumption percentage (month-to-date)",
	}, []string{"tenant_id"})

	BudgetAlertTriggered = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cost_budget_alert_triggered_total",
		Help: "Budget threshold alerts triggered",
	}, []string{"tenant_id", "threshold_percentage"})

	// Cost allocation tags
	CostAllocatedByTag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_allocated_by_tag_usd",
		Help: "Cost allocation by tags (env, team, project)",
	}, []string{"tag_key", "tag_value"})

	UnallocatedCost = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cost_unallocated_usd",
		Help: "Cost without proper allocation tags",
	})

	// Spot/preemptible savings
	SpotInstanceSavings = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_spot_instance_savings_usd",
		Help: "Savings from spot/preemptible instances",
	}, []string{"instance_type"})

	ReservedInstanceCoverage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_reserved_instance_coverage_percentage",
		Help: "Percentage of compute covered by RIs/Savings Plans",
	}, []string{"instance_type"})

	ReservedInstanceSavings = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_reserved_instance_savings_usd",
		Help: "Savings from reserved instances",
	}, []string{"term_months", "payment_option"})

	// Scale optimization
	AutoScalingEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cost_autoscaling_events_total",
		Help: "Auto-scaling events triggered",
	}, []string{"service", "direction"})

	OverprovisionedCapacity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_overprovisioned_capacity_percentage",
		Help: "Over-provisioned capacity percentage",
	}, []string{"service"})

	// Data transfer costs
	DataTransferCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_data_transfer_usd",
		Help: "Data transfer costs by direction",
	}, []string{"direction", "region_pair"})

	CDNCacheHitRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_cdn_cache_hit_ratio_percentage",
		Help: "CDN cache hit rate (affects origin costs)",
	}, []string{"cdn_provider"})

	// Storage optimization
	StorageTierDistribution = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_storage_tier_distribution_bytes",
		Help: "Data distribution across storage tiers",
	}, []string{"storage_type", "tier"})

	InfrequentAccessSavings = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_infrequent_access_savings_usd",
		Help: "Savings from IA/Glacier storage tiers",
	}, []string{"storage_type"})

	// Database costs
	DatabaseCostPerReadWrite = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_database_per_rw_usd",
		Help: "Cost per read/write operation",
	}, []string{"database_type", "operation"})

	DatabaseConnectionCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_database_connection_usd",
		Help: "Cost per database connection-hour",
	}, []string{"database_type"})

	// Cost per request
	CostPerAPIRequest = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_per_api_request_usd",
		Help: "Cost per API request",
	}, []string{"endpoint", "method"})

	CostPerTelemetryMessage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_per_telemetry_message_usd",
		Help: "Cost per telemetry message ingested",
	}, []string{"message_type"})

	// Optimization opportunity metrics
	RightsizingOpportunities = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_rightsizing_opportunities_count",
		Help: "Count of instances that could be downsized",
	}, []string{"instance_type", "current_size", "recommended_size"})

	RightsizingSavings = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_rightsizing_savings_usd",
		Help: "Potential monthly savings from rightsizing",
	}, []string{"instance_type"})

	OrphanedResourceCost = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_orphaned_resource_usd",
		Help: "Cost of unattached/orphaned resources",
	}, []string{"resource_type"})

	// Cost attribution
	CostByEnvironment = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_by_environment_usd",
		Help: "Cost breakdown by environment (prod/staging/dev)",
	}, []string{"environment"})

	CostByTeam = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cost_by_team_usd",
		Help: "Cost attributed to teams",
	}, []string{"team"})
)

// Global metrics instance
var GlobalMetrics = struct {
	CloudCostTotal                *prometheus.GaugeVec
	ComputeCost                   *prometheus.GaugeVec
	StorageCost                   *prometheus.GaugeVec
	NetworkCost                   *prometheus.GaugeVec
	ManagedServiceCost            *prometheus.GaugeVec
	CostPerOrganization           *prometheus.GaugeVec
	CostPerTrip                   *prometheus.GaugeVec
	CostPerDriver                 *prometheus.GaugeVec
	CostPerVehicle                *prometheus.GaugeVec
	ResourceUtilizationEfficiency *prometheus.GaugeVec
	CostUnderutilizationWaste     *prometheus.GaugeVec
	IdleResourceCost              *prometheus.GaugeVec
	ObservabilityCostTotal        *prometheus.GaugeVec
	MetricCardinalityCost         *prometheus.GaugeVec
	LogIngestionCost              *prometheus.GaugeVec
	TraceStorageCost              *prometheus.GaugeVec
	CostForecast30Days            prometheus.Gauge
	CostForecastConfidence        prometheus.Gauge
	CostGrowthRate                *prometheus.GaugeVec
	CostAnomalyScore              *prometheus.GaugeVec
	CostAnomalyValue              *prometheus.GaugeVec
	BudgetSetTotal                *prometheus.GaugeVec
	BudgetConsumptionPercentage   *prometheus.GaugeVec
	BudgetAlertTriggered          *prometheus.CounterVec
	CostAllocatedByTag            *prometheus.GaugeVec
	UnallocatedCost               prometheus.Gauge
	SpotInstanceSavings           *prometheus.GaugeVec
	ReservedInstanceCoverage      *prometheus.GaugeVec
	ReservedInstanceSavings       *prometheus.GaugeVec
	AutoScalingEvents             *prometheus.CounterVec
	OverprovisionedCapacity       *prometheus.GaugeVec
	DataTransferCost              *prometheus.GaugeVec
	CDNCacheHitRatio              *prometheus.GaugeVec
	StorageTierDistribution       *prometheus.GaugeVec
	InfrequentAccessSavings       *prometheus.GaugeVec
	DatabaseCostPerReadWrite      *prometheus.GaugeVec
	DatabaseConnectionCost        *prometheus.GaugeVec
	CostPerAPIRequest             *prometheus.GaugeVec
	CostPerTelemetryMessage       *prometheus.GaugeVec
	RightsizingOpportunities      *prometheus.GaugeVec
	RightsizingSavings            *prometheus.GaugeVec
	OrphanedResourceCost          *prometheus.GaugeVec
	CostByEnvironment             *prometheus.GaugeVec
	CostByTeam                    *prometheus.GaugeVec
}{
	CloudCostTotal:                CloudCostTotal,
	ComputeCost:                   ComputeCost,
	StorageCost:                   StorageCost,
	NetworkCost:                   NetworkCost,
	ManagedServiceCost:            ManagedServiceCost,
	CostPerOrganization:           CostPerOrganization,
	CostPerTrip:                   CostPerTrip,
	CostPerDriver:                 CostPerDriver,
	CostPerVehicle:                CostPerVehicle,
	ResourceUtilizationEfficiency: ResourceUtilizationEfficiency,
	CostUnderutilizationWaste:     CostUnderutilizationWaste,
	IdleResourceCost:              IdleResourceCost,
	ObservabilityCostTotal:        ObservabilityCostTotal,
	MetricCardinalityCost:         MetricCardinalityCost,
	LogIngestionCost:              LogIngestionCost,
	TraceStorageCost:              TraceStorageCost,
	CostForecast30Days:            CostForecast30Days,
	CostForecastConfidence:        CostForecastConfidence,
	CostGrowthRate:                CostGrowthRate,
	CostAnomalyScore:              CostAnomalyScore,
	CostAnomalyValue:              CostAnomalyValue,
	BudgetSetTotal:                BudgetSetTotal,
	BudgetConsumptionPercentage:   BudgetConsumptionPercentage,
	BudgetAlertTriggered:          BudgetAlertTriggered,
	CostAllocatedByTag:            CostAllocatedByTag,
	UnallocatedCost:               UnallocatedCost,
	SpotInstanceSavings:           SpotInstanceSavings,
	ReservedInstanceCoverage:      ReservedInstanceCoverage,
	ReservedInstanceSavings:       ReservedInstanceSavings,
	AutoScalingEvents:             AutoScalingEvents,
	OverprovisionedCapacity:       OverprovisionedCapacity,
	DataTransferCost:              DataTransferCost,
	CDNCacheHitRatio:              CDNCacheHitRatio,
	StorageTierDistribution:       StorageTierDistribution,
	InfrequentAccessSavings:       InfrequentAccessSavings,
	DatabaseCostPerReadWrite:      DatabaseCostPerReadWrite,
	DatabaseConnectionCost:        DatabaseConnectionCost,
	CostPerAPIRequest:             CostPerAPIRequest,
	CostPerTelemetryMessage:       CostPerTelemetryMessage,
	RightsizingOpportunities:      RightsizingOpportunities,
	RightsizingSavings:            RightsizingSavings,
	OrphanedResourceCost:          OrphanedResourceCost,
	CostByEnvironment:             CostByEnvironment,
	CostByTeam:                    CostByTeam,
}
