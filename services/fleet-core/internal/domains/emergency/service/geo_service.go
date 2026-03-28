package service

import (
	"context"
)

type EmergencyGeoService struct{}

func NewEmergencyGeoService() *EmergencyGeoService {
	return &EmergencyGeoService{}
}

type NearbyResource struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Distance float64 `json:"distance_km"`
	Phone    string  `json:"phone"`
}

func (s *EmergencyGeoService) EnrichWithNearbyResources(ctx context.Context, lat, lon float64) map[string]interface{} {
	// In a real production system, this would query a spatial DB (PostGIS) or external Maps API
	// For this module validation, we simulate finding the nearest resources.

	resources := map[string]interface{}{
		"nearest_police_station": s.getNearest(lat, lon, "POLICE"),
		"nearest_hospital":       s.getNearest(lat, lon, "HOSPITAL"),
		"nearest_fire_station":   s.getNearest(lat, lon, "FIRE_STATION"),
		"geo_context_enriched":   true,
	}

	return resources
}

func (s *EmergencyGeoService) getNearest(lat, lon float64, resourceType string) NearbyResource {
	// Simulated lookup
	switch resourceType {
	case "POLICE":
		return NearbyResource{Name: "City South Police Station", Type: "POLICE", Distance: 1.2, Phone: "+91100"}
	case "HOSPITAL":
		return NearbyResource{Name: "General Memorial Hospital", Type: "HOSPITAL", Distance: 2.5, Phone: "+91102"}
	case "FIRE_STATION":
		return NearbyResource{Name: "Central Fire Station", Type: "FIRE_STATION", Distance: 3.1, Phone: "+91101"}
	default:
		return NearbyResource{Name: "Unknown", Type: resourceType, Distance: 0}
	}
}
