package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmergencyRepository interface {
	CreateIncident(ctx context.Context, incident *models.SOSIncident) error
	GetIncidentByID(ctx context.Context, id uuid.UUID) (*models.SOSIncident, error)
	ListActiveIncidents(ctx context.Context) ([]models.SOSIncident, error)
	UpdateIncidentStatus(ctx context.Context, id uuid.UUID, status models.SOSStatus) error
	LogIncidentAction(ctx context.Context, log *models.IncidentLog) error
	GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error)
}

type emergencyRepository struct {
	db *pgxpool.Pool
}

func NewEmergencyRepository(db *pgxpool.Pool) EmergencyRepository {
	return &emergencyRepository{db: db}
}

func (r *emergencyRepository) CreateIncident(ctx context.Context, incident *models.SOSIncident) error {
	query := `
		INSERT INTO sos_incidents (
			id, vehicle_id, driver_id, trip_id, timestamp, latitude, longitude, 
			location_address, status, severity, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13
		)
	`
	_, err := r.db.Exec(ctx, query,
		incident.ID, incident.VehicleID, incident.DriverID, incident.TripID, incident.Timestamp, incident.Latitude, incident.Longitude,
		incident.LocationAddress, incident.Status, incident.Severity, incident.Metadata, incident.CreatedAt, incident.UpdatedAt,
	)
	return err
}

func (r *emergencyRepository) GetIncidentByID(ctx context.Context, id uuid.UUID) (*models.SOSIncident, error) {
	query := `
		SELECT i.id, i.vehicle_id, i.driver_id, i.trip_id, i.timestamp, i.latitude, i.longitude, 
			   i.location_address, i.status, i.severity, i.metadata, i.created_at, i.updated_at,
			   v.registration_number as vehicle_reg_number, COALESCE(d.name, '') as driver_name
		FROM sos_incidents i
		JOIN vehicles v ON i.vehicle_id = v.id
		LEFT JOIN drivers d ON i.driver_id = d.id
		WHERE i.id = $1
	`
	var incident models.SOSIncident
	// Assuming pgx handling, scanning carefully. 
	// To simplify, mapping manually or using scany if available, but doing manual scan for now to be safe with standard pgx
	err := r.db.QueryRow(ctx, query, id).Scan(
		&incident.ID, &incident.VehicleID, &incident.DriverID, &incident.TripID, &incident.Timestamp, &incident.Latitude, &incident.Longitude,
		&incident.LocationAddress, &incident.Status, &incident.Severity, &incident.Metadata, &incident.CreatedAt, &incident.UpdatedAt,
		&incident.VehicleRegNumber, &incident.DriverName,
	)
	return &incident, err
}

func (r *emergencyRepository) ListActiveIncidents(ctx context.Context) ([]models.SOSIncident, error) {
	query := `
		SELECT i.id, i.vehicle_id, i.driver_id, i.trip_id, i.timestamp, i.latitude, i.longitude, 
			   i.location_address, i.status, i.severity, i.metadata, i.created_at, i.updated_at,
			   v.registration_number as vehicle_reg_number, COALESCE(d.name, '') as driver_name
		FROM sos_incidents i
		JOIN vehicles v ON i.vehicle_id = v.id
		LEFT JOIN drivers d ON i.driver_id = d.id
		WHERE i.status IN ('TRIGGERED', 'ACKNOWLEDGED')
		ORDER BY i.timestamp DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.SOSIncident
	for rows.Next() {
		var incident models.SOSIncident
		if err := rows.Scan(
			&incident.ID, &incident.VehicleID, &incident.DriverID, &incident.TripID, &incident.Timestamp, &incident.Latitude, &incident.Longitude,
			&incident.LocationAddress, &incident.Status, &incident.Severity, &incident.Metadata, &incident.CreatedAt, &incident.UpdatedAt,
			&incident.VehicleRegNumber, &incident.DriverName,
		); err != nil {
			return nil, err
		}
		incidents = append(incidents, incident)
	}
	return incidents, nil
}

func (r *emergencyRepository) UpdateIncidentStatus(ctx context.Context, id uuid.UUID, status models.SOSStatus) error {
	query := `UPDATE sos_incidents SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	return err
}

func (r *emergencyRepository) LogIncidentAction(ctx context.Context, log *models.IncidentLog) error {
	query := `
		INSERT INTO incident_logs (id, incident_id, user_id, action, notes, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.IncidentID, log.UserID, log.Action, log.Notes, log.Timestamp,
	)
	return err
}

func (r *emergencyRepository) GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	query := `SELECT id, organization_id, name, role, phone, email, priority, is_active, created_at, updated_at FROM emergency_contacts WHERE organization_id = $1 AND is_active = TRUE ORDER BY priority ASC`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.EmergencyContact
	for rows.Next() {
		var contact models.EmergencyContact
		if err := rows.Scan(
			&contact.ID, &contact.OrganizationID, &contact.Name, &contact.Role, &contact.Phone, &contact.Email, &contact.Priority, &contact.IsActive, &contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}
