package models

import (
	"time"

	"github.com/google/uuid"
)

// PredictiveModel represents a trained ML model for predictive analytics
type PredictiveModel struct {
	ID                string    `json:"id"`
	OrganizationID    uuid.UUID `json:"organization_id" validate:"required"`
	Name              string    `json:"name" validate:"required,min=3,max=100"`
	Version           string    `json:"version" validate:"required,semver"`
	ModelType         string    `json:"model_type" validate:"required,oneof=regression classification clustering"`
	Description       string    `json:"description" validate:"max=500"`
	Algorithm         string    `json:"algorithm" validate:"required,oneof=random_forest neural_network svm xgboost linear_regression"`
	InputFeatures     []string  `json:"input_features" validate:"required,gt=0,dive,required"`
	OutputVariable    string    `json:"output_variable" validate:"required"`
	TrainingDataStart time.Time `json:"training_data_start" validate:"required,lte"`
	TrainingDataEnd   time.Time `json:"training_data_end" validate:"required,gtfield=TrainingDataStart"`
	Accuracy          float64   `json:"accuracy" validate:"gte=0,lte=1"`
	Precision         *float64  `json:"precision,omitempty" validate:"omitempty,gte=0,lte=1"`
	Recall            *float64  `json:"recall,omitempty" validate:"omitempty,gte=0,lte=1"`
	F1Score           *float64  `json:"f1_score,omitempty" validate:"omitempty,gte=0,lte=1"`
	Status            string    `json:"status" validate:"required,oneof=active inactive training error"`
	TrainedAt         time.Time `json:"trained_at" validate:"required"`
	ExpiresAt         time.Time `json:"expires_at" validate:"omitempty,gtfield=TrainedAt"`

	// Enhanced Registry Metadata (Phase 7)
	ModelArtifactURL   string    `json:"model_artifact_url"`
	ArtifactChecksum   string    `json:"artifact_checksum"`
	InferenceLatencyMS int       `json:"inference_latency_ms"`
	TrainingJobID      uuid.UUID `json:"training_job_id"`
	FeatureSchema      string    `json:"feature_schema"` // JSON string of expected features

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MLFeatureStore represents a reusable feature vector for ML (Phase 8: Versioned)
type MLFeatureStore struct {
	OrganizationID   uuid.UUID `json:"organization_id"`
	EntityType       string    `json:"entity_type"` // VEHICLE, DRIVER
	EntityID         uuid.UUID `json:"entity_id"`
	FeatureName      string    `json:"feature_name"`
	FeatureValue     float64   `json:"feature_value"`
	FeatureVersion   string    `json:"feature_version"` // e.g., v1, v2
	FeatureLineage   string    `json:"feature_lineage"` // Transformation logic ID or snippet (Phase 11)
	FeatureSource    string    `json:"feature_source"`  // e.g., stream, batch
	FeatureTimestamp time.Time `json:"feature_timestamp"`
}

// ModelDeployment represents a deployment strategy for a model (A/B, Shadow)
type ModelDeployment struct {
	ID                 string    `json:"id"`
	ModelID            string    `json:"model_id" validate:"required,uuid4"`
	OrganizationID     uuid.UUID `json:"organization_id" validate:"required"`
	DeploymentStrategy string    `json:"deployment_strategy" validate:"required,oneof=standard ab shadow"`
	TrafficSplit       float64   `json:"traffic_split" validate:"gte=0,lte=100"` // Percentage for A/B
	IsActive           bool      `json:"is_active"`
	DeployedAt         time.Time `json:"deployed_at"`
	CreatedAt          time.Time `json:"created_at"`
}

// DriverRiskEvent represents a specific risky driving event for coaching (Phase 8)
type DriverRiskEvent struct {
	ID             string    `json:"id"`
	DriverID       uuid.UUID `json:"driver_id" validate:"required"`
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	RiskType       string    `json:"risk_type" validate:"required"` // HARSH_BRAKE, OVERSPEED, etc.
	ScoreDelta     float64   `json:"score_delta"`
	Metadata       string    `json:"metadata"` // JSON string for lat/long/trip_id
	EventTime      time.Time `json:"event_time" validate:"required"`
	CreatedAt      time.Time `json:"created_at"`
}

// AnalyticsQuota tracks organization-level compute credits (ACU)
type AnalyticsQuota struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	DailyLimit     int64     `json:"daily_limit"`
	UsedCredits    int64     `json:"used_credits"`
	LastResetAt    time.Time `json:"last_reset_at"`
}

