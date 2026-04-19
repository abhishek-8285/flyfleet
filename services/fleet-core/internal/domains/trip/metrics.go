package trip

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Trip lifecycle metrics
	TripsStartedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_started_total",
		Help: "Total trips started",
	}, []string{"tenant_id", "trip_type", "vehicle_type"})

	TripsCancelledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_cancelled_total",
		Help: "Total trips cancelled",
	}, []string{"tenant_id", "cancellation_reason", "cancelled_by"})

	TripsPausedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_paused_total",
		Help: "Total trips paused (break, vehicle issue)",
	}, []string{"tenant_id", "pause_reason"})

	TripsResumedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_resumed_total",
		Help: "Total paused trips resumed",
	}, []string{"tenant_id"})

	// Trip duration metrics
	TripDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "trip_duration_seconds",
		Help:    "Trip duration from start to completion",
		Buckets: prometheus.ExponentialBuckets(300, 2, 12), // 5min to ~20hrs
	}, []string{"tenant_id", "trip_type", "vehicle_type"})

	TripDistanceKm = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "trip_distance_km",
		Help:    "Trip distance in kilometers",
		Buckets: prometheus.ExponentialBuckets(1, 2, 14),
	}, []string{"tenant_id", "trip_type"})

	TripPlannedVsActualRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_planned_vs_actual_ratio",
		Help: "Actual vs planned metrics ratio (1=exact)",
	}, []string{"tenant_id", "metric_type"})

	// Status tracking
	ActiveTripsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "active_trips_total",
		Help: "Currently active trips",
	}, []string{"tenant_id", "trip_type"})

	TripsInProgress = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trips_in_progress_total",
		Help: "Trips currently in progress",
	}, []string{"tenant_id"})

	// Waypoint metrics
	WaypointsVisitedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_waypoints_visited_total",
		Help: "Waypoints successfully visited",
	}, []string{"tenant_id", "waypoint_type"})

	WaypointsSkippedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_waypoints_skipped_total",
		Help: "Waypoints skipped during trip",
	}, []string{"tenant_id", "skip_reason"})

	WaypointArrivalTimeVariance = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "trip_waypoint_arrival_time_variance_seconds",
		Help:    "Difference between actual and planned waypoint arrival. Negative = early, positive = late.",
		Buckets: prometheus.LinearBuckets(-300, 30, 21), // -300 to +300 in 30s buckets
	}, []string{"tenant_id"})

	// Driver behavior metrics
	TripsWithGoodDriverBehavior = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_good_driver_behavior_total",
		Help: "Trips with good driver behavior (no harsh events)",
	}, []string{"tenant_id"})

	TripsWithHarshEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_harsh_events_total",
		Help: "Trips with harsh driving events",
	}, []string{"tenant_id", "harsh_event_type"})

	// Idling metrics
	IdlingTimeSeconds = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_idling_time_seconds",
		Help: "Total idling time during trips",
	}, []string{"tenant_id", "trip_id"})

	IdlingFuelWastedLiters = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_idling_fuel_wasted_liters",
		Help: "Fuel wasted due to idling",
	}, []string{"tenant_id"})

	// Speed metrics
	SpeedingEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_speeding_events_total",
		Help: "Speeding events during trips",
	}, []string{"tenant_id", "speed_limit_exceeded_by"})

	AverageTripSpeed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_average_speed_kmh",
		Help: "Average speed during trips",
	}, []string{"tenant_id", "trip_id"})

	// Client interaction metrics
	ClientCheckinsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_client_checkins_total",
		Help: "Client check-in/pickup confirmations",
	}, []string{"tenant_id", "checkin_type"})

	ProofOfDeliveryCapturedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_proof_of_delivery_captured_total",
		Help: "Proof of delivery captured (photo/signature)",
	}, []string{"tenant_id", "pod_type"})

	// Revenue metrics
	TripRevenueTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_revenue_total",
		Help: "Total trip revenue in INR",
	}, []string{"tenant_id", "trip_type", "payment_method"})

	TripCostTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_cost_total",
		Help: "Total trip cost in INR",
	}, []string{"tenant_id", "cost_type"})

	TripProfitMargin = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_profit_margin_percentage",
		Help: "Trip profit margin percentage",
	}, []string{"tenant_id", "trip_type"})

	// Waiting time metrics
	WaitingTimeSeconds = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_waiting_time_seconds",
		Help: "Total waiting time during trips",
	}, []string{"tenant_id", "waiting_type"})

	ClientLateArrivalTime = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_client_late_arrival_time_seconds",
		Help: "Client arrival delay time",
	}, []string{"tenant_id"})

	// Route adherence
	RouteDeviationDistanceKm = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_route_deviation_distance_km",
		Help: "Distance deviated from planned route",
	}, []string{"tenant_id", "deviation_type"})

	RouteAdherenceScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_route_adherence_score_percentage",
		Help: "Route adherence percentage",
	}, []string{"tenant_id", "trip_id"})

	// Customer rating
	TripCustomerRating = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_customer_rating",
		Help: "Customer trip rating (1-5)",
	}, []string{"tenant_id", "trip_id", "rating_category"})

	// Driver performance metrics
	DriverTripsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "driver_trips_completed_total",
		Help: "Trips completed by driver",
	}, []string{"tenant_id", "driver_id"})

	DriverTripRating = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "driver_trip_rating",
		Help: "Average driver trip rating",
	}, []string{"tenant_id", "driver_id"})

	// Fleet utilization
	FleetUtilizationRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fleet_utilization_rate_percentage",
		Help: "Vehicle fleet utilization percentage",
	}, []string{"tenant_id", "vehicle_type"})

	// Deadhead mileage
	DeadheadDistanceKm = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_deadhead_distance_km",
		Help: "Empty/return trip distance (deadhead)",
	}, []string{"tenant_id", "deadhead_type"})

	DeadheadPercentage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_deadhead_percentage",
		Help: "Deadhead distance as % of total distance",
	}, []string{"tenant_id"})

	// Trip assignment metrics
	TripsAutoAssignedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_auto_assigned_total",
		Help: "Trips automatically assigned to drivers",
	}, []string{"tenant_id"})

	TripsManuallyAssignedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trips_manually_assigned_total",
		Help: "Trips manually assigned by dispatchers",
	}, []string{"tenant_id"})

	AssignmentAcceptanceRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_assignment_acceptance_rate_percentage",
		Help: "Automatic assignment acceptance rate",
	}, []string{"tenant_id"})

	AssignmentRejectionRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_assignment_rejection_rate_percentage",
		Help: "Automatic assignment rejection rate",
	}, []string{"tenant_id", "rejection_reason"})

	// Time window compliance
	TimeWindowCompliance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_time_window_compliance_percentage",
		Help: "Trips completed within assigned time window",
	}, []string{"tenant_id"})

	EarlyArrivalBufferMinutes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "trip_early_arrival_buffer_minutes",
		Help:    "Early arrival time buffer before scheduled",
		Buckets: prometheus.LinearBuckets(0, 5, 20),
	}, []string{"tenant_id"})

	// Multi-drop efficiency
	MultiDropTripsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_multi_drop_total",
		Help: "Multi-drop trips (multiple waypoints)",
	}, []string{"tenant_id", "drop_count"})

	MultiDropEfficiency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trip_multi_drop_efficiency_ratio",
		Help: "Multi-drop efficiency vs individual trips",
	}, []string{"tenant_id"})

	// Trip state metrics
	TripStateTransitionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_state_transitions_total",
		Help: "Trip state machine transitions",
	}, []string{"tenant_id", "from_state", "to_state"})

	TripStateDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "trip_state_duration_seconds",
		Help:    "Time spent in each trip state",
		Buckets: prometheus.ExponentialBuckets(60, 2, 10),
	}, []string{"tenant_id", "state"})

	// Cancellation penalty
	CancellationPenaltyAppliedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trip_cancellation_penalty_applied_total",
		Help: "Cancellation penalties charged",
	}, []string{"tenant_id", "penalty_type"})
)

