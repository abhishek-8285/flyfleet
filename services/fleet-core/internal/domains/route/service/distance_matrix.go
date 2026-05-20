package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/route/models"
)

// DistanceMatrixProvider defines the interface for fetching distance and travel time
type DistanceMatrixProvider interface {
	GetDistanceAndTime(ctx context.Context, origin, destination models.Location, vehicle *models.VehicleProfile) (float64, int, error)
	GetRoutePolyline(ctx context.Context, points []models.Location) ([][]float64, error)
}

// OSRMDistanceProvider implements DistanceMatrixProvider using a self-hosted OSRM instance
type OSRMDistanceProvider struct {
	BaseURL  string
	Fallback *HaversineDistanceProvider
}

type osrmResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

func (p *OSRMDistanceProvider) GetDistanceAndTime(ctx context.Context, origin, destination models.Location, vehicle *models.VehicleProfile) (float64, int, error) {
	if p.BaseURL == "" {
		return p.Fallback.GetDistanceAndTime(ctx, origin, destination, vehicle)
	}

	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=false",
		p.BaseURL, origin.Longitude, origin.Latitude, destination.Longitude, destination.Latitude)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return p.Fallback.GetDistanceAndTime(ctx, origin, destination, vehicle)
	}

	client := &http.Client{Timeout: 3000 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return p.Fallback.GetDistanceAndTime(ctx, origin, destination, vehicle)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.Fallback.GetDistanceAndTime(ctx, origin, destination, vehicle)
	}

	var data osrmResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Code != "Ok" || len(data.Routes) == 0 {
		return p.Fallback.GetDistanceAndTime(ctx, origin, destination, vehicle)
	}

	// OSRM returns distance in meters and duration in seconds
	distKM := data.Routes[0].Distance / 1000.0
	durationMin := int(data.Routes[0].Duration / 60.0)

	return distKM, durationMin, nil
}

func (p *OSRMDistanceProvider) GetRoutePolyline(ctx context.Context, points []models.Location) ([][]float64, error) {
	if len(points) <= 1 {
		return [][]float64{}, nil
	}
	if p.BaseURL == "" {
		return p.fallbackPolyline(points), nil
	}

	// Build OSRM query
	// Format: /route/v1/driving/lon1,lat1;lon2,lat2;...?overview=full&geometries=geojson
	var coordsStr string
	for i, pt := range points {
		if i > 0 {
			coordsStr += ";"
		}
		coordsStr += fmt.Sprintf("%f,%f", pt.Longitude, pt.Latitude)
	}

	url := fmt.Sprintf("%s/route/v1/driving/%s?overview=full&geometries=geojson", p.BaseURL, coordsStr)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return p.fallbackPolyline(points), nil
	}

	client := &http.Client{Timeout: 5000 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return p.fallbackPolyline(points), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.fallbackPolyline(points), nil
	}

	type osrmGeomResponse struct {
		Code   string `json:"code"`
		Routes []struct {
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"` // [[lon, lat], ...]
			} `json:"geometry"`
		} `json:"routes"`
	}

	var data osrmGeomResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Code != "Ok" || len(data.Routes) == 0 {
		return p.fallbackPolyline(points), nil
	}

	// Convert [lon, lat] from GeoJSON to [lat, lon] for Leaflet
	geojsonCoords := data.Routes[0].Geometry.Coordinates
	result := make([][]float64, len(geojsonCoords))
	for i, coord := range geojsonCoords {
		if len(coord) >= 2 {
			result[i] = []float64{coord[1], coord[0]}
		}
	}
	return result, nil
}

func (p *OSRMDistanceProvider) fallbackPolyline(points []models.Location) [][]float64 {
	// Simple straight line fallback
	result := make([][]float64, len(points))
	for i, pt := range points {
		result[i] = []float64{pt.Latitude, pt.Longitude}
	}
	return result
}

// HaversineDistanceProvider implements DistanceMatrixProvider using the Haversine formula
type HaversineDistanceProvider struct{}

func (p *HaversineDistanceProvider) GetDistanceAndTime(ctx context.Context, origin, destination models.Location, vehicle *models.VehicleProfile) (float64, int, error) {
	dist := HaversineDistance(origin.Latitude, origin.Longitude, destination.Latitude, destination.Longitude)

	speed := 40.0
	if vehicle != nil && vehicle.Speed > 0 {
		speed = vehicle.Speed
	}

	// time = (distance / speed) * 60 minutes
	travelTime := int((dist / speed) * 60)

	return dist, travelTime, nil
}

func (p *HaversineDistanceProvider) GetRoutePolyline(ctx context.Context, points []models.Location) ([][]float64, error) {
	result := make([][]float64, len(points))
	for i, pt := range points {
		result[i] = []float64{pt.Latitude, pt.Longitude}
	}
	return result, nil
}

// MockTrafficProvider simulates a traffic-aware API
type MockTrafficProvider struct {
	Fallback *HaversineDistanceProvider
}

func (p *MockTrafficProvider) GetDistanceAndTime(ctx context.Context, origin, destination models.Location, vehicle *models.VehicleProfile) (float64, int, error) {
	dist, travelTime, err := p.Fallback.GetDistanceAndTime(ctx, origin, destination, vehicle)
	if err != nil {
		return 0, 0, err
	}

	// Simulate traffic (10% to 50% increase in time)
	// In a real implementation, this would call Google or MapmyIndia
	trafficFactor := 1.2
	return dist, int(float64(travelTime) * trafficFactor), nil
}

func (p *MockTrafficProvider) GetRoutePolyline(ctx context.Context, points []models.Location) ([][]float64, error) {
	return p.Fallback.GetRoutePolyline(ctx, points)
}

// HaversineDistance calculates the distance between two points in KM
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in KM

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