// PredictionResult stores the result of a model prediction
type PredictionResult struct {
	ID                string                 `json:"id"`
	ModelID           string                 `json:"model_id" validate:"required,uuid4"`
	PredictionInput   map[string]interface{} `json:"prediction_input" validate:"required"`
	PredictedValue    interface{}            `json:"predicted_value" validate:"required"`
	ConfidenceScore   float64                `json:"confidence_score" validate:"required,gte=0,lte=1"`
	ProbabilityScores map[string]float64     `json:"probability_scores,omitempty"`
	FeatureImportance map[string]float64     `json:"feature_importance,omitempty"`
	AnomalyDetected   bool                   `json:"anomaly_detected"`
	AnomalyScore      *float64               `json:"anomaly_score,omitempty" validate:"omitempty,gte=0,lte=1"`
	Explanation       string                 `json:"explanation,omitempty" validate:"max=1000"`
	OrganizationID    uuid.UUID              `json:"organization_id" validate:"required"`
	RequestID         string                 `json:"request_id" validate:"omitempty,uuid4"`
	PredictedAt       time.Time              `json:"predicted_at" validate:"required"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// AnalyticsQueryPlan represents a slow query plan for optimization (Phase 10)
type AnalyticsQueryPlan struct {
	ID                string                 `json:"id"`
	OrganizationID    uuid.UUID              `json:"organization_id"`
	QueryType         string                 `json:"query_type"`
	QueryParams       map[string]interface{} `json:"query_params"`
	ExecutionDuration int64                  `json:"execution_duration_ms"`
	QueryPlan         string                 `json:"query_plan"`
	CreatedAt         time.Time              `json:"created_at"`
}

// AnalyticsJob represents an analytics job (batch or real-time)
type AnalyticsJob struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name" validate:"required"`
	Type              string                 `json:"type" validate:"required,oneof=batch real_time streaming"`
	Description       string                 `json:"description"`
	Status            string                 `json:"status" validate:"required,oneof=pending running completed failed cancelled"`
	JobType           string                 `json:"job_type" validate:"required,oneof=prediction training evaluation inference"`
	Parameters        map[string]interface{} `json:"parameters"`
	InputSource       string                 `json:"input_source"`
	OutputDestination string                 `json:"output_destination"`
	StartTime         *time.Time             `json:"start_time,omitempty"`
	EndTime           *time.Time             `json:"end_time,omitempty"`
	Progress          float64                `json:"progress" validate:"gte=0,lte=100"`
	ErrorDetails      *string                `json:"error_details,omitempty"`
	ResultKey         *string                `json:"result_key,omitempty"` // Redis/Storage key for compressed results
	OrganizationID    uuid.UUID              `json:"organization_id" validate:"required"`
	IdempotencyKey    string                 `json:"idempotency_key,omitempty"` // For Phase 9 deduplication
	CreatedBy         uuid.UUID              `json:"created_by" validate:"required"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// FleetAnalytics represents aggregated analytics for fleet operations
type FleetAnalytics struct {
	ID                    string    `json:"id"`
	OrganizationID        uuid.UUID `json:"organization_id" validate:"required"`
	StartDate             time.Time `json:"start_date" validate:"required,lte"`
	EndDate               time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	TotalVehicles         int       `json:"total_vehicles" validate:"gte=0"`
	ActiveVehicles        int       `json:"active_vehicles" validate:"gte=0,ltefield=TotalVehicles"`
	TotalDistance         float64   `json:"total_distance" validate:"gte=0"`          // in km
	AverageSpeed          float64   `json:"average_speed" validate:"gte=0,lte=300"`   // in km/h
	TotalFuelConsumption  float64   `json:"total_fuel_consumption" validate:"gte=0"`  // in liters
	AverageFuelEfficiency float64   `json:"average_fuel_efficiency" validate:"gte=0"` // km/liter
	TotalTrips            int       `json:"total_trips" validate:"gte=0"`
	CompletedTrips        int       `json:"completed_trips" validate:"gte=0,ltefield=TotalTrips"`
	CancelledTrips        int       `json:"cancelled_trips" validate:"gte=0,ltefield=TotalTrips"`
	AverageTripDuration   float64   `json:"average_trip_duration" validate:"gte=0"` // in hours
	IdleTimePercentage    float64   `json:"idle_time_percentage" validate:"gte=0,lte=100"`
	RoutingEfficiency     float64   `json:"routing_efficiency"`                // percentage improvement over baseline
	FuelCost              float64   `json:"fuel_cost" validate:"gte=0"`        // in currency
	MaintenanceCost       float64   `json:"maintenance_cost" validate:"gte=0"` // in currency
	Revenue               float64   `json:"revenue" validate:"gte=0"`          // in currency
	ProfitMargin          float64   `json:"profit_margin"`                     // percentage
	SafetyScore           float64   `json:"safety_score" validate:"gte=0,lte=100"`
	AccidentsCount        int       `json:"accidents_count" validate:"gte=0"`
	IncidentsCount        int       `json:"incidents_count" validate:"gte=0"`
	EmergencyStops        int       `json:"emergency_stops" validate:"gte=0"`
	GeofenceViolations    int       `json:"geofence_violations" validate:"gte=0"`
	AlertsTriggered       int       `json:"alerts_triggered" validate:"gte=0"`
	DataPointsCollected   int64     `json:"data_points_collected" validate:"gte=0"`
	ProcessingTime        float64   `json:"processing_time" validate:"gte=0"` // in seconds
	GeneratedAt           time.Time `json:"generated_at" validate:"required"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// DriverBehaviorAnalytics represents driver behavior analysis
type DriverBehaviorAnalytics struct {
	ID                          string    `json:"id"`
	DriverID                    uuid.UUID `json:"driver_id" validate:"required"`
	OrganizationID              uuid.UUID `json:"organization_id" validate:"required"`
	AnalysisPeriodStart         time.Time `json:"analysis_period_start" validate:"required"`
	AnalysisPeriodEnd           time.Time `json:"analysis_period_end" validate:"required"`
	TotalDrivingHours           float64   `json:"total_driving_hours" validate:"gte=0"`
	AggressiveAccelerationCount int       `json:"aggressive_acceleration_count"`
	HardBrakingCount            int       `json:"hard_braking_count"`
	SharpTurnCount              int       `json:"sharp_turn_count"`
	OverSpeedingCount           int       `json:"over_speeding_count"`
	IdlingTime                  float64   `json:"idling_time" validate:"gte=0"`     // in hours
	AverageSpeed                float64   `json:"average_speed" validate:"gte=0"`   // in km/h
	MaxSpeed                    float64   `json:"max_speed" validate:"gte=0"`       // in km/h
	FuelEfficiency              float64   `json:"fuel_efficiency" validate:"gte=0"` // km/liter
	SafetyScore                 float64   `json:"safety_score" validate:"gte=0,lte=100"`
	EfficiencyScore             float64   `json:"efficiency_score" validate:"gte=0,lte=100"`
	ComfortScore                float64   `json:"comfort_score" validate:"gte=0,lte=100"`
	Ranking                     int       `json:"ranking" validate:"gte=1"`
	FatigueEventCount           int       `json:"fatigue_event_count"`
	ViolationCount              int       `json:"violation_count"`
	RiskLevel                   string    `json:"risk_level"` // Safe, Moderate, Risky, Dangerous
	ImprovementSuggestions      []string  `json:"improvement_suggestions"`
	BehaviorTrend               string    `json:"behavior_trend" validate:"omitempty,oneof=improving declining stable"`
	GeneratedAt                 time.Time `json:"generated_at"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

// PredictiveMaintenanceAnalytics represents predictive maintenance insights
type PredictiveMaintenanceAnalytics struct {
	ID                  string                     `json:"id"`
	VehicleID           uuid.UUID                  `json:"vehicle_id" validate:"required"`
	OrganizationID      uuid.UUID                  `json:"organization_id" validate:"required"`
	AnalysisDate        time.Time                  `json:"analysis_date" validate:"required"`
	NextServiceDue      time.Time                  `json:"next_service_due"`
	ServicePriority     string                     `json:"service_priority" validate:"omitempty,oneof=high medium low"`
	ComponentRisks      map[string]float64         `json:"component_risks"`
	FailurePrediction   *FailurePrediction         `json:"failure_prediction,omitempty"`
	MaintenanceSchedule []MaintenanceScheduleEntry `json:"maintenance_schedule"`
	PartsNeeded         []PartRequirement          `json:"parts_needed"`
	CostEstimation      float64                    `json:"cost_estimation" validate:"gte=0"`
	UptimePrediction    float64                    `json:"uptime_prediction" validate:"gte=0,lte=100"`
	ReliabilityScore    float64                    `json:"reliability_score" validate:"gte=0,lte=100"`
	Recommendations     []string                   `json:"recommendations"`
	GeneratedAt         time.Time                  `json:"generated_at"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
}

// FailurePrediction represents a predicted component failure
type FailurePrediction struct {
	Component            string    `json:"component" validate:"required"`
	PredictedFailureDate time.Time `json:"predicted_failure_date" validate:"required"`
	ConfidenceScore      float64   `json:"confidence_score" validate:"gte=0,lte=1"`
	UrgencyLevel         string    `json:"urgency_level" validate:"required,oneof=critical high medium low"`
	ExpectedDowntime     float64   `json:"expected_downtime" validate:"gte=0"`
	ExpectedCost         float64   `json:"expected_cost" validate:"gte=0"`
	Description          string    `json:"description"`
}

// MaintenanceScheduleEntry represents a scheduled maintenance entry
type MaintenanceScheduleEntry struct {
	ID             string     `json:"id"`
	Task           string     `json:"task" validate:"required"`
	DueDate        time.Time  `json:"due_date" validate:"required"`
	Priority       string     `json:"priority" validate:"required,oneof=high medium low"`
	EstimatedHours float64    `json:"estimated_hours" validate:"gte=0"`
	CostEstimate   float64    `json:"cost_estimate" validate:"gte=0"`
	Dependencies   []string   `json:"dependencies"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"`
	Status         string     `json:"status" validate:"required,oneof=planned in_progress completed cancelled"`
}

// PartRequirement represents a part that may be needed for maintenance
type PartRequirement struct {
	PartNumber    string  `json:"part_number" validate:"required"`
	Description   string  `json:"description"`
	Quantity      int     `json:"quantity" validate:"gte=1"`
	EstimatedCost float64 `json:"estimated_cost" validate:"gte=0"`
	Availability  string  `json:"availability" validate:"omitempty,oneof=in_stock low_stock out_of_stock"`
	Supplier      string  `json:"supplier"`
	LeadTimeDays  int     `json:"lead_time_days" validate:"gte=0"`
}

// UpdatePredictiveModelRequest represents partial updates for a predictive model
type UpdatePredictiveModelRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Version     *string `json:"version,omitempty" validate:"omitempty,semver"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive training error"`
}