// Global metrics instance
var GlobalMetrics = struct {
	TripsStartedTotal               *prometheus.CounterVec
	TripsCancelledTotal             *prometheus.CounterVec
	TripsPausedTotal                *prometheus.CounterVec
	TripsResumedTotal               *prometheus.CounterVec
	TripDuration                    *prometheus.HistogramVec
	TripDistanceKm                  *prometheus.HistogramVec
	TripPlannedVsActualRatio        *prometheus.GaugeVec
	ActiveTripsTotal                *prometheus.GaugeVec
	TripsInProgress                 *prometheus.GaugeVec
	WaypointsVisitedTotal           *prometheus.CounterVec
	WaypointsSkippedTotal           *prometheus.CounterVec
	WaypointArrivalTimeVariance     *prometheus.HistogramVec
	TripsWithGoodDriverBehavior     *prometheus.CounterVec
	TripsWithHarshEvents            *prometheus.CounterVec
	IdlingTimeSeconds               *prometheus.CounterVec
	IdlingFuelWastedLiters          *prometheus.CounterVec
	SpeedingEventsTotal             *prometheus.CounterVec
	AverageTripSpeed                *prometheus.GaugeVec
	ClientCheckinsTotal             *prometheus.CounterVec
	ProofOfDeliveryCapturedTotal    *prometheus.CounterVec
	TripRevenueTotal                *prometheus.CounterVec
	TripCostTotal                   *prometheus.CounterVec
	TripProfitMargin                *prometheus.GaugeVec
	WaitingTimeSeconds              *prometheus.CounterVec
	ClientLateArrivalTime           *prometheus.CounterVec
	RouteDeviationDistanceKm        *prometheus.CounterVec
	RouteAdherenceScore             *prometheus.GaugeVec
	TripCustomerRating              *prometheus.GaugeVec
	DriverTripsCompleted            *prometheus.CounterVec
	DriverTripRating                *prometheus.GaugeVec
	FleetUtilizationRate            *prometheus.GaugeVec
	DeadheadDistanceKm              *prometheus.CounterVec
	DeadheadPercentage              *prometheus.GaugeVec
	TripsAutoAssignedTotal          *prometheus.CounterVec
	TripsManuallyAssignedTotal      *prometheus.CounterVec
	AssignmentAcceptanceRate        *prometheus.GaugeVec
	AssignmentRejectionRate         *prometheus.GaugeVec
	TimeWindowCompliance            *prometheus.GaugeVec
	EarlyArrivalBufferMinutes       *prometheus.HistogramVec
	MultiDropTripsTotal             *prometheus.CounterVec
	MultiDropEfficiency             *prometheus.GaugeVec
	TripStateTransitionsTotal       *prometheus.CounterVec
	TripStateDuration               *prometheus.HistogramVec
	CancellationPenaltyAppliedTotal *prometheus.CounterVec
}{
	TripsStartedTotal:               TripsStartedTotal,
	TripsCancelledTotal:             TripsCancelledTotal,
	TripsPausedTotal:                TripsPausedTotal,
	TripsResumedTotal:               TripsResumedTotal,
	TripDuration:                    TripDuration,
	TripDistanceKm:                  TripDistanceKm,
	TripPlannedVsActualRatio:        TripPlannedVsActualRatio,
	ActiveTripsTotal:                ActiveTripsTotal,
	TripsInProgress:                 TripsInProgress,
	WaypointsVisitedTotal:           WaypointsVisitedTotal,
	WaypointsSkippedTotal:           WaypointsSkippedTotal,
	WaypointArrivalTimeVariance:     WaypointArrivalTimeVariance,
	TripsWithGoodDriverBehavior:     TripsWithGoodDriverBehavior,
	TripsWithHarshEvents:            TripsWithHarshEvents,
	IdlingTimeSeconds:               IdlingTimeSeconds,
	IdlingFuelWastedLiters:          IdlingFuelWastedLiters,
	SpeedingEventsTotal:             SpeedingEventsTotal,
	AverageTripSpeed:                AverageTripSpeed,
	ClientCheckinsTotal:             ClientCheckinsTotal,
	ProofOfDeliveryCapturedTotal:    ProofOfDeliveryCapturedTotal,
	TripRevenueTotal:                TripRevenueTotal,
	TripCostTotal:                   TripCostTotal,
	TripProfitMargin:                TripProfitMargin,
	WaitingTimeSeconds:              WaitingTimeSeconds,
	ClientLateArrivalTime:           ClientLateArrivalTime,
	RouteDeviationDistanceKm:        RouteDeviationDistanceKm,
	RouteAdherenceScore:             RouteAdherenceScore,
	TripCustomerRating:              TripCustomerRating,
	DriverTripsCompleted:            DriverTripsCompleted,
	DriverTripRating:                DriverTripRating,
	FleetUtilizationRate:            FleetUtilizationRate,
	DeadheadDistanceKm:              DeadheadDistanceKm,
	DeadheadPercentage:              DeadheadPercentage,
	TripsAutoAssignedTotal:          TripsAutoAssignedTotal,
	TripsManuallyAssignedTotal:      TripsManuallyAssignedTotal,
	AssignmentAcceptanceRate:        AssignmentAcceptanceRate,
	AssignmentRejectionRate:         AssignmentRejectionRate,
	TimeWindowCompliance:            TimeWindowCompliance,
	EarlyArrivalBufferMinutes:       EarlyArrivalBufferMinutes,
	MultiDropTripsTotal:             MultiDropTripsTotal,
	MultiDropEfficiency:             MultiDropEfficiency,
	TripStateTransitionsTotal:       TripStateTransitionsTotal,
	TripStateDuration:               TripStateDuration,
	CancellationPenaltyAppliedTotal: CancellationPenaltyAppliedTotal,
}
