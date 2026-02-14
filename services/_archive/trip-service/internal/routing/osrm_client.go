package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/flyfleet/trip-service/internal/models"
)

// OSRMClient handles communication with OSRM routing service
type OSRMClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewOSRMClient creates a new OSRM client
func NewOSRMClient(baseURL string, logger *zap.Logger) *OSRMClient {
	return &OSRMClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// RouteRequest represents a routing request to OSRM
type RouteRequest struct {
	Coordinates    [][]float64 `json:"coordinates"`
	Alternatives   bool        `json:"alternatives"`
	Steps          bool        `json:"steps"`
	Annotations    bool        `json:"annotations"`
	Geometries     string      `json:"geometries"`
	Overview       string      `json:"overview"`
	ContinueStraight bool      `json:"continue_straight"`
}

// OSRMResponse represents the response from OSRM
type OSRMResponse struct {
	Code      string     `json:"code"`
	Routes    []OSRMRoute `json:"routes"`
	Waypoints []Waypoint `json:"waypoints"`
}

// OSRMRoute represents a route returned by OSRM
type OSRMRoute struct {
	Distance   float64     `json:"distance"`   // meters
	Duration   float64     `json:"duration"`   // seconds
	Geometry   string      `json:"geometry"`   // encoded polyline
	Legs       []RouteLeg  `json:"legs"`
	WeightName string      `json:"weight_name"`
	Weight     float64     `json:"weight"`
}

// RouteLeg represents a leg of a route
type RouteLeg struct {
	Distance    float64      `json:"distance"`
	Duration    float64      `json:"duration"`
	Steps       []RouteStep  `json:"steps"`
	Summary     string       `json:"summary"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

// RouteStep represents a step in a route
type RouteStep struct {
	Distance    float64     `json:"distance"`
	Duration    float64     `json:"duration"`
	Geometry    string      `json:"geometry"`
	Name        string      `json:"name"`
	Mode        string      `json:"mode"`
	Maneuver    Maneuver    `json:"maneuver"`
	Intersections []Intersection `json:"intersections"`
}

// Maneuver represents a navigation maneuver
type Maneuver struct {
	Location      []float64 `json:"location"`
	BearingBefore int       `json:"bearing_before"`
	BearingAfter  int       `json:"bearing_after"`
	Instruction   string    `json:"instruction"`
	Type          string    `json:"type"`
	Modifier      string    `json:"modifier,omitempty"`
}

// Intersection represents a road intersection
type Intersection struct {
	Location []float64 `json:"location"`
	Bearings []int     `json:"bearings"`
	Entry    []bool    `json:"entry"`
	In       int       `json:"in,omitempty"`
	Out      int       `json:"out,omitempty"`
}

// Annotations contains detailed information about the route
type Annotations struct {
	Distance []float64 `json:"distance"`
	Duration []float64 `json:"duration"`
	Speed    []float64 `json:"speed"`
}

// Waypoint represents a waypoint in the route
type Waypoint struct {
	Hint       string    `json:"hint"`
	Distance   float64   `json:"distance"`
	Name       string    `json:"name"`
	Location   []float64 `json:"location"`
}

// OptimizeRoute optimizes a route using OSRM
func (c *OSRMClient) OptimizeRoute(ctx context.Context, waypoints []models.Waypoint, options *models.OptimizeRouteRequest) (*models.RouteOptimizationResult, error) {
	// Convert waypoints to coordinate array
	coordinates := make([][]float64, len(waypoints))
	for i, wp := range waypoints {
		coordinates[i] = []float64{wp.Location.Longitude, wp.Location.Latitude}
	}

	// Build OSRM request
	request := RouteRequest{
		Coordinates:      coordinates,
		Alternatives:     true,
		Steps:           true,
		Annotations:     true,
		Geometries:      "polyline",
		Overview:        "full",
		ContinueStraight: false,
	}

	// Build query parameters
	params := url.Values{}
	params.Set("alternatives", fmt.Sprintf("%t", request.Alternatives))
	params.Set("steps", fmt.Sprintf("%t", request.Steps))
	params.Set("annotations", fmt.Sprintf("%t", request.Annotations))
	params.Set("geometries", request.Geometries)
	params.Set("overview", request.Overview)
	params.Set("continue_straight", fmt.Sprintf("%t", request.ContinueStraight))

	// Add coordinates to path
	coordStr := ""
	for i, coord := range coordinates {
		if i > 0 {
			coordStr += ";"
		}
		coordStr += fmt.Sprintf("%.6f,%.6f", coord[0], coord[1])
	}

	// Construct full URL
	fullURL := fmt.Sprintf("%s/route/v1/driving/%s?%s", c.baseURL, coordStr, params.Encode())

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM returned status %d", resp.StatusCode)
	}

	// Parse response
	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if osrmResp.Code != "Ok" {
		return nil, fmt.Errorf("OSRM returned error code: %s", osrmResp.Code)
	}

	// Convert OSRM response to our format
	result := &models.RouteOptimizationResult{
		Routes: make([]models.OptimizedRoute, len(osrmResp.Routes)),
	}

	for i, route := range osrmResp.Routes {
		optimizedRoute := models.OptimizedRoute{
			Polyline:        route.Geometry,
			DistanceKm:      route.Distance / 1000.0, // Convert meters to kilometers
			DurationMinutes: int(route.Duration / 60.0), // Convert seconds to minutes
			// In a real implementation, you would calculate tolls based on the route
			Tolls: []models.TollInfo{
				{Name: "Route-based toll calculation", CostINR: 0.0}, // Placeholder
			},
			FuelCostEstimate: (route.Distance / 1000.0) * 8.0, // Assuming 8 INR per km
			WeatherAlerts:    []string{}, // Would integrate with weather API
			Instructions:     c.extractInstructions(route.Legs),
		}
		
		result.Routes[i] = optimizedRoute
	}

	return result, nil
}

// extractInstructions extracts turn-by-turn instructions from route legs
func (c *OSRMClient) extractInstructions(legs []RouteLeg) []string {
	var instructions []string
	
	for _, leg := range legs {
		for _, step := range leg.Steps {
			if step.Maneuver.Instruction != "" {
				instructions = append(instructions, step.Maneuver.Instruction)
			}
		}
	}
	
	return instructions
}

// MockRoutingClient provides mock routing for testing
type MockRoutingClient struct {
	logger *zap.Logger
}

// NewMockRoutingClient creates a new mock routing client
func NewMockRoutingClient(logger *zap.Logger) *MockRoutingClient {
	return &MockRoutingClient{logger: logger}
}

// OptimizeRoute provides mock route optimization
func (m *MockRoutingClient) OptimizeRoute(ctx context.Context, waypoints []models.Waypoint, options *models.OptimizeRouteRequest) (*models.RouteOptimizationResult, error) {
	m.logger.Info("using mock routing client")
	
	// Calculate total distance (simple straight-line approximation)
	totalDistance := 0.0
	for i := 0; i < len(waypoints)-1; i++ {
		distance := calculateDistance(
			waypoints[i].Location.Latitude, waypoints[i].Location.Longitude,
			waypoints[i+1].Location.Latitude, waypoints[i+1].Location.Longitude,
		)
		totalDistance += distance
	}

	result := &models.RouteOptimizationResult{
		Routes: []models.OptimizedRoute{
			{
				Polyline:        "mock_encoded_polyline",
				DistanceKm:      totalDistance,
				DurationMinutes: int(totalDistance / 60.0 * 60), // Assuming 60 km/h average speed
				Tolls: []models.TollInfo{
					{Name: "Expressway Toll", CostINR: 450.0},
				},
				FuelCostEstimate: totalDistance * 8.0, // 8 INR per km
				WeatherAlerts:    []string{},
				Instructions: []string{
					"Start from first waypoint",
					"Follow the planned route",
					"Arrive at destination",
				},
			},
		},
	}

	return result, nil
}

// calculateDistance calculates distance between two coordinates using haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers
	
	// Convert degrees to radians
	lat1Rad := lat1 * 3.14159265359 / 180
	lon1Rad := lon1 * 3.14159265359 / 180
	lat2Rad := lat2 * 3.14159265359 / 180
	lon2Rad := lon2 * 3.14159265359 / 180
	
	// Haversine formula
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
	    math.Cos(lat1Rad)*math.Cos(lat2Rad)*
	    math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadius * c
}
