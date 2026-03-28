package maintenance

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Maintenance scheduling metrics
	MaintenanceTasksScheduledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_tasks_scheduled_total",
		Help: "Total maintenance tasks scheduled",
	}, []string{"org_id", "maintenance_type", "priority"})

	MaintenanceTasksCompletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_tasks_completed_total",
		Help: "Total maintenance tasks completed",
	}, []string{"org_id", "maintenance_type", "completion_type"})

	MaintenanceTasksOverdueTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_tasks_overdue_total",
		Help: "Currently overdue maintenance tasks",
	}, []string{"org_id", "overdue_by_days_bucket"})

	MaintenanceTaskCompletionRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_task_completion_rate_percentage",
		Help: "Maintenance task completion rate",
	}, []string{"org_id", "maintenance_type", "period"})

	// Predictive maintenance metrics
	PredictionModelsActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_prediction_models_active",
		Help: "Active predictive maintenance models",
	}, []string{"org_id", "model_type"})

	MaintenanceAlertsGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_alerts_generated_total",
		Help: "Total maintenance alerts generated",
	}, []string{"org_id", "alert_type", "severity"})

	MaintenanceAlertsResolvedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_alerts_resolved_total",
		Help: "Total maintenance alerts resolved",
	}, []string{"org_id", "resolution_type"})

	FalsePositiveAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_false_positive_alerts_total",
		Help: "False positive maintenance alerts",
	}, []string{"org_id", "alert_type"})

	// Parts inventory metrics
	PartsStockLevel = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_parts_stock_level",
		Help: "Current parts inventory level",
	}, []string{"org_id", "part_number", "warehouse"})

	PartsBelowReorderLevel = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_parts_below_reorder_level_total",
		Help: "Parts below reorder threshold",
	}, []string{"org_id", "part_category"})

	PartsOrdersPlacedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_parts_orders_placed_total",
		Help: "Total parts purchase orders placed",
	}, []string{"org_id", "supplier_id"})

	PartsDeliveryDelayDays = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "maintenance_parts_delivery_delay_days",
		Help:    "Parts delivery delay from promised date",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	}, []string{"org_id", "supplier_id"})

	// Cost metrics
	MaintenanceCostTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_cost_total",
		Help: "Total maintenance cost in INR",
	}, []string{"org_id", "cost_type", "vehicle_type"})

	MaintenanceCostPerVehicle = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_cost_per_vehicle",
		Help: "Average maintenance cost per vehicle in INR",
	}, []string{"org_id", "period"})

	UnplannedMaintenanceCostRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_unplanned_cost_ratio_percentage",
		Help: "Unplanned maintenance as % of total maintenance cost",
	}, []string{"org_id"})

	// Vehicle downtime metrics
	VehicleDowntimeHours = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_vehicle_downtime_hours",
		Help: "Total vehicle downtime in hours",
	}, []string{"org_id", "downtime_type", "vehicle_type"})

	DowntimePerVehicle = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_downtime_per_vehicle_hours",
		Help: "Average downtime per vehicle per month",
	}, []string{"org_id", "vehicle_type"})

	MeanTimeToRepair = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "maintenance_mttr_hours",
		Help:    "Mean Time To Repair (downtime to fix)",
		Buckets: prometheus.ExponentialBuckets(0.5, 2, 10),
	}, []string{"org_id", "maintenance_type"})

	MeanTimeBetweenFailures = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "maintenance_mtbf_hours",
		Help:    "Mean Time Between Failures",
		Buckets: prometheus.ExponentialBuckets(10, 2, 10),
	}, []string{"org_id", "vehicle_type"})

	// Workshop metrics
	WorkshopsActiveTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_workshops_active_total",
		Help: "Active maintenance workshops/garages",
	}, []string{"org_id", "workshop_type"})

	WorkshopUtilization = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_workshop_utilization_percentage",
		Help: "Workshop capacity utilization",
	}, []string{"org_id", "workshop_id"})

	TechnicianUtilization = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_technician_utilization_percentage",
		Help: "Technician time utilization",
	}, []string{"org_id", "skill_level"})

	// Vendor metrics
	VendorPerformanceScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_vendor_performance_score",
		Help: "Vendor SLA compliance score (0-100)",
	}, []string{"org_id", "vendor_id", "service_type"})

	VendorJobsCompletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_vendor_jobs_completed_total",
		Help: "Jobs completed by vendors",
	}, []string{"org_id", "vendor_id", "job_type"})

	// Technician metrics
	TechnicianProductivity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_technician_productivity_jobs_per_day",
		Help: "Technician productivity (jobs/day)",
	}, []string{"org_id", "technician_id"})

	TechnicianCertificationStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "maintenance_technician_certification_status",
		Help: "Technician certification validity (0=expired, 1=valid)",
	}, []string{"org_id", "technician_id", "certification_type"})

	// Warranty metrics
	WarrantyClaimsFiledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_warranty_claims_filed_total",
		Help: "Warranty claims filed with OEMs",
	}, []string{"org_id", "component_type"})

	WarrantyClaimsApprovedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_warranty_claims_approved_total",
		Help: "Warranty claims approved",
	}, []string{"org_id", "component_type"})

	WarrantyRecoveryAmount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_warranty_recovery_amount_total",
		Help: "Total warranty recovery amount in INR",
	}, []string{"org_id"})

	// Tire metrics
	TireReplacementsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_tire_replacements_total",
		Help: "Total tire replacements",
	}, []string{"org_id", "vehicle_type", "tire_position"})

	TireLifeKm = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "maintenance_tire_life_km",
		Help:    "Tire lifespan in kilometers",
		Buckets: prometheus.ExponentialBuckets(10000, 2, 10),
	}, []string{"org_id", "tire_type"})

	// Battery metrics
	BatteryReplacementsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_battery_replacements_total",
		Help: "Total battery replacements",
	}, []string{"org_id", "battery_type"})

	BatteryLifeMonths = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "maintenance_battery_life_months",
		Help:    "Battery lifespan in months",
		Buckets: prometheus.LinearBuckets(12, 6, 10),
	}, []string{"org_id", "battery_type"})

	// Oil change metrics
	OilChangesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "maintenance_oil_changes_total",
		Help: "Total oil changes performed",
	}, []string{"org_id", "oil_type"})

	OilChangeIntervalKm = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "maintenance_oil_change_interval_km",
		Help:    "Actual km between oil changes",
		Buckets: prometheus.LinearBuckets(5000, 1000, 20),
	}, []string{"org_id", "oil_type"})
)