// UpdatePredictionResultRequest represents partial updates for a prediction result
type UpdatePredictionResultRequest struct {
	PredictedValue  interface{} `json:"predicted_value,omitempty"`
	ConfidenceScore *float64    `json:"confidence_score,omitempty" validate:"omitempty,gte=0,lte=1"`
	AnomalyDetected *bool       `json:"anomaly_detected,omitempty"`
	AnomalyScore    *float64    `json:"anomaly_score,omitempty" validate:"omitempty,gte=0,lte=1"`
	Explanation     *string     `json:"explanation,omitempty" validate:"omitempty,max=1000"`
}

// UpdateAnalyticsJobRequest represents partial updates for an analytics job
type UpdateAnalyticsJobRequest struct {
	Name         *string                `json:"name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Status       *string                `json:"status,omitempty" validate:"omitempty,oneof=pending running completed failed cancelled"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Progress     *float64               `json:"progress,omitempty" validate:"omitempty,gte=0,lte=100"`
	ErrorDetails *string                `json:"error_details,omitempty"`
}

// UpdateFleetAnalyticsRequest represents partial updates for fleet analytics
type UpdateFleetAnalyticsRequest struct {
	TotalVehicles         *int     `json:"total_vehicles,omitempty" validate:"omitempty,gte=0"`
	ActiveVehicles        *int     `json:"active_vehicles,omitempty" validate:"omitempty,gte=0"`
	TotalDistance         *float64 `json:"total_distance,omitempty" validate:"omitempty,gte=0"`
	AverageSpeed          *float64 `json:"average_speed,omitempty" validate:"omitempty,gte=0,lte=300"`
	TotalFuelConsumption  *float64 `json:"total_fuel_consumption,omitempty" validate:"omitempty,gte=0"`
	AverageFuelEfficiency *float64 `json:"average_fuel_efficiency,omitempty" validate:"omitempty,gte=0"`
	SafetyScore           *float64 `json:"safety_score,omitempty" validate:"omitempty,gte=0,lte=100"`
}

