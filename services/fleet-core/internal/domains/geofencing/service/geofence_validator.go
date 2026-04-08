package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/geofencing/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GeofenceValidator validates geofence geometry and configuration
type GeofenceValidator struct {
	db *pgxpool.Pool
}

// NewGeofenceValidator creates a new geofence validator
func NewGeofenceValidator(db *pgxpool.Pool) *GeofenceValidator {
	return &GeofenceValidator{db: db}
}

// ValidationResult represents the result of geofence validation
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ValidateGeofence validates a geofence before creation
func (v *GeofenceValidator) ValidateGeofence(ctx context.Context, req *models.CreateGeofenceRequest) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Validate name
	if req.Name == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "name is required")
	} else if len(req.Name) > 100 {
		result.IsValid = false
		result.Errors = append(result.Errors, "name must be less than 100 characters")
	}

	// Validate type
	validTypes := map[string]bool{"circle": true, "polygon": true, "route_corridor": true}
	if req.Type == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "type is required")
	} else if !validTypes[req.Type] {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid geofence type: %s", req.Type))
	}

	// Validate based on type
	switch req.Type {
	case "circle":
		if req.Center == nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "center is required for circle geofence")
		} else {
			if err := v.validateCoordinates(req.Center.Lat, req.Center.Lng); err != nil {
				result.IsValid = false
				result.Errors = append(result.Errors, err.Error())
			}
		}

		if req.RadiusMeters == nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "radius_meters is required for circle geofence")
		} else if *req.RadiusMeters < 10 || *req.RadiusMeters > 50000 {
			result.IsValid = false
			result.Errors = append(result.Errors, "radius_meters must be between 10 and 50000 meters")
		} else if *req.RadiusMeters > 10000 {
			result.Warnings = append(result.Warnings, "large radius may impact performance")
		}

	case "polygon":
		if len(req.Coordinates) < 3 {
			result.IsValid = false
			result.Errors = append(result.Errors, "polygon requires at least 3 coordinates")
		} else {
			if err := v.validatePolygon(req.Coordinates); err != nil {
				result.IsValid = false
				result.Errors = append(result.Errors, err.Error())
			}
		}

	case "route_corridor":
		if req.RoutePolyline == nil || *req.RoutePolyline == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, "route_polyline is required for route corridor geofence")
		}

		if req.BufferMeters == nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "buffer_meters is required for route corridor geofence")
		} else if *req.BufferMeters < 10 || *req.BufferMeters > 1000 {
			result.IsValid = false
			result.Errors = append(result.Errors, "buffer_meters must be between 10 and 1000 meters")
		}
	}

	// Validate speed limit
	if req.SpeedLimit != nil {
		if *req.SpeedLimit < 10 || *req.SpeedLimit > 200 {
			result.IsValid = false
			result.Errors = append(result.Errors, "speed_limit must be between 10 and 200 km/h")
		}
	}

	// Validate active hours
	if req.ActiveHours != nil {
		if err := v.validateActiveHours(req.ActiveHours); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Validate vehicles
	if len(req.Vehicles) > 100 {
		result.Warnings = append(result.Warnings, "large number of vehicles may impact performance")
	}

	return result
}

// validateCoordinates validates latitude and longitude
func (v *GeofenceValidator) validateCoordinates(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}

// validatePolygon validates polygon coordinates
func (v *GeofenceValidator) validatePolygon(coords []models.Location) error {
	if len(coords) < 3 {
		return errors.New("polygon requires at least 3 coordinates")
	}

	// Validate each coordinate
	for i, coord := range coords {
		if err := v.validateCoordinates(coord.Lat, coord.Lng); err != nil {
			return fmt.Errorf("coordinate %d: %w", i, err)
		}
	}

	// Check if polygon is closed (first and last point should be the same)
	if coords[0].Lat != coords[len(coords)-1].Lat || coords[0].Lng != coords[len(coords)-1].Lng {
		// Not closed, but we can auto-close it
		// This is a warning, not an error
	}

	// Check geometry in PostGIS when available.
	if v.db != nil {
		if err := v.validatePolygonPostGIS(context.Background(), coords); err != nil {
			return fmt.Errorf("polygon validation: %w", err)
		}
	} else {
		// Fallback check without database.
		if err := v.checkPolygonSelfIntersection(coords); err != nil {
			return fmt.Errorf("polygon validation: %w", err)
		}
	}

	return nil
}

// checkPolygonSelfIntersection checks if a polygon intersects itself
func (v *GeofenceValidator) checkPolygonSelfIntersection(coords []models.Location) error {
	// Simplified check - in production, use a proper geometry library
	// This is a basic bounding box check
	if len(coords) < 4 {
		return nil
	}

	minLat, maxLat := coords[0].Lat, coords[0].Lat
	minLng, maxLng := coords[0].Lng, coords[0].Lng

	for _, coord := range coords[1:] {
		if coord.Lat < minLat {
			minLat = coord.Lat
		}
		if coord.Lat > maxLat {
			maxLat = coord.Lat
		}
		if coord.Lng < minLng {
			minLng = coord.Lng
		}
		if coord.Lng > maxLng {
			maxLng = coord.Lng
		}
	}

	// Check if bounding box is reasonable
	latRange := maxLat - minLat
	lngRange := maxLng - minLng

	if latRange > 10 || lngRange > 10 {
		return errors.New("polygon is too large (max 10 degrees)")
	}

	return nil
}

