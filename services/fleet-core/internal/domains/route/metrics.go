package route

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Route planning metrics
	RoutePlansGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_plans_generated_total",
		Help: "Total route plans generated",
	}, []string{"tenant_id", "optimization_strategy"})

	RoutePlanningRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_planning_requests_total",
		Help: "Total route planning requests",
	}, []string{"tenant_id", "planning_type"})

	RoutePlanningSuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_planning_success_rate_percentage",
		Help: "Route planning request success rate",
	}, []string{"tenant_id"})

	RoutePlanningDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "route_planning_duration_seconds",
		Help:    "Route planning algorithm execution time",
		Buckets: prometheus.ExponentialBuckets(0.5, 2, 10),
	}, []string{"tenant_id", "optimization_strategy", "waypoint_count"})

	// Waypoint metrics
	WaypointsProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_waypoints_processed_total",
		Help: "Total waypoints processed in route planning",
	}, []string{"tenant_id", "waypoint_type"})

	WaypointValidationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_waypoint_validation_total",
		Help: "Waypoint validation attempts",
	}, []string{"tenant_id", "validation_result"})

	// ETA metrics
	RouteETAAccuracy = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "route_eta_accuracy_seconds",
		Help:    "ETA prediction accuracy (actual - predicted). Negative = early, positive = late.",
		Buckets: prometheus.LinearBuckets(-600, 60, 21), // -600 to +600 in 60s buckets
	}, []string{"tenant_id", "route_length", "traffic_conditions"})

	OnTimeArrivalRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_on_time_arrival_rate_percentage",
		Help: "Percentage of arrivals within threshold of ETA",
	}, []string{"tenant_id", "threshold_minutes"})

	// Distance and time metrics
	RouteDistanceKm = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "route_distance_km",
		Help:    "Planned route distance in kilometers",
		Buckets: prometheus.ExponentialBuckets(1, 2, 14),
	}, []string{"tenant_id", "vehicle_type"})

	RouteDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "route_duration_seconds",
		Help:    "Planned route duration in seconds",
		Buckets: prometheus.ExponentialBuckets(300, 2, 12),
	}, []string{"tenant_id"})

	RouteActualVsPlannedRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_actual_vs_planned_ratio",
		Help: "Actual distance/time ratio to planned (1=exact)",
	}, []string{"tenant_id", "metric_type"})

	// Optimization metrics
	RouteOptimizationScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_optimization_score",
		Help: "Route optimization effectiveness score (0-100)",
	}, []string{"tenant_id", "optimization_strategy"})

	FuelEfficiencyImprovement = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_fuel_efficiency_improvement_percentage",
		Help: "Fuel efficiency improvement from optimized routing",
	}, []string{"tenant_id"})

	CostSavingsFromOptimization = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_cost_savings_total",
		Help: "Total cost savings from route optimization in INR",
	}, []string{"tenant_id", "cost_category"})

	// Traffic metrics
	TrafficImpactFactor = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_traffic_impact_factor",
		Help: "Traffic multiplier applied to route duration",
	}, []string{"tenant_id", "route_id", "time_of_day"})

	TrafficDelayMinutes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_traffic_delay_minutes",
		Help: "Traffic-induced delay in minutes",
	}, []string{"tenant_id", "road_type"})

	// Real-time rerouting
	ReroutesTriggeredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_reroutes_triggered_total",
		Help: "Automatic rerouting events",
	}, []string{"tenant_id", "trigger_reason"})

	RerouteAcceptanceRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_reroute_acceptance_rate_percentage",
		Help: "Percentage of suggested reroutes accepted by drivers",
	}, []string{"tenant_id"})

	RerouteImprovement = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_reroute_improvement_minutes",
		Help: "Time saved by rerouting in minutes",
	}, []string{"tenant_id", "reroute_type"})

	// Zone/area metrics
	GeoFencedRoutesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_geo_fenced_routes_total",
		Help: "Routes passing through geofenced zones",
	}, []string{"tenant_id", "zone_type"})

	RestrictedAreaViolations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_restricted_area_violations_total",
		Help: "Routes entering restricted areas",
	}, []string{"tenant_id", "restriction_type"})

	// Toll avoidance
	TollAvoidanceEnabled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_toll_avoidance_enabled_percentage",
		Help: "Routes with toll avoidance enabled",
	}, []string{"tenant_id"})

	TollCostSavings = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_toll_cost_savings_total",
		Help: "Total toll cost savings in INR",
	}, []string{"tenant_id"})

	// Driver preference compliance
	DriverPreferenceCompliance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_driver_preference_compliance_percentage",
		Help: "Route adherence to driver preferences",
	}, []string{"tenant_id", "preference_type"})

	// Multi-stop optimization
	MultiStopRoutesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_multi_stop_routes_total",
		Help: "Routes with multiple waypoints",
	}, []string{"tenant_id", "stop_count_bucket"})

	StopSequenceOptimal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_stop_sequence_optimal_percentage",
		Help: "Stop sequence optimally ordered",
	}, []string{"tenant_id"})

	// Real-time tracking
	RouteProgressPercentage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_progress_percentage",
		Help: "Current route completion percentage",
	}, []string{"tenant_id", "route_id"})

	RouteDeviationDetected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_deviation_detected_total",
		Help: "Significant route deviations from planned path",
	}, []string{"tenant_id", "deviation_type"})

	// Weather impact
	WeatherImpactFactor = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_weather_impact_factor",
		Help: "Weather impact multiplier on duration",
	}, []string{"tenant_id", "weather_condition"})

	WeatherDelaysTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "route_weather_delays_total",
		Help: "Routes delayed due to weather",
	}, []string{"tenant_id", "weather_type"})

	// Fuel efficiency metrics
	PlannedVsActualFuelEfficiency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_planned_vs_actual_fuel_efficiency_ratio",
		Help: "Actual fuel efficiency vs planned (1=exact)",
	}, []string{"tenant_id", "vehicle_type"})

	// Route comparison
	AlternativeRoutesGenerated = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "route_alternatives_generated_total",
		Help: "Alternative routes generated for comparison",
	}, []string{"tenant_id", "comparison_criteria"})
)

