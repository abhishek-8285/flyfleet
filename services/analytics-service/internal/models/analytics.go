package models

import (
	"time"

	"github.com/google/uuid"
)

// PredictiveModel represents a trained ML model for predictive analytics
type PredictiveModel struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Version           string    `json:"version"`
	ModelType         string    `json:"model_type"`         // regression, classification, clustering
	Description       string    `json:"description"`
	Algorithm         string    `json:"algorithm"`          // random_forest, neural_network, svm, etc.
	InputFeatures     []string  `json:"input_features"`     // list of feature names
	OutputVariable    string    `json:"output_variable"`    // target variable name
	TrainingDataStart time.Time `json:"training_data_start"`
	TrainingDataEnd   time.Time `json:"training_data_end"`
	Accuracy          float64   `json:"accuracy"`           // model accuracy score
	Precision         *float64  `json:"precision,omitempty"` // precision score
	Recall            *float64  `json:"recall,omitempty"`    // recall score
	F1Score           *float64  `json:"f1_score,omitempty"`  // F1 score
	Status            string    `json:"status"`              // active, inactive, training, error
	TrainedAt         time.Time `json:"trained_at"`
	ExpiresAt         time.Time `json:"expires_at"`          // when model should be retrained
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by"`
}

// PredictionResult stores the result of a model prediction
type PredictionResult struct {
	ID                  string                 `json:"id"`
	ModelID             string                 `json:"model_id"`
	PredictionInput     map[string]interface{} `json:"prediction_input"` // input features used for prediction
	PredictedValue      interface{}            `json:"predicted_value"`  // predicted outcome
	ConfidenceScore     float64                `json:"confidence_score"` // confidence in prediction
	ProbabilityScores   map[string]float64     `json:"probability_scores,omitempty"` // for classification models
	FeatureImportance   map[string]float64     `json:"feature_importance,omitempty"` // importance of each feature
	AnomalyDetected     bool                   `json:"anomaly_detected"`             // for anomaly detection
	AnomalyScore        *float64               `json:"anomaly_score,omitempty"`      // anomaly score if detected
	Explanation         string                 `json:"explanation,omitempty"`        // explanation of prediction
	OrganizationID      uuid.UUID              `json:"organization_id"`
	RequestID           string                 `json:"request_id"`     // ID of the prediction request
	PredictedAt         time.Time              `json:"predicted_at"`
	CreatedAt           time.Time              `json:"created_at"`
	CreatedBy           uuid.UUID              `json:"created_by"`
	UpdatedAt           time.Time              `json:"updated_at"`
	UpdatedBy           uuid.UUID              `json:"updated_by"`
}

// AnalyticsJob represents an analytics job (batch or real-time)
type AnalyticsJob struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`              // batch, real_time, streaming
	Description      string    `json:"description"`
	Status           string    `json:"status"`            // pending, running, completed, failed, cancelled
	JobType          string    `json:"job_type"`          // prediction, training, evaluation, inference
	Parameters       map[string]interface{} `json:"parameters"` // job-specific parameters
	InputSource      string    `json:"input_source"`      // data source for the job
	OutputDestination string   `json:"output_destination"` // destination for results
	StartTime        *time.Time `json:"start_time,omitempty"`
	EndTime          *time.Time `json:"end_time,omitempty"`
	Progress         float64   `json:"progress"`          // percentage completion
	ErrorDetails     *string   `json:"error_details,omitempty"`
	OrganizationID   uuid.UUID `json:"organization_id"`
	CreatedBy        uuid.UUID `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	UpdatedBy        uuid.UUID `json:"updated_by"`
}

// FleetAnalytics represents aggregated analytics for fleet operations
type FleetAnalytics struct {
	ID                    string    `json:"id"`
	OrganizationID        uuid.UUID `json:"organization_id"`
	StartDate             time.Time `json:"start_date"`
	EndDate               time.Time `json:"end_date"`
	TotalVehicles         int       `json:"total_vehicles"`
	ActiveVehicles        int       `json:"active_vehicles"`
	TotalDistance         float64   `json:"total_distance"`          // in km
	AverageSpeed          float64   `json:"average_speed"`           // in km/h
	TotalFuelConsumption  float64   `json:"total_fuel_consumption"`  // in liters
	AverageFuelEfficiency float64   `json:"average_fuel_efficiency"` // km/liter
	TotalTrips            int       `json:"total_trips"`
	CompletedTrips        int       `json:"completed_trips"`
	CancelledTrips        int       `json:"cancelled_trips"`
	AverageTripDuration   float64   `json:"average_trip_duration"`   // in hours
	IdleTimePercentage    float64   `json:"idle_time_percentage"`    // percentage of idle time
	RoutingEfficiency     float64   `json:"routing_efficiency"`      // percentage improvement over baseline
	FuelCost              float64   `json:"fuel_cost"`               // in currency
	MaintenanceCost       float64   `json:"maintenance_cost"`        // in currency
	Revenue               float64   `json:"revenue"`                 // in currency
	ProfitMargin          float64   `json:"profit_margin"`           // percentage
	SafetyScore           float64   `json:"safety_score"`            // safety rating
	AccidentsCount        int       `json:"accidents_count"`
	IncidentsCount        int       `json:"incidents_count"`
	EmergencyStops        int       `json:"emergency_stops"`
	GeofenceViolations    int       `json:"geofence_violations"`
	AlertsTriggered       int       `json:"alerts_triggered"`
	DataPointsCollected   int64     `json:"data_points_collected"`
	ProcessingTime        float64   `json:"processing_time"`         // in seconds
	GeneratedAt           time.Time `json:"generated_at"`
	CreatedAt             time.Time `json:"created_at"`
	CreatedBy             uuid.UUID `json:"created_by"`
	UpdatedAt             time.Time `json:"updated_at"`
	UpdatedBy             uuid.UUID `json:"updated_by"`
}

