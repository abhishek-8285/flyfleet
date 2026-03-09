package models

import (
	"time"

	"github.com/google/uuid"
)

// RouteOptimizationRequest represents a request for route optimization
type RouteOptimizationRequest struct {
	ID                string                   `json:"id"`
	OrganizationID    uuid.UUID                `json:"organization_id"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description"`
	StartTime         time.Time                `json:"start_time"`
	EndTime           time.Time                `json:"end_time"`
	StartTimeWindow   *time.Time               `json:"start_time_window,omitempty"`
	EndTimeWindow     *time.Time               `json:"end_time_window,omitempty"`
	Constraints       RouteConstraints         `json:"constraints"`
	Stops             []Stop                   `json:"stops"`
	Vehicles          []VehicleProfile         `json:"vehicles"`
	Priorities        RoutePriorities          `json:"priorities"`
	OptimizationGoals OptimizationGoals        `json:"optimization_goals"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	Status            string                   `json:"status"`
	Result            *RouteOptimizationResult `json:"result,omitempty"`
}

// Stop represents a stop in the route
type Stop struct {
	ID              string     `json:"id"`
	Address         string     `json:"address"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	ServiceDuration int        `json:"service_duration"` // in minutes
	TimeWindowStart *time.Time `json:"time_window_start,omitempty"`
	TimeWindowEnd   *time.Time `json:"time_window_end,omitempty"`
	Priority        int        `json:"priority"` // 1 (high) to 5 (low)
	Demand          int        `json:"demand"`   // units needed at this stop
	Notes           string     `json:"notes"`
}

// VehicleProfile represents a vehicle with its capabilities
type VehicleProfile struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Capacity        int       `json:"capacity"`
	Speed           float64   `json:"speed"` // average speed in km/h
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	StartLocation   Location  `json:"start_location"`
	EndLocation     Location  `json:"end_location"`
	CostPerKM       float64   `json:"cost_per_km"`
	CostPerHour     float64   `json:"cost_per_hour"`
	Available       bool      `json:"available"`
	CurrentLocation *Location `json:"current_location,omitempty"`
}

// Location represents a geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

// RouteConstraints represents constraints for the route optimization
type RouteConstraints struct {
	MaxTravelTime      int      `json:"max_travel_time"` // in minutes
	MaxDistance        float64  `json:"max_distance"`    // in km
	MaxStopsPerRoute   int      `json:"max_stops_per_route"`
	MaxVehicles        int      `json:"max_vehicles"`
	MinStopsPerVehicle int      `json:"min_stops_per_vehicle"`
	IncludeTolls       *bool    `json:"include_tolls,omitempty"`
	ExcludeHighways    *bool    `json:"exclude_highways,omitempty"`
	PreferredRoads     []string `json:"preferred_roads,omitempty"`
	ForbiddenAreas     []string `json:"forbidden_areas,omitempty"`
}

// RoutePriorities represents priorities for the route optimization
type RoutePriorities struct {
	TimePriority          float64 `json:"time_priority"`          // 0.0 to 1.0
	CostPriority          float64 `json:"cost_priority"`          // 0.0 to 1.0
	ServicePriority       float64 `json:"service_priority"`       // 0.0 to 1.0
	EnvironmentalPriority float64 `json:"environmental_priority"` // 0.0 to 1.0
}

// OptimizationGoals represents the goals for optimization
type OptimizationGoals struct {
	MinimizeTravelTime          bool `json:"minimize_travel_time"`
	MinimizeDistance            bool `json:"minimize_distance"`
	MinimizeCost                bool `json:"minimize_cost"`
	MinimizeVehicles            bool `json:"minimize_vehicles"`
	MaximizeEfficiency          bool `json:"maximize_efficiency"`
	RespectTimeWindows          bool `json:"respect_time_windows"`
	MaximizeCapacityUtilization bool `json:"maximize_capacity_utilization"`
}

// RouteOptimizationResult represents the result of route optimization
type RouteOptimizationResult struct {
	ID                 string             `json:"id"`
	RequestID          string             `json:"request_id"`
	OptimizationTime   time.Time          `json:"optimization_time"`
	TotalDistance      float64            `json:"total_distance"` // in km
	TotalTime          int                `json:"total_time"`     // in minutes
	TotalCost          float64            `json:"total_cost"`
	TotalVehiclesUsed  int                `json:"total_vehicles_used"`
	TotalStops         int                `json:"total_stops"`
	Routes             []Route            `json:"routes"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
	CreatedAt          time.Time          `json:"created_at"`
}

// Route represents an optimized route for a single vehicle
type Route struct {
	VehicleID         string      `json:"vehicle_id"`
	Stops             []RouteStop `json:"stops"`
	Distance          float64     `json:"distance"` // in km
	Time              int         `json:"time"`     // in minutes
	Cost              float64     `json:"cost"`
	CapacityUsed      int         `json:"capacity_used"`
	StartTime         time.Time   `json:"start_time"`
	EndTime           time.Time   `json:"end_time"`
	OptimizationScore float64     `json:"optimization_score"` // 0.0 to 1.0
}

// RouteStop represents a stop in a route with timing and sequence information
type RouteStop struct {
	StopID        string    `json:"stop_id"`
	Sequence      int       `json:"sequence"`
	ArrivalTime   time.Time `json:"arrival_time"`
	DepartureTime time.Time `json:"departure_time"`
	ServiceTime   int       `json:"service_time"` // in minutes
	WaitingTime   int       `json:"waiting_time"` // in minutes
	Load          int       `json:"load"`         // load after service
}

// PerformanceMetrics contains metrics about the optimization performance
type PerformanceMetrics struct {
	ImprovementPercentage float64  `json:"improvement_percentage"`
	ExecutionTimeMs       int64    `json:"execution_time_ms"`
	ConvergenceIterations int      `json:"convergence_iterations"`
	ObjectiveValue        float64  `json:"objective_value"`
	InfeasibleConstraints []string `json:"infeasible_constraints,omitempty"`
}

// OptimizationJob represents a background optimization job
type OptimizationJob struct {
	ID          string     `json:"id"`
	RequestID   string     `json:"request_id"`
	Status      string     `json:"status"` // pending, running, completed, failed
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       *string    `json:"error,omitempty"`
	ResultID    *string    `json:"result_id,omitempty"`
}