// Global metrics instance
var GlobalMetrics = struct {
	RoutePlansGeneratedTotal      *prometheus.CounterVec
	RoutePlanningRequestsTotal    *prometheus.CounterVec
	RoutePlanningSuccessRate      *prometheus.GaugeVec
	RoutePlanningDuration         *prometheus.HistogramVec
	WaypointsProcessedTotal       *prometheus.CounterVec
	WaypointValidationTotal       *prometheus.CounterVec
	RouteETAAccuracy              *prometheus.HistogramVec
	OnTimeArrivalRate             *prometheus.GaugeVec
	RouteDistanceKm               *prometheus.HistogramVec
	RouteDurationSeconds          *prometheus.HistogramVec
	RouteActualVsPlannedRatio     *prometheus.GaugeVec
	RouteOptimizationScore        *prometheus.GaugeVec
	FuelEfficiencyImprovement     *prometheus.GaugeVec
	CostSavingsFromOptimization   *prometheus.CounterVec
	TrafficImpactFactor           *prometheus.GaugeVec
	TrafficDelayMinutes           *prometheus.GaugeVec
	ReroutesTriggeredTotal        *prometheus.CounterVec
	RerouteAcceptanceRate         *prometheus.GaugeVec
	RerouteImprovement            *prometheus.GaugeVec
	GeoFencedRoutesTotal          *prometheus.CounterVec
	RestrictedAreaViolations      *prometheus.CounterVec
	TollAvoidanceEnabled          *prometheus.GaugeVec
	TollCostSavings               *prometheus.CounterVec
	DriverPreferenceCompliance    *prometheus.GaugeVec
	MultiStopRoutesTotal          *prometheus.CounterVec
	StopSequenceOptimal           *prometheus.GaugeVec
	RouteProgressPercentage       *prometheus.GaugeVec
	RouteDeviationDetected        *prometheus.CounterVec
	WeatherImpactFactor           *prometheus.GaugeVec
	WeatherDelaysTotal            *prometheus.CounterVec
	PlannedVsActualFuelEfficiency *prometheus.GaugeVec
	AlternativeRoutesGenerated    *prometheus.GaugeVec
}{
	RoutePlansGeneratedTotal:      RoutePlansGeneratedTotal,
	RoutePlanningRequestsTotal:    RoutePlanningRequestsTotal,
	RoutePlanningSuccessRate:      RoutePlanningSuccessRate,
	RoutePlanningDuration:         RoutePlanningDuration,
	WaypointsProcessedTotal:       WaypointsProcessedTotal,
	WaypointValidationTotal:       WaypointValidationTotal,
	RouteETAAccuracy:              RouteETAAccuracy,
	OnTimeArrivalRate:             OnTimeArrivalRate,
	RouteDistanceKm:               RouteDistanceKm,
	RouteDurationSeconds:          RouteDurationSeconds,
	RouteActualVsPlannedRatio:     RouteActualVsPlannedRatio,
	RouteOptimizationScore:        RouteOptimizationScore,
	FuelEfficiencyImprovement:     FuelEfficiencyImprovement,
	CostSavingsFromOptimization:   CostSavingsFromOptimization,
	TrafficImpactFactor:           TrafficImpactFactor,
	TrafficDelayMinutes:           TrafficDelayMinutes,
	ReroutesTriggeredTotal:        ReroutesTriggeredTotal,
	RerouteAcceptanceRate:         RerouteAcceptanceRate,
	RerouteImprovement:            RerouteImprovement,
	GeoFencedRoutesTotal:          GeoFencedRoutesTotal,
	RestrictedAreaViolations:      RestrictedAreaViolations,
	TollAvoidanceEnabled:          TollAvoidanceEnabled,
	TollCostSavings:               TollCostSavings,
	DriverPreferenceCompliance:    DriverPreferenceCompliance,
	MultiStopRoutesTotal:          MultiStopRoutesTotal,
	StopSequenceOptimal:           StopSequenceOptimal,
	RouteProgressPercentage:       RouteProgressPercentage,
	RouteDeviationDetected:        RouteDeviationDetected,
	WeatherImpactFactor:           WeatherImpactFactor,
	WeatherDelaysTotal:            WeatherDelaysTotal,
	PlannedVsActualFuelEfficiency: PlannedVsActualFuelEfficiency,
	AlternativeRoutesGenerated:    AlternativeRoutesGenerated,
}