// UpdateDriverBehaviorAnalyticsRequest represents partial updates for driver behavior analytics
type UpdateDriverBehaviorAnalyticsRequest struct {
	SafetyScore            *float64 `json:"safety_score,omitempty" validate:"omitempty,gte=0,lte=100"`
	EfficiencyScore        *float64 `json:"efficiency_score,omitempty" validate:"omitempty,gte=0,lte=100"`
	ComfortScore           *float64 `json:"comfort_score,omitempty" validate:"omitempty,gte=0,lte=100"`
	ImprovementSuggestions []string `json:"improvement_suggestions,omitempty"`
	BehaviorTrend          *string  `json:"behavior_trend,omitempty" validate:"omitempty,oneof=improving declining stable"`
}

// UpdatePredictiveMaintenanceAnalyticsRequest represents partial updates for predictive maintenance
type UpdatePredictiveMaintenanceAnalyticsRequest struct {
	NextServiceDue   *time.Time `json:"next_service_due,omitempty"`
	ServicePriority  *string    `json:"service_priority,omitempty" validate:"omitempty,oneof=high medium low"`
	CostEstimation   *float64   `json:"cost_estimation,omitempty" validate:"omitempty,gte=0"`
	ReliabilityScore *float64   `json:"reliability_score,omitempty" validate:"omitempty,gte=0,lte=100"`
	Recommendations  []string   `json:"recommendations,omitempty"`
}