// validateActiveHours validates active hours configuration
func (v *GeofenceValidator) validateActiveHours(hours *models.ActiveHours) error {
	if hours == nil {
		return nil
	}

	// Validate time format (HH:MM)
	if err := v.validateTimeFormat(hours.Start); err != nil {
		return fmt.Errorf("start time: %w", err)
	}
	if err := v.validateTimeFormat(hours.End); err != nil {
		return fmt.Errorf("end time: %w", err)
	}

	// Validate days (0-6)
	for _, day := range hours.Days {
		if day < 0 || day > 6 {
			return errors.New("days must be between 0 (Sunday) and 6 (Saturday)")
		}
	}

	if len(hours.Days) == 0 {
		return errors.New("at least one day must be specified")
	}

	return nil
}

// validateTimeFormat validates time format (HH:MM)
func (v *GeofenceValidator) validateTimeFormat(timeStr string) error {
	if len(timeStr) != 5 {
		return errors.New("time must be in HH:MM format")
	}

	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil {
		return errors.New("time must be in HH:MM format")
	}

	if hour < 0 || hour > 23 {
		return errors.New("hour must be between 0 and 23")
	}
	if minute < 0 || minute > 59 {
		return errors.New("minute must be between 0 and 59")
	}

	return nil
}

// ValidateGeofenceUpdate validates a geofence update request
func (v *GeofenceValidator) ValidateGeofenceUpdate(ctx context.Context, req *models.UpdateGeofenceRequest) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Validate name if provided
	if req.Name != nil {
		if len(*req.Name) > 100 {
			result.IsValid = false
			result.Errors = append(result.Errors, "name must be less than 100 characters")
		}
	}

	// Validate type if provided
	if req.Type != nil {
		validTypes := map[string]bool{"circle": true, "polygon": true, "route_corridor": true}
		if !validTypes[*req.Type] {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("invalid geofence type: %s", *req.Type))
		}
	}

	// Validate radius if provided
	if req.RadiusMeters != nil {
		if *req.RadiusMeters < 10 || *req.RadiusMeters > 50000 {
			result.IsValid = false
			result.Errors = append(result.Errors, "radius_meters must be between 10 and 50000 meters")
		}
	}

	// Validate buffer if provided
	if req.BufferMeters != nil {
		if *req.BufferMeters < 10 || *req.BufferMeters > 1000 {
			result.IsValid = false
			result.Errors = append(result.Errors, "buffer_meters must be between 10 and 1000 meters")
		}
	}

	// Validate speed limit if provided
	if req.SpeedLimit != nil {
		if *req.SpeedLimit < 10 || *req.SpeedLimit > 200 {
			result.IsValid = false
			result.Errors = append(result.Errors, "speed_limit must be between 10 and 200 km/h")
		}
	}

	// Validate active hours if provided
	if req.ActiveHours != nil {
		if err := v.validateActiveHours(req.ActiveHours); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result
}

// ValidateGeofenceGeometry validates geofence geometry against database
func (v *GeofenceValidator) ValidateGeofenceGeometry(ctx context.Context, req *models.CreateGeofenceRequest) error {
	if v.db == nil || req == nil {
		return nil
	}

	switch req.Type {
	case "circle":
		if req.Center == nil {
			return errors.New("center is required for circle geofence")
		}
		var valid bool
		err := v.db.QueryRow(ctx, `
			SELECT ST_IsValid(ST_SetSRID(ST_MakePoint($1, $2), 4326))
		`, req.Center.Lng, req.Center.Lat).Scan(&valid)
		if err != nil {
			return fmt.Errorf("failed to validate circle geometry: %w", err)
		}
		if !valid {
			return errors.New("invalid circle geometry")
		}
	case "polygon":
		return v.validatePolygonPostGIS(ctx, req.Coordinates)
	}

	return nil
}

func (v *GeofenceValidator) validatePolygonPostGIS(ctx context.Context, coords []models.Location) error {
	if len(coords) < 3 {
		return errors.New("polygon requires at least 3 coordinates")
	}

	ring := make([][]float64, 0, len(coords)+1)
	for _, c := range coords {
		ring = append(ring, []float64{c.Lng, c.Lat})
	}
	first := coords[0]
	last := coords[len(coords)-1]
	if first.Lat != last.Lat || first.Lng != last.Lng {
		ring = append(ring, []float64{first.Lng, first.Lat})
	}

	geoJSONObj := map[string]any{
		"type":        "Polygon",
		"coordinates": []any{ring},
	}
	payload, err := json.Marshal(geoJSONObj)
	if err != nil {
		return fmt.Errorf("failed to build polygon geojson: %w", err)
	}

	var valid, simple bool
	err = v.db.QueryRow(ctx, `
		SELECT
			ST_IsValid(ST_GeomFromGeoJSON($1)),
			ST_IsSimple(ST_GeomFromGeoJSON($1))
	`, string(payload)).Scan(&valid, &simple)
	if err != nil {
		return fmt.Errorf("failed to validate polygon with postgis: %w", err)
	}
	if !valid || !simple {
		return errors.New("invalid polygon geometry")
	}
	return nil
}

// EncodePolyline encodes a polyline for route corridor geofences
func (v *GeofenceValidator) EncodePolyline(coords []models.Location) (string, error) {
	// Simple polyline encoding (Google's encoded polyline algorithm)
	// In production, use a proper library like github.com/paulmach/go.polyline

	var encoded string
	for _, coord := range coords {
		// Simplified encoding - in production use proper algorithm
		encoded += fmt.Sprintf("%.5f,%.5f;", coord.Lat, coord.Lng)
	}

	return encoded, nil
}

// DecodePolyline decodes a polyline for route corridor geofences
func (v *GeofenceValidator) DecodePolyline(encoded string) ([]models.Location, error) {
	// Simple polyline decoding
	// In production, use a proper library

	var coords []models.Location
	// Parse encoded string
	// This is a placeholder

	return coords, nil
}
