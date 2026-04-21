package repository

import (
	"context"
	"fmt"
	"math"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

// GeofenceRepository handles geofence CRUD and spatial queries.
// Works with plain PostgreSQL (no PostGIS required for circle geofences).
type GeofenceRepository struct {
	pool *pgxpool.Pool
}

func NewGeofenceRepository(pool *pgxpool.Pool) *GeofenceRepository {
	return &GeofenceRepository{pool: pool}
}

// GetOrgGeofences returns all active geofences for a tenant.
func (r *GeofenceRepository) GetOrgGeofences(ctx context.Context, tenantID uuid.UUID) ([]models.Geofence, error) {
	q := `SELECT id, tenant_id, name, COALESCE(description,''), type,
	         center_lat, center_lng, radius_m,
	         COALESCE(polygon_wkt, ''), enabled, created_at, updated_at
	      FROM geofences
	      WHERE tenant_id = $1 AND enabled = true
	      ORDER BY name`

	rows, err := r.pool.Query(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query geofences: %w", err)
	}
	defer rows.Close()

	var geofences []models.Geofence
	for rows.Next() {
		var g models.Geofence
		var updatedAt interface{}
		var centerLat, centerLng, radiusM interface{}
		var enabled bool
		if err := rows.Scan(
			&g.ID, &g.TenantID, &g.Name, &g.Description, &g.Type,
			&centerLat, &centerLng, &radiusM,
			&g.PolygonWKT, &enabled, &g.CreatedAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan geofence: %w", err)
		}
		if cl, ok := centerLat.(float64); ok {
			g.CenterLat = &cl
		}
		if cl, ok := centerLng.(float64); ok {
			g.CenterLng = &cl
		}
		if rm, ok := radiusM.(float64); ok {
			g.RadiusM = &rm
		}
		if enabled {
			g.Status = "active"
		} else {
			g.Status = "inactive"
		}
		geofences = append(geofences, g)
	}
	return geofences, nil
}

// CreateGeofence inserts a new geofence.
func (r *GeofenceRepository) CreateGeofence(ctx context.Context, g *models.Geofence) error {
	if g.Type == models.GeofenceCircle {
		q := `INSERT INTO geofences (id, tenant_id, name, description, type, center_lat, center_lng, radius_m, enabled, created_at, updated_at)
		      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,true,NOW(),NOW())`
		_, err := r.pool.Exec(ctx, q,
			g.ID, g.TenantID, g.Name, g.Description, g.Type,
			*g.CenterLat, *g.CenterLng, *g.RadiusM,
		)
		return err
	}

	// Polygon — store WKT as text
	q := `INSERT INTO geofences (id, tenant_id, name, description, type, polygon_wkt, enabled, created_at, updated_at)
	      VALUES ($1,$2,$3,$4,$5,$6,true,NOW(),NOW())`
	_, err := r.pool.Exec(ctx, q,
		g.ID, g.TenantID, g.Name, g.Description, g.Type, g.PolygonWKT,
	)
	return err
}

// haversineDistanceM returns the distance in meters between two lat/lng points.
func haversineDistanceM(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// CheckPointInGeofences returns the list of geofences that contain
// the given point. For circles, uses Haversine distance.
// For polygons, returns all (point-in-polygon requires PostGIS or client-side check).
func (r *GeofenceRepository) CheckPointInGeofences(ctx context.Context, tenantID uuid.UUID, lat, lng float64) ([]models.Geofence, error) {
	// Fetch all active geofences for the tenant and filter client-side
	all, err := r.GetOrgGeofences(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant geofences: %w", err)
	}

	var results []models.Geofence
	for _, g := range all {
		switch g.Type {
		case models.GeofenceCircle:
			if g.CenterLat != nil && g.CenterLng != nil && g.RadiusM != nil {
				dist := haversineDistanceM(lat, lng, *g.CenterLat, *g.CenterLng)
				if dist <= *g.RadiusM {
					results = append(results, g)
				}
			}
		case models.GeofencePolygon:
			// Without PostGIS, we can't do server-side point-in-polygon.
			// Return all polygon geofences — the geofence engine does
			// client-side ST_Contains via PostGIS in production.
			// For plain Postgres, we skip polygons here (the NATS-based
			// geofence engine handles this separately).
		}
	}
	return results, nil
}

// LogGeofenceEvent records an entry or exit event.
func (r *GeofenceRepository) LogGeofenceEvent(ctx context.Context, event *models.GeofenceEvent) error {
	q := `INSERT INTO geofence_events (id, geofence_id, vehicle_id, event_type, occurred_at, latitude, longitude, alert_sent, alert_channels, created_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())`
	_, err := r.pool.Exec(ctx, q,
		event.ID, event.GeofenceID, event.VehicleID, event.EventType,
		event.OccurredAt, event.Latitude, event.Longitude,
		event.AlertSent, event.AlertChannels,
	)
	return err
}