// VehicleDailyStats represents pre-aggregated daily metrics for a vehicle
type VehicleDailyStats struct {
	VehicleID              uuid.UUID `json:"vehicle_id"`
	OrganizationID         uuid.UUID `json:"organization_id"`
	Date                   time.Time `json:"date"`
	TotalDistance          float64   `json:"total_distance"`
	TotalFuelConsumed      float64   `json:"total_fuel_consumed"`
	TotalEngineHours       float64   `json:"total_engine_hours"`
	TotalIdleHours         float64   `json:"total_idle_hours"`
	MaxSpeed               float64   `json:"max_speed"`
	AvgSpeed               float64   `json:"avg_speed"`
	HarshBrakingCount      int       `json:"harsh_braking_count"`
	RapidAccelerationCount int       `json:"rapid_acceleration_count"`
	SharpTurnCount         int       `json:"sharp_turn_count"`
	OverspeedingCount      int       `json:"overspeeding_count"`
	UpdatedAt              time.Time `json:"updated_at"`
	LastRefreshedAt        time.Time `json:"last_refreshed_at"`
}

// DriverDailyStats represents pre-aggregated daily metrics for a driver
type DriverDailyStats struct {
	DriverID          uuid.UUID `json:"driver_id"`
	OrganizationID    uuid.UUID `json:"organization_id"`
	Date              time.Time `json:"date"`
	TotalDrivingHours float64   `json:"total_driving_hours"`
	TotalDistance     float64   `json:"total_distance"`
	TotalIdleHours    float64   `json:"total_idle_hours"`
	SafetyScore       float64   `json:"safety_score"`
	EcoScore          float64   `json:"eco_score"`
	ViolationCount    int       `json:"violation_count"`
	FatigueEventCount int       `json:"fatigue_event_count"`
	UpdatedAt         time.Time `json:"updated_at"`
	LastRefreshedAt   time.Time `json:"last_refreshed_at"`
}

// FleetHourlyStats represents pre-aggregated hourly metrics for the fleet
type FleetHourlyStats struct {
	OrganizationID     uuid.UUID `json:"organization_id"`
	Timestamp          time.Time `json:"timestamp"`
	ActiveVehicles     int       `json:"active_vehicles"`
	TotalDistance      float64   `json:"total_distance"`
	TotalFuelConsumed  float64   `json:"total_fuel_consumed"`
	AverageUtilization float64   `json:"average_utilization"`
	UpdatedAt          time.Time `json:"updated_at"`
	LastRefreshedAt    time.Time `json:"last_refreshed_at"`
}

// AnalyticsQueryLog represents an audit log entry for an analytics query
type AnalyticsQueryLog struct {
	ID             uuid.UUID              `json:"id"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
	QueryType      string                 `json:"query_type"`
	Parameters     map[string]interface{} `json:"parameters"`
	DurationMs     int64                  `json:"duration_ms"`
	RowCount       int                    `json:"row_count"`
	CacheHit       bool                   `json:"cache_hit"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// AnalyticsAccessLog provides high-fidelity tracking of analytics usage (Phase 11)
type AnalyticsAccessLog struct {
	ID             uuid.UUID              `json:"id"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	UserID         uuid.UUID              `json:"user_id"`
	ResourceID     string                 `json:"resource_id"`   // e.g., model_id, report_id
	ResourceType   string                 `json:"resource_type"` // MODEL, REPORT, KPI
	AccessContext  map[string]interface{} `json:"access_context"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	CreatedAt      time.Time              `json:"created_at"`
}

// MaterializedFleetReport stores summarized data for executive dashboards (Phase 11)
type MaterializedFleetReport struct {
	ID             string    `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ReportType     string    `json:"report_type"` // MONTHLY_SUMMARY, QUARTERLY_RISK
	MetricKey      string    `json:"metric_key"`
	MetricValue    float64   `json:"metric_value"`
	Period         string    `json:"period"` // 2026-03
	GeneratedAt    time.Time `json:"generated_at"`
}
