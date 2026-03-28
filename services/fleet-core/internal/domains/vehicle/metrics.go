package vehicle

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Vehicle lifecycle metrics
	VehiclesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicles_total",
		Help: "Total vehicles in fleet",
	}, []string{"org_id", "vehicle_type", "status"})

	VehiclesActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicles_active_total",
		Help: "Currently active vehicles (running trips)",
	}, []string{"org_id", "vehicle_type"})

	VehiclesInactive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicles_inactive_total",
		Help: "Currently inactive vehicles",
	}, []string{"org_id", "vehicle_type", "inactive_reason"})

	VehicleRegistrationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_registrations_total",
		Help: "Total vehicle registrations",
	}, []string{"org_id", "vehicle_type"})

	VehicleDeregistrationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_deregistrations_total",
		Help: "Total vehicle deregistrations",
	}, []string{"org_id", "deregistration_reason"})

	// Telemetry metrics
	VehicleTelemetryMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_telemetry_messages_total",
		Help: "Total telemetry messages received",
	}, []string{"org_id", "vehicle_id", "message_type"})

	VehicleTelemetryLostTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_telemetry_lost_total",
		Help: "Telemetry messages lost/dropped",
	}, []string{"org_id", "vehicle_id", "loss_reason"})

	VehicleTelemetryFrequency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_telemetry_frequency_hz",
		Help: "Actual telemetry message frequency per vehicle",
	}, []string{"org_id", "vehicle_id"})

	// Health metrics
	VehicleHealthScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_health_score",
		Help: "Vehicle health score (0-100)",
	}, []string{"org_id", "vehicle_id"})

	VehicleFaultCodesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_fault_codes_total",
		Help: "OBD fault codes detected",
	}, []string{"org_id", "vehicle_id", "fault_code", "severity"})

	CheckEngineAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_check_engine_alerts_total",
		Help: "Check engine light alerts",
	}, []string{"org_id", "vehicle_id", "alert_type"})

	// Location metrics
	VehicleLocationUpdatesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_location_updates_total",
		Help: "GPS location updates",
	}, []string{"org_id", "vehicle_id"})

	VehicleGpsFixQuality = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_gps_fix_quality",
		Help: "GPS fix quality (0-5, 5=best)",
	}, []string{"org_id", "vehicle_id"})

	// Fuel metrics
	VehicleFuelLevel = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_fuel_level_percentage",
		Help: "Vehicle fuel level percentage",
	}, []string{"org_id", "vehicle_id"})

	VehicleFuelConsumptionRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_fuel_consumption_rate_lph",
		Help: "Fuel consumption rate in liters/hour",
	}, []string{"org_id", "vehicle_id"})

	VehicleFuelEfficiency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_fuel_efficiency_kmpl",
		Help: "Fuel efficiency in km/l",
	}, []string{"org_id", "vehicle_id"})

	LowFuelAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_low_fuel_alerts_total",
		Help: "Low fuel level alerts",
	}, []string{"org_id", "vehicle_id"})

	// Odometer metrics
	VehicleOdometerKm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_odometer_km",
		Help: "Vehicle odometer reading in km",
	}, []string{"org_id", "vehicle_id"})

	// Engine metrics
	VehicleEngineRPM = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_engine_rpm",
		Help: "Engine RPM",
	}, []string{"org_id", "vehicle_id"})

	VehicleEngineTemperature = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_engine_temperature_celsius",
		Help: "Engine coolant temperature",
	}, []string{"org_id", "vehicle_id"})

	VehicleSpeedKmh = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_speed_kmh",
		Help: "Vehicle speed in km/h",
	}, []string{"org_id", "vehicle_id"})

	// Battery metrics
	VehicleBatteryVoltage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_battery_voltage",
		Help: "Vehicle battery voltage",
	}, []string{"org_id", "vehicle_id"})

	LowBatteryAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_low_battery_alerts_total",
		Help: "Low vehicle battery alerts",
	}, []string{"org_id", "vehicle_id"})

	// Tire metrics
	TirePressurePSI = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_tire_pressure_psi",
		Help: "Tire pressure in PSI",
	}, []string{"org_id", "vehicle_id", "position"})

	LowTirePressureAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_low_tire_pressure_alerts_total",
		Help: "Low tire pressure alerts",
	}, []string{"org_id", "vehicle_id", "position"})

	// Driver-vehicle assignment
	DriverVehicleAssignmentsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "driver_vehicle_assignments_total",
		Help: "Active driver-vehicle assignments",
	}, []string{"org_id"})

	DriverVehicleSwapTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "driver_vehicle_swap_total",
		Help: "Driver-vehicle reassignments",
	}, []string{"org_id", "swap_reason"})

	// Immobilizer metrics
	ImmobilizerCommandsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_immobilizer_commands_total",
		Help: "Immobilizer/remote immobilization commands",
	}, []string{"org_id", "vehicle_id", "command", "result"})

	ImmobilizerCommandLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vehicle_immobilizer_command_latency_seconds",
		Help:    "Immobilizer command acknowledgement latency",
		Buckets: prometheus.ExponentialBuckets(0.5, 2, 8),
	}, []string{"org_id"})

	// Theft/recovery metrics
	TheftAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_theft_alerts_total",
		Help: "Theft/unauthorized movement alerts",
	}, []string{"org_id", "alert_type"})

	RecoveredVehiclesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicles_recovered_total",
		Help: "Stolen vehicles recovered",
	}, []string{"org_id", "recovery_type"})

	// History metrics
	VehicleHistoryLookupsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_history_lookups_total",
		Help: "Historical vehicle data queries",
	}, []string{"org_id", "lookup_type"})

	HistoricalDataRetentionDays = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_historical_data_retention_days",
		Help: "Historical data retention period in days",
	}, []string{"org_id", "data_type"})

	// Compliance metrics
	VehicleRCExpiryAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_rc_expiry_alerts_total",
		Help: "Registration certificate expiry alerts",
	}, []string{"org_id", "days_to_expiry_bucket"})

	VehicleInsuranceExpiryAlerts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_insurance_expiry_alerts_total",
		Help: "Insurance expiry alerts",
	}, []string{"org_id", "days_to_expiry_bucket"})

	VehiclePUCExpiryAlerts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_puc_expiry_alerts_total",
		Help: "PUC certificate expiry alerts",
	}, []string{"org_id", "days_to_expiry_bucket"})

	// Service history
	ServiceIntervalsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_service_intervals_completed_total",
		Help: "Scheduled service intervals completed",
	}, []string{"org_id", "service_type"})

	ServiceIntervalsMissed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vehicle_service_intervals_missed_total",
		Help: "Scheduled service intervals missed",
	}, []string{"org_id", "service_type"})

	// Tag/FASTag metrics
	VehicleFASTagCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_fastag_count",
		Help: "FASTags assigned per vehicle",
	}, []string{"org_id", "vehicle_id"})

	// GPS tracking state
	VehicleGPSTrackingEnabled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vehicle_gps_tracking_enabled_percentage",
		Help: "Percentage of vehicles with GPS tracking enabled",
	}, []string{"org_id"})

	VehicleOfflineDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vehicle_offline_duration_seconds",
		Help:    "Duration vehicles are offline",
		Buckets: prometheus.ExponentialBuckets(60, 2, 10),
	}, []string{"org_id", "offline_reason"})
)

