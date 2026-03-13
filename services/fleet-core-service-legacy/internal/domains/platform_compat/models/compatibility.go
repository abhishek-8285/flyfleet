package models

import (
	"encoding/json"
	"time"
)

// CompatibilityStatus represents compatibility check status
type CompatibilityStatus string

const (
	CompatibilityStatusOK       CompatibilityStatus = "ok"
	CompatibilityStatusDegraded CompatibilityStatus = "degraded"
	CompatibilityStatusError    CompatibilityStatus = "error"
)

// CompatibilityResponse represents a compatibility check response
type CompatibilityResponse struct {
	Status  CompatibilityStatus `json:"status"`
	Mode    string              `json:"mode"`
	Message string              `json:"message,omitempty"`
	Service string              `json:"service,omitempty"`
}

// DashboardStatsResponse represents dashboard statistics response
type DashboardStatsResponse struct {
	Status      string                 `json:"status"`
	Mode        string                 `json:"mode"`
	Data        DashboardStatsData     `json:"data"`
	LastUpdated time.Time              `json:"last_updated"`
}

// DashboardStatsData represents dashboard statistics data
type DashboardStatsData struct {
	Fleet       map[string]interface{} `json:"fleet"`
	Alerts      map[string]interface{} `json:"alerts"`
	Maintenance map[string]interface{} `json:"maintenance"`
	Utilization map[string]interface{} `json:"utilization"`
}

// AdminDashboardResponse represents admin dashboard data response
type AdminDashboardResponse struct {
	Status       string                  `json:"status"`
	Mode         string                  `json:"mode"`
	Data         AdminDashboardData      `json:"data"`
	LastUpdated  time.Time               `json:"last_updated"`
}

// AdminDashboardData represents admin dashboard data
type AdminDashboardData struct {
	KPIs         map[string]interface{} `json:"kpis"`
	RecentAlerts []interface{}          `json:"recent_alerts"`
	Widgets      []interface{}          `json:"widgets"`
}

// ComplianceHOSRequest represents HOS compliance request
type ComplianceHOSRequest struct {
	DriverID    string    `json:"driver_id"`
	DutyStatus  string    `json:"duty_status"`
	Timestamp   time.Time `json:"timestamp"`
	Location    Location  `json:"location"`
	VehicleID   string    `json:"vehicle_id"`
	Odometer    int       `json:"odometer"`
}

// ComplianceHOSResponse represents HOS compliance response
type ComplianceHOSResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Target  string `json:"target"`
}

// TelemetryRequest represents telemetry data request
type TelemetryRequest struct {
	VehicleID   string                 `json:"vehicle_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Location    Location               `json:"location"`
	Speed       float64                `json:"speed"`
	Heading     float64                `json:"heading"`
	Odometer    int                    `json:"odometer"`
	EngineHours float64                `json:"engine_hours"`
	Sensors     map[string]interface{} `json:"sensors"`
}

// TelemetryResponse represents telemetry ingestion response
type TelemetryResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DashcamTelemetryRequest represents dashcam telemetry request
type DashcamTelemetryRequest struct {
	VehicleID string    `json:"vehicle_id"`
	Timestamp time.Time `json:"timestamp"`
	ImageURL  string    `json:"image_url"`
	EventType string    `json:"event_type"`
}

// MLPredictionRequest represents ML prediction request
type MLPredictionRequest struct {
	RouteID string `json:"route_id"`
	Date    string `json:"date"`
}

// MLPredictionResponse represents ML prediction response
type MLPredictionResponse struct {
	RouteID         string  `json:"route_id"`
	Date            string  `json:"date"`
	PredictedDemand int     `json:"predicted_demand"`
	ConfidenceScore float64 `json:"confidence_score"`
}

// MLPricingRequest represents ML pricing suggestion request
type MLPricingRequest struct {
	RouteID      string `json:"route_id"`
	Date         string `json:"date"`
	BaseCapacity int    `json:"base_capacity"`
}

// MLPricingResponse represents ML pricing suggestion response
type MLPricingResponse struct {
	RouteID             string  `json:"route_id"`
	Date                string  `json:"date"`
	SuggestedMultiplier string  `json:"suggested_multiplier"`
}

// MLTrainingRequest represents ML training job request
type MLTrainingRequest struct {
	ModelType   string                 `json:"model_type"`
	Parameters  map[string]interface{} `json:"parameters"`
	TrainingData TrainingData          `json:"training_data"`
}

// TrainingData represents training data for ML
type TrainingData struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Features  []string `json:"features"`
}

// MLTrainingResponse represents ML training job response
type MLTrainingResponse struct {
	Status string `json:"status"`
	Job    string `json:"job"`
	JobID  string `json:"job_id,omitempty"`
}

// RouteOptimizationRequest represents route optimization request
type RouteOptimizationRequest struct {
	Locations []Location `json:"locations"`
	VehicleID string     `json:"vehicle_id,omitempty"`
	StartTime string     `json:"start_time,omitempty"`
}

// RouteOptimizationResponse represents route optimization response
type RouteOptimizationResponse struct {
	Optimized bool     `json:"optimized"`
	Stops     int      `json:"stops"`
	Route     []string `json:"route,omitempty"`
	Distance  float64  `json:"distance,omitempty"`
	Duration  float64  `json:"duration,omitempty"`
}

// GetOptimizedRouteResponse represents get optimized route response
type GetOptimizedRouteResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Route  []string `json:"route,omitempty"`
}

// ServiceHealthRequest represents service health check request
type ServiceHealthRequest struct {
	ServiceName string `json:"service_name"`
}

// ServiceHealthResponse represents service health check response
type ServiceHealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Mode    string `json:"mode"`
}

// Location represents a geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

func (l *Location) UnmarshalJSON(data []byte) error {
	type Alias Location
	aux := struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if l.Latitude == 0 && aux.Lat != 0 {
		l.Latitude = aux.Lat
	}
	if l.Longitude == 0 && aux.Lng != 0 {
		l.Longitude = aux.Lng
	}
	return nil
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
	Details string `json:"details,omitempty"`
}

// RateLimitResponse represents rate limit error response
type RateLimitResponse struct {
	Error      string `json:"error"`
	RetryAfter int    `json:"retry_after"`
}