// DriverBehaviorAnalytics represents driver behavior analysis
type DriverBehaviorAnalytics struct {
	ID                  string    `json:"id"`
	DriverID            uuid.UUID `json:"driver_id"`
	OrganizationID      uuid.UUID `json:"organization_id"`
	AnalysisPeriodStart time.Time `json:"analysis_period_start"`
	AnalysisPeriodEnd   time.Time `json:"analysis_period_end"`
	TotalDrivingHours   float64   `json:"total_driving_hours"`
	AggressiveAccelerationCount int `json:"aggressive_acceleration_count"`
	HardBrakingCount          int `json:"hard_braking_count"`
	SharpTurnCount            int `json:"sharp_turn_count"`
	OverSpeedingCount         int `json:"over_speeding_count"`
	IdlingTime                float64 `json:"idling_time"`           // in hours
	AverageSpeed              float64 `json:"average_speed"`           // in km/h
	MaxSpeed                  float64 `json:"max_speed"`               // in km/h
	FuelEfficiency            float64 `json:"fuel_efficiency"`         // km/liter
	SafetyScore               float64 `json:"safety_score"`            // 0-100 scale
	EfficiencyScore           float64 `json:"efficiency_score"`        // 0-100 scale
	ComfortScore              float64 `json:"comfort_score"`           // 0-100 scale
	Ranking                   int     `json:"ranking"`                 // rank among drivers
	ImprovementSuggestions    []string `json:"improvement_suggestions"` // suggestions for improvement
	BehaviorTrend             string  `json:"behavior_trend"`          // improving, declining, stable
	GeneratedAt               time.Time `json:"generated_at"`
	CreatedAt                 time.Time `json:"created_at"`
	CreatedBy                 uuid.UUID `json:"created_by"`
	UpdatedAt                 time.Time `json:"updated_at"`
	UpdatedBy                 uuid.UUID `json:"updated_by"`
}

// PredictiveMaintenanceAnalytics represents predictive maintenance insights
type PredictiveMaintenanceAnalytics struct {
	ID                    string    `json:"id"`
	VehicleID             uuid.UUID `json:"vehicle_id"`
	OrganizationID        uuid.UUID `json:"organization_id"`
	AnalysisDate          time.Time `json:"analysis_date"`
	NextServiceDue        time.Time `json:"next_service_due"`
	ServicePriority       string    `json:"service_priority"`        // high, medium, low
	ComponentRisks        map[string]float64 `json:"component_risks"` // risk scores for each component
	FailurePrediction     *FailurePrediction `json:"failure_prediction,omitempty"` // predicted failures
	MaintenanceSchedule   []MaintenanceScheduleEntry `json:"maintenance_schedule"` // recommended schedule
	PartsNeeded           []PartRequirement `json:"parts_needed"` // parts that may be needed
	CostEstimation        float64   `json:"cost_estimation"`     // estimated cost in currency
	UptimePrediction      float64   `json:"uptime_prediction"`   // predicted uptime percentage
	ReliabilityScore      float64   `json:"reliability_score"`   // reliability rating
	Recommendations       []string  `json:"recommendations"`     // maintenance recommendations
	GeneratedAt           time.Time `json:"generated_at"`
	CreatedAt             time.Time `json:"created_at"`
	CreatedBy             uuid.UUID `json:"created_by"`
	UpdatedAt             time.Time `json:"updated_at"`
	UpdatedBy             uuid.UUID `json:"updated_by"`
}

// FailurePrediction represents a predicted component failure
type FailurePrediction struct {
	Component           string    `json:"component"`           // name of the component
	PredictedFailureDate time.Time `json:"predicted_failure_date"` // predicted date of failure
	ConfidenceScore     float64   `json:"confidence_score"`    // confidence in prediction
	UrgencyLevel        string    `json:"urgency_level"`       // critical, high, medium, low
	ExpectedDowntime    float64   `json:"expected_downtime"`   // expected downtime in hours
	ExpectedCost        float64   `json:"expected_cost"`       // expected cost in currency
	Description         string    `json:"description"`         // description of potential failure
}

// MaintenanceScheduleEntry represents a scheduled maintenance entry
type MaintenanceScheduleEntry struct {
	ID             string    `json:"id"`
	Task           string    `json:"task"`              // description of maintenance task
	DueDate        time.Time `json:"due_date"`          // when task is due
	Priority       string    `json:"priority"`          // high, medium, low
	EstimatedHours float64   `json:"estimated_hours"`   // estimated time to complete
	CostEstimate   float64   `json:"cost_estimate"`     // estimated cost
	Dependencies   []string  `json:"dependencies"`      // other tasks this depends on
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"` // assigned technician
	Status         string    `json:"status"`            // planned, in_progress, completed, cancelled
}

// PartRequirement represents a part that may be needed for maintenance
type PartRequirement struct {
	PartNumber    string  `json:"part_number"`
	Description   string  `json:"description"`
	Quantity      int     `json:"quantity"`
	EstimatedCost float64 `json:"estimated_cost"` // per unit cost
	Availability  string  `json:"availability"`   // in_stock, low_stock, out_of_stock
	Supplier      string  `json:"supplier"`
	LeadTimeDays  int     `json:"lead_time_days"` // days to receive part
}