// Global metrics instance
var GlobalMetrics = struct {
	VehiclesTotal                 *prometheus.GaugeVec
	VehiclesActive                *prometheus.GaugeVec
	VehiclesInactive              *prometheus.GaugeVec
	VehicleRegistrationsTotal     *prometheus.CounterVec
	VehicleDeregistrationsTotal   *prometheus.CounterVec
	VehicleTelemetryMessagesTotal *prometheus.CounterVec
	VehicleTelemetryLostTotal     *prometheus.CounterVec
	VehicleTelemetryFrequency     *prometheus.GaugeVec
	VehicleHealthScore            *prometheus.GaugeVec
	VehicleFaultCodesTotal        *prometheus.CounterVec
	CheckEngineAlertsTotal        *prometheus.CounterVec
	VehicleLocationUpdatesTotal   *prometheus.CounterVec
	VehicleGpsFixQuality          *prometheus.GaugeVec
	VehicleFuelLevel              *prometheus.GaugeVec
	VehicleFuelConsumptionRate    *prometheus.GaugeVec
	VehicleFuelEfficiency         *prometheus.GaugeVec
	LowFuelAlertsTotal            *prometheus.CounterVec
	VehicleOdometerKm             *prometheus.GaugeVec
	VehicleEngineRPM              *prometheus.GaugeVec
	VehicleEngineTemperature      *prometheus.GaugeVec
	VehicleSpeedKmh               *prometheus.GaugeVec
	VehicleBatteryVoltage         *prometheus.GaugeVec
	LowBatteryAlertsTotal         *prometheus.CounterVec
	TirePressurePSI               *prometheus.GaugeVec
	LowTirePressureAlertsTotal    *prometheus.CounterVec
	DriverVehicleAssignmentsTotal *prometheus.GaugeVec
	DriverVehicleSwapTotal        *prometheus.CounterVec
	ImmobilizerCommandsTotal      *prometheus.CounterVec
	ImmobilizerCommandLatency     *prometheus.HistogramVec
	TheftAlertsTotal              *prometheus.CounterVec
	RecoveredVehiclesTotal        *prometheus.CounterVec
	VehicleHistoryLookupsTotal    *prometheus.CounterVec
	HistoricalDataRetentionDays   *prometheus.GaugeVec
	VehicleRCExpiryAlertsTotal    *prometheus.CounterVec
	VehicleInsuranceExpiryAlerts  *prometheus.CounterVec
	VehiclePUCExpiryAlerts        *prometheus.CounterVec
	ServiceIntervalsCompleted     *prometheus.CounterVec
	ServiceIntervalsMissed        *prometheus.CounterVec
	VehicleFASTagCount            *prometheus.GaugeVec
	VehicleGPSTrackingEnabled     *prometheus.GaugeVec
	VehicleOfflineDuration        *prometheus.HistogramVec
}{
	VehiclesTotal:                 VehiclesTotal,
	VehiclesActive:                VehiclesActive,
	VehiclesInactive:              VehiclesInactive,
	VehicleRegistrationsTotal:     VehicleRegistrationsTotal,
	VehicleDeregistrationsTotal:   VehicleDeregistrationsTotal,
	VehicleTelemetryMessagesTotal: VehicleTelemetryMessagesTotal,
	VehicleTelemetryLostTotal:     VehicleTelemetryLostTotal,
	VehicleTelemetryFrequency:     VehicleTelemetryFrequency,
	VehicleHealthScore:            VehicleHealthScore,
	VehicleFaultCodesTotal:        VehicleFaultCodesTotal,
	CheckEngineAlertsTotal:        CheckEngineAlertsTotal,
	VehicleLocationUpdatesTotal:   VehicleLocationUpdatesTotal,
	VehicleGpsFixQuality:          VehicleGpsFixQuality,
	VehicleFuelLevel:              VehicleFuelLevel,
	VehicleFuelConsumptionRate:    VehicleFuelConsumptionRate,
	VehicleFuelEfficiency:         VehicleFuelEfficiency,
	LowFuelAlertsTotal:            LowFuelAlertsTotal,
	VehicleOdometerKm:             VehicleOdometerKm,
	VehicleEngineRPM:              VehicleEngineRPM,
	VehicleEngineTemperature:      VehicleEngineTemperature,
	VehicleSpeedKmh:               VehicleSpeedKmh,
	VehicleBatteryVoltage:         VehicleBatteryVoltage,
	LowBatteryAlertsTotal:         LowBatteryAlertsTotal,
	TirePressurePSI:               TirePressurePSI,
	LowTirePressureAlertsTotal:    LowTirePressureAlertsTotal,
	DriverVehicleAssignmentsTotal: DriverVehicleAssignmentsTotal,
	DriverVehicleSwapTotal:        DriverVehicleSwapTotal,
	ImmobilizerCommandsTotal:      ImmobilizerCommandsTotal,
	ImmobilizerCommandLatency:     ImmobilizerCommandLatency,
	TheftAlertsTotal:              TheftAlertsTotal,
	RecoveredVehiclesTotal:        RecoveredVehiclesTotal,
	VehicleHistoryLookupsTotal:    VehicleHistoryLookupsTotal,
	HistoricalDataRetentionDays:   HistoricalDataRetentionDays,
	VehicleRCExpiryAlertsTotal:    VehicleRCExpiryAlertsTotal,
	VehicleInsuranceExpiryAlerts:  VehicleInsuranceExpiryAlerts,
	VehiclePUCExpiryAlerts:        VehiclePUCExpiryAlerts,
	ServiceIntervalsCompleted:     ServiceIntervalsCompleted,
	ServiceIntervalsMissed:        ServiceIntervalsMissed,
	VehicleFASTagCount:            VehicleFASTagCount,
	VehicleGPSTrackingEnabled:     VehicleGPSTrackingEnabled,
	VehicleOfflineDuration:        VehicleOfflineDuration,
}