// Global metrics instance
var GlobalMetrics = struct {
	MaintenanceTasksScheduledTotal  *prometheus.CounterVec
	MaintenanceTasksCompletedTotal  *prometheus.CounterVec
	MaintenanceTasksOverdueTotal    *prometheus.GaugeVec
	MaintenanceTaskCompletionRate   *prometheus.GaugeVec
	PredictionModelsActive          *prometheus.GaugeVec
	MaintenanceAlertsGeneratedTotal *prometheus.CounterVec
	MaintenanceAlertsResolvedTotal  *prometheus.CounterVec
	FalsePositiveAlertsTotal        *prometheus.CounterVec
	PartsStockLevel                 *prometheus.GaugeVec
	PartsBelowReorderLevel          *prometheus.GaugeVec
	PartsOrdersPlacedTotal          *prometheus.CounterVec
	PartsDeliveryDelayDays          *prometheus.HistogramVec
	MaintenanceCostTotal            *prometheus.CounterVec
	MaintenanceCostPerVehicle       *prometheus.GaugeVec
	UnplannedMaintenanceCostRatio   *prometheus.GaugeVec
	VehicleDowntimeHours            *prometheus.CounterVec
	DowntimePerVehicle              *prometheus.GaugeVec
	MeanTimeToRepair                *prometheus.HistogramVec
	MeanTimeBetweenFailures         *prometheus.HistogramVec
	WorkshopsActiveTotal            *prometheus.GaugeVec
	WorkshopUtilization             *prometheus.GaugeVec
	TechnicianUtilization           *prometheus.GaugeVec
	VendorPerformanceScore          *prometheus.GaugeVec
	VendorJobsCompletedTotal        *prometheus.CounterVec
	TechnicianProductivity          *prometheus.GaugeVec
	TechnicianCertificationStatus   *prometheus.GaugeVec
	WarrantyClaimsFiledTotal        *prometheus.CounterVec
	WarrantyClaimsApprovedTotal     *prometheus.CounterVec
	WarrantyRecoveryAmount          *prometheus.CounterVec
	TireReplacementsTotal           *prometheus.CounterVec
	TireLifeKm                      *prometheus.HistogramVec
	BatteryReplacementsTotal        *prometheus.CounterVec
	BatteryLifeMonths               *prometheus.HistogramVec
	OilChangesTotal                 *prometheus.CounterVec
	OilChangeIntervalKm             *prometheus.HistogramVec
}{
	MaintenanceTasksScheduledTotal:  MaintenanceTasksScheduledTotal,
	MaintenanceTasksCompletedTotal:  MaintenanceTasksCompletedTotal,
	MaintenanceTasksOverdueTotal:    MaintenanceTasksOverdueTotal,
	MaintenanceTaskCompletionRate:   MaintenanceTaskCompletionRate,
	PredictionModelsActive:          PredictionModelsActive,
	MaintenanceAlertsGeneratedTotal: MaintenanceAlertsGeneratedTotal,
	MaintenanceAlertsResolvedTotal:  MaintenanceAlertsResolvedTotal,
	FalsePositiveAlertsTotal:        FalsePositiveAlertsTotal,
	PartsStockLevel:                 PartsStockLevel,
	PartsBelowReorderLevel:          PartsBelowReorderLevel,
	PartsOrdersPlacedTotal:          PartsOrdersPlacedTotal,
	PartsDeliveryDelayDays:          PartsDeliveryDelayDays,
	MaintenanceCostTotal:            MaintenanceCostTotal,
	MaintenanceCostPerVehicle:       MaintenanceCostPerVehicle,
	UnplannedMaintenanceCostRatio:   UnplannedMaintenanceCostRatio,
	VehicleDowntimeHours:            VehicleDowntimeHours,
	DowntimePerVehicle:              DowntimePerVehicle,
	MeanTimeToRepair:                MeanTimeToRepair,
	MeanTimeBetweenFailures:         MeanTimeBetweenFailures,
	WorkshopsActiveTotal:            WorkshopsActiveTotal,
	WorkshopUtilization:             WorkshopUtilization,
	TechnicianUtilization:           TechnicianUtilization,
	VendorPerformanceScore:          VendorPerformanceScore,
	VendorJobsCompletedTotal:        VendorJobsCompletedTotal,
	TechnicianProductivity:          TechnicianProductivity,
	TechnicianCertificationStatus:   TechnicianCertificationStatus,
	WarrantyClaimsFiledTotal:        WarrantyClaimsFiledTotal,
	WarrantyClaimsApprovedTotal:     WarrantyClaimsApprovedTotal,
	WarrantyRecoveryAmount:          WarrantyRecoveryAmount,
	TireReplacementsTotal:           TireReplacementsTotal,
	TireLifeKm:                      TireLifeKm,
	BatteryReplacementsTotal:        BatteryReplacementsTotal,
	BatteryLifeMonths:               BatteryLifeMonths,
	OilChangesTotal:                 OilChangesTotal,
	OilChangeIntervalKm:             OilChangeIntervalKm,
}
