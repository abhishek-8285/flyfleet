package repository

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

// GeofenceRepository handles geofence CRUD and spatial queries
// using PostGIS.
type GeofenceRepository struct {
	pool *pgxpool.Pool
}

func NewGeofenceRepository(pool *pgxpool.Pool) *GeofenceRepository {
	return &GeofenceRepository{pool: pool}
}

// GetOrgGeofences returns all active geofences for an organization.
func (r *GeofenceRepository) GetOrgGeofences(ctx context.Context, orgID uuid.UUID) ([]models.Geofence, error) {
	q := `SELECT id, org_id, name, COALESCE(description,''), type,
	         center_lat, center_lng, radius_m,
	         COALESCE(ST_AsText(polygon), ''), status, created_at, updated_at
	      FROM geofences
	      WHERE org_id = $1 AND status = 'active'`

	rows, err := r.pool.Query(ctx, q, orgID)
	if err != nil {
		return nil, fmt.Errorf("query geofences: %w", err)
	}
	defer rows.Close()

	var geofences []models.Geofence
	for rows.Next() {
		var g models.Geofence
		var updatedAt interface{}
		var centerLat, centerLng, radiusM interface{}
		if err := rows.Scan(
			&g.ID, &g.OrgID, &g.Name, &g.Description, &g.Type,
			&centerLat, &centerLng, &radiusM,
			&g.PolygonWKT, &g.Status, &g.CreatedAt, &updatedAt,
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
		geofences = append(geofences, g)
	}
	return geofences, nil
}

// CreateGeofence inserts a new geofence.
func (r *GeofenceRepository) CreateGeofence(ctx context.Context, g *models.Geofence) error {
	if g.Type == models.GeofenceCircle {
		q := `INSERT INTO geofences (id, org_id, name, description, type, center_lat, center_lng, radius_m, status, created_at, updated_at)
		      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'active',NOW(),NOW())`
		_, err := r.pool.Exec(ctx, q,
			g.ID, g.OrgID, g.Name, g.Description, g.Type,
			*g.CenterLat, *g.CenterLng, *g.RadiusM,
		)
		return err
	}

	q := `INSERT INTO geofences (id, org_id, name, description, type, polygon, status, created_at, updated_at)
	      VALUES ($1,$2,$3,$4,$5,ST_GeomFromText($6, 4326),'active',NOW(),NOW())`
	_, err := r.pool.Exec(ctx, q,
		g.ID, g.OrgID, g.Name, g.Description, g.Type, g.PolygonWKT,
	)
	return err
}

// CheckPointInGeofences returns the list of geofences that contain
// the given point. Uses PostGIS spatial indexes.
func (r *GeofenceRepository) CheckPointInGeofences(ctx context.Context, orgID uuid.UUID, lat, lng float64) ([]models.Geofence, error) {
	q := `SELECT id, org_id, name, COALESCE(description,''), type,
	         center_lat, center_lng, radius_m,
	         COALESCE(ST_AsText(polygon), ''), status, created_at, updated_at
	      FROM geofences
	      WHERE org_id = $1 AND status = 'active'
	      AND (
	        (type = 'circle' AND ST_DWithin(
	          ST_SetSRID(ST_MakePoint($2, $3), 4326),
	          ST_SetSRID(ST_MakePoint(center_lng, center_lat), 4326),
	          radius_m
	        ))
	        OR
	        (type = 'polygon' AND ST_Contains(polygon, ST_SetSRID(ST_MakePoint($2, $3), 4326)))
	      )`

	rows, err := r.pool.Query(ctx, q, orgID, lng, lat)
	if err != nil {
		return nil, fmt.Errorf("spatial geofence check: %w", err)
	}
	defer rows.Close()

	var results []models.Geofence
	for rows.Next() {
		var g models.Geofence
		var updatedAt interface{}
		var centerLat, centerLng, radiusM interface{}
		if err := rows.Scan(
			&g.ID, &g.OrgID, &g.Name, &g.Description, &g.Type,
			&centerLat, &centerLng, &radiusM,
			&g.PolygonWKT, &g.Status, &g.CreatedAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan geofence result: %w", err)
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
		results = append(results, g)
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
