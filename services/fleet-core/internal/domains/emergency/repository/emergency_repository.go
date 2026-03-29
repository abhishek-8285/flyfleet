package repository

import (
	"context"
	"fmt"
	"time"

	sharedmetrics "github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmergencyRepository interface {
	CreateIncident(ctx context.Context, incident *models.SOSIncident) error
	GetIncidentByID(ctx context.Context, id uuid.UUID) (*models.SOSIncident, error)
	GetActiveIncidentsByOrg(ctx context.Context, orgID uuid.UUID) ([]models.SOSIncident, error)
	ListActiveIncidents(ctx context.Context) ([]models.SOSIncident, error)
	UpdateIncidentStatus(ctx context.Context, id uuid.UUID, status models.SOSStatus, userID *uuid.UUID) error
	LogIncidentAction(ctx context.Context, log *models.IncidentLog) error
	GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error)
	GetRecentSOSByVehicle(ctx context.Context, vehicleID uuid.UUID, minutes int, incidentType models.IncidentType) (*models.SOSIncident, error)
	GetIncidentMetrics(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*IncidentMetrics, error)
	GetIncidentAuditHistory(ctx context.Context, incidentID uuid.UUID) (*models.IncidentAuditHistory, error)
	CreateEscalation(ctx context.Context, escalation *models.EmergencyEscalation) error
	UpdateEscalationStatus(ctx context.Context, id uuid.UUID, status string, respondedAt *time.Time) error
	CreateContact(ctx context.Context, contact *models.EmergencyContact) error
	ListContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error)
	CreateResponseUnit(ctx context.Context, unit *models.EmergencyResponseUnit) error
	ListResponseUnits(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyResponseUnit, error)
	UpdateVehicleHaltState(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, halted bool) error
	GetVehicleHaltState(ctx context.Context, vehicleID uuid.UUID) (bool, error)
	LogTimelineEvent(ctx context.Context, event *models.IncidentTimeline) error
	GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentTimeline, error)
	CreateMedia(ctx context.Context, media *models.IncidentMedia) error
	GetMediaByIncident(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentMedia, error)
	CreateVehicleSnapshot(ctx context.Context, snapshot *models.IncidentVehicleSnapshot) error
	GetVehicleSnapshot(ctx context.Context, incidentID uuid.UUID) (*models.IncidentVehicleSnapshot, error)
	GetIncidentsByDay(ctx context.Context, orgID uuid.UUID, days int) ([]models.DailyIncidentCount, error)
	GetIncidentTypeDistribution(ctx context.Context, orgID uuid.UUID) ([]models.IncidentTypeCount, error)
	GetVehicleSafetyScores(ctx context.Context, orgID uuid.UUID) ([]models.VehicleSafetyScore, error)
	ArchiveOldSnapshots(ctx context.Context, olderThan time.Time) (int64, error)
	ArchiveOldIncidents(ctx context.Context, olderThan time.Time) (int64, error)
	UpdateDriverFeedback(ctx context.Context, incidentID uuid.UUID, feedback string, notes string) error
	LogSLAViolation(ctx context.Context, violation *models.SLAViolation) error
	ArchiveOldMedia(ctx context.Context, olderThan time.Time) (int64, error)
	// Phase 1: Contact Acknowledgement Tracking
	RecordContactNotification(ctx context.Context, incidentID, contactID uuid.UUID) error
	UpdateContactResponse(ctx context.Context, incidentID, contactID uuid.UUID, status string, notes string) error
	// Phase 6: Analytics
	GetTopLocations(ctx context.Context, orgID uuid.UUID, limit int) ([]string, error)
	GetIncidentsCount(ctx context.Context, orgID uuid.UUID, since time.Time) (int, error)
	// Phase 4: Incident Recovery Worker
	GetStalledIncidents(ctx context.Context, stalledFor time.Duration) ([]models.SOSIncident, error)
	// Phase 3: Offline SOS Event Queue
	EnqueueDeviceEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload []byte) error
	GetPendingDeviceEvents(ctx context.Context, limit int) ([]models.DeviceEventQueueItem, error)
	MarkDeviceEventDelivered(ctx context.Context, id uuid.UUID) error
	IncrementDeviceEventRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error
	// Phase 2: Authority Delivery logs
	LogAuthorityDelivery(ctx context.Context, log *models.AuthorityDeliveryLog) error
	UpdateAuthorityDelivery(ctx context.Context, id uuid.UUID, status string, responseCode int, responseBody string) error
	// Phase 5: Incident Timeline Storage
	CreateTimelineEvent(ctx context.Context, event *models.TimelineEvent) error
	GetTimelineEvents(ctx context.Context, incidentID uuid.UUID) ([]models.TimelineEvent, error)
	// Phase 4: Geo Routing
	GetAuthoritiesByLocation(ctx context.Context, lat, lng float64) ([]models.EmergencyResponseUnit, error)
	// Phase 10: Production Lockdown
	Ping(ctx context.Context) error
	GetCriticalIncidentsForSnapshot(ctx context.Context, since time.Time) ([]models.SOSIncident, error)
	CreateIncidentSnapshot(ctx context.Context, snapshot *models.IncidentSnapshot) error
	ExportIncidents(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]models.SOSIncident, error)
}

type emergencyRepository struct {
	db *pgxpool.Pool
}

func (r *emergencyRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func NewEmergencyRepository(db *pgxpool.Pool) EmergencyRepository {
	return &emergencyRepository{db: db}
}

func (r *emergencyRepository) CreateIncident(ctx context.Context, incident *models.SOSIncident) error {
	query := `
		INSERT INTO sos_incidents (id, vehicle_id, driver_id, trip_id, device_id, org_id, timestamp, latitude, longitude, location_address, status, severity, incident_type, panic_source, speed, heading, metadata, high_risk_location, simulation_mode, driver_feedback, feedback_notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		incident.ID, incident.VehicleID, incident.DriverID, incident.TripID, incident.DeviceID, incident.OrgID,
		incident.Timestamp, incident.Latitude, incident.Longitude, incident.LocationAddress,
		incident.Status, incident.Severity, incident.IncidentType, incident.PanicSource,
		incident.Speed, incident.Heading, incident.Metadata, incident.HighRiskLocation, incident.SimulationMode,
		incident.DriverFeedback, incident.FeedbackNotes, incident.CreatedAt, incident.UpdatedAt,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to create SOS incident: %w", err)
	}
	return nil
}

func (r *emergencyRepository) GetIncidentByID(ctx context.Context, id uuid.UUID) (*models.SOSIncident, error) {
	query := `
		SELECT id, vehicle_id, driver_id, trip_id, device_id, org_id, timestamp, latitude, longitude,
			   location_address, status, severity, incident_type, panic_source, speed, heading,
			   metadata, high_risk_location, simulation_mode, driver_feedback, feedback_notes,
			   acknowledged_at, resolved_at, acknowledged_by, resolved_by, created_at, updated_at
		FROM sos_incidents
		WHERE id = $1
	`
	start := time.Now()
	var incident models.SOSIncident
	err := r.db.QueryRow(ctx, query, id).Scan(
		&incident.ID, &incident.VehicleID, &incident.DriverID, &incident.TripID, &incident.DeviceID, &incident.OrgID,
		&incident.Timestamp, &incident.Latitude, &incident.Longitude, &incident.LocationAddress,
		&incident.Status, &incident.Severity, &incident.IncidentType, &incident.PanicSource,
		&incident.Speed, &incident.Heading, &incident.Metadata, &incident.HighRiskLocation,
		&incident.SimulationMode, &incident.DriverFeedback, &incident.FeedbackNotes,
		&incident.AcknowledgedAt, &incident.ResolvedAt, &incident.AcknowledgedBy, &incident.ResolvedBy,
		&incident.CreatedAt, &incident.UpdatedAt,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get SOS incident by ID: %w", err)
	}
	return &incident, nil
}

func (r *emergencyRepository) GetActiveIncidentsByOrg(ctx context.Context, orgID uuid.UUID) ([]models.SOSIncident, error) {
	query := `
		SELECT i.id, i.vehicle_id, i.driver_id, i.trip_id, i.device_id, i.org_id, i.timestamp, i.latitude, i.longitude,
			   i.location_address, i.status, i.severity, i.incident_type, i.panic_source, i.speed, i.heading,
			   i.metadata, i.acknowledged_by, i.resolved_by, i.created_at, i.updated_at,
			   v.registration_number as vehicle_reg_number, d.full_name as driver_name
		FROM sos_incidents i
		LEFT JOIN vehicles v ON i.vehicle_id = v.id
		LEFT JOIN drivers d ON i.driver_id = d.id
		WHERE i.org_id = $1 AND i.status IN ('TRIGGERED', 'ACKNOWLEDGED')
		ORDER BY i.timestamp DESC
	`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query active incidents: %w", err)
	}
	defer rows.Close()

	var incidents []models.SOSIncident
	for rows.Next() {
		var incident models.SOSIncident
		err := rows.Scan(
			&incident.ID, &incident.VehicleID, &incident.DriverID, &incident.TripID, &incident.DeviceID, &incident.OrgID,
			&incident.Timestamp, &incident.Latitude, &incident.Longitude, &incident.LocationAddress,
			&incident.Status, &incident.Severity, &incident.IncidentType, &incident.PanicSource,
			&incident.Speed, &incident.Heading, &incident.Metadata, &incident.AcknowledgedBy,
			&incident.ResolvedBy, &incident.CreatedAt, &incident.UpdatedAt,
			&incident.VehicleRegNumber, &incident.DriverName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SOS incident: %w", err)
		}
		incidents = append(incidents, incident)
	}
	return incidents, nil
}

func (r *emergencyRepository) ListActiveIncidents(ctx context.Context) ([]models.SOSIncident, error) {
	query := `
		SELECT i.id, i.vehicle_id, i.driver_id, i.trip_id, i.device_id, i.org_id, i.timestamp, i.latitude, i.longitude,
			   i.location_address, i.status, i.severity, i.incident_type, i.panic_source, i.speed, i.heading,
			   i.metadata, i.acknowledged_at, i.resolved_at, i.acknowledged_by, i.resolved_by, i.created_at, i.updated_at,
			   v.registration_number as vehicle_reg_number, d.full_name as driver_name
		FROM sos_incidents i
		LEFT JOIN vehicles v ON i.vehicle_id = v.id
		LEFT JOIN drivers d ON i.driver_id = d.id
		WHERE i.status IN ('TRIGGERED', 'ACKNOWLEDGED')
		ORDER BY i.timestamp DESC
	`
	start := time.Now()
	rows, err := r.db.Query(ctx, query)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query all active incidents: %w", err)
	}
	defer rows.Close()

	var incidents []models.SOSIncident
	for rows.Next() {
		var incident models.SOSIncident
		err := rows.Scan(
			&incident.ID, &incident.VehicleID, &incident.DriverID, &incident.TripID, &incident.DeviceID, &incident.OrgID,
			&incident.Timestamp, &incident.Latitude, &incident.Longitude, &incident.LocationAddress,
			&incident.Status, &incident.Severity, &incident.IncidentType, &incident.PanicSource,
			&incident.Speed, &incident.Heading, &incident.Metadata, &incident.AcknowledgedAt,
			&incident.ResolvedAt, &incident.AcknowledgedBy, &incident.ResolvedBy, &incident.CreatedAt, &incident.UpdatedAt,
			&incident.VehicleRegNumber, &incident.DriverName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SOS incident: %w", err)
		}
		incidents = append(incidents, incident)
	}
	return incidents, nil
}

func (r *emergencyRepository) UpdateIncidentStatus(ctx context.Context, id uuid.UUID, status models.SOSStatus, userID *uuid.UUID) error {
	var query string
	var args []interface{}

	if status == models.StatusAcknowledged {
		query = `UPDATE sos_incidents SET status = $1, acknowledged_by = $2, acknowledged_at = NOW(), updated_at = NOW() WHERE id = $3`
		args = []interface{}{status, userID, id}
	} else if status == models.StatusResolved || status == models.StatusFalseAlarm {
		query = `UPDATE sos_incidents SET status = $1, resolved_by = $2, resolved_at = NOW(), updated_at = NOW() WHERE id = $3`
		args = []interface{}{status, userID, id}
	} else {
		query = `UPDATE sos_incidents SET status = $1, updated_at = NOW() WHERE id = $2`
		args = []interface{}{status, id}
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query, args...)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}
	return nil
}

func (r *emergencyRepository) LogIncidentAction(ctx context.Context, log *models.IncidentLog) error {
	query := `
		INSERT INTO incident_logs (id, incident_id, user_id, action, notes, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		log.ID, log.IncidentID, log.UserID, log.Action, log.Notes, log.Timestamp,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to log incident action: %w", err)
	}
	return nil
}

func (r *emergencyRepository) GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	query := `SELECT id, org_id, name, role, phone, email, priority, is_active, created_at, updated_at FROM emergency_contacts WHERE org_id = $1 AND is_active = TRUE ORDER BY priority ASC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query emergency contacts: %w", err)
	}
	defer rows.Close()

	var contacts []models.EmergencyContact
	for rows.Next() {
		var contact models.EmergencyContact
		if err := rows.Scan(
			&contact.ID, &contact.OrgID, &contact.Name, &contact.Role, &contact.Phone, &contact.Email, &contact.Priority, &contact.IsActive, &contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan emergency contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("emergency contact iteration error: %w", err)
	}

	return contacts, nil
}

func (r *emergencyRepository) GetRecentSOSByVehicle(ctx context.Context, vehicleID uuid.UUID, minutes int, incidentType models.IncidentType) (*models.SOSIncident, error) {
	query := `
		SELECT id, vehicle_id, driver_id, trip_id, device_id, org_id, timestamp, latitude, longitude,
			   location_address, status, severity, incident_type, panic_source, speed, heading,
			   metadata, acknowledged_by, resolved_by, created_at, updated_at
		FROM sos_incidents
		WHERE vehicle_id = $1
		  AND incident_type = $3
		  AND timestamp >= NOW() - ($2 || ' minutes')::INTERVAL
		ORDER BY timestamp DESC
		LIMIT 1
	`

	start := time.Now()
	var incident models.SOSIncident
	err := r.db.QueryRow(ctx, query, vehicleID, minutes, incidentType).Scan(
		&incident.ID, &incident.VehicleID, &incident.DriverID, &incident.TripID, &incident.DeviceID, &incident.OrgID,
		&incident.Timestamp, &incident.Latitude, &incident.Longitude, &incident.LocationAddress,
		&incident.Status, &incident.Severity, &incident.IncidentType, &incident.PanicSource,
		&incident.Speed, &incident.Heading, &incident.Metadata, &incident.AcknowledgedBy,
		&incident.ResolvedBy, &incident.CreatedAt, &incident.UpdatedAt,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recent SOS by vehicle: %w", err)
	}

	return &incident, nil
}

type IncidentMetrics struct {
	TotalIncidents       int     `json:"total_incidents"`
	AcknowledgedCount    int     `json:"acknowledged_count"`
	ResolvedCount        int     `json:"resolved_count"`
	FalseAlarmCount      int     `json:"false_alarm_count"`
	AvgResponseTimeMin   float64 `json:"avg_response_time_min"`
	AvgResolutionTimeMin float64 `json:"avg_resolution_time_min"`
	CriticalCount        int     `json:"critical_count"`
	MediumCount          int     `json:"medium_count"`
	LowCount             int     `json:"low_count"`
}

func (r *emergencyRepository) GetIncidentMetrics(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*IncidentMetrics, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status IN ('TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED', 'FALSE_ALARM')) as total,
			COUNT(*) FILTER (WHERE status = 'ACKNOWLEDGED') as acknowledged,
			COUNT(*) FILTER (WHERE status = 'RESOLVED') as resolved,
			COUNT(*) FILTER (WHERE status = 'FALSE_ALARM') as false_alarm,
			COUNT(*) FILTER (WHERE severity = 'CRITICAL') as critical,
			COUNT(*) FILTER (WHERE severity = 'MEDIUM') as medium,
			COUNT(*) FILTER (WHERE severity = 'LOW') as low,
			AVG(EXTRACT(EPOCH FROM (acknowledged_at - timestamp)) / 60) FILTER (WHERE status IN ('ACKNOWLEDGED', 'RESOLVED')) as avg_response,
			AVG(EXTRACT(EPOCH FROM (resolved_at - timestamp)) / 60) FILTER (WHERE status = 'RESOLVED') as avg_resolution
		FROM sos_incidents i
		WHERE i.org_id = $1
		  AND i.timestamp >= $2
		  AND i.timestamp <= $3
	`

	start := time.Now()
	var im IncidentMetrics
	var avgResponse, avgResolution *float64

	err := r.db.QueryRow(ctx, query, orgID, startDate, endDate).Scan(
		&im.TotalIncidents, &im.AcknowledgedCount, &im.ResolvedCount,
		&im.FalseAlarmCount, &im.CriticalCount, &im.MediumCount, &im.LowCount,
		&avgResponse, &avgResolution,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		return nil, fmt.Errorf("failed to get incident metrics: %w", err)
	}

	if avgResponse != nil {
		im.AvgResponseTimeMin = *avgResponse
	}
	if avgResolution != nil {
		im.AvgResolutionTimeMin = *avgResolution
	}

	return &im, nil
}

func (r *emergencyRepository) GetIncidentAuditHistory(ctx context.Context, incidentID uuid.UUID) (*models.IncidentAuditHistory, error) {
	incident, err := r.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, nil
	}

	query := `
		SELECT id, incident_id, user_id, action, notes, timestamp
		FROM incident_logs
		WHERE incident_id = $1
		ORDER BY timestamp ASC
	`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, incidentID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	history := &models.IncidentAuditHistory{
		IncidentID: incidentID,
		Logs:       []models.IncidentLog{},
		Timeline:   []models.IncidentTimeline{},
	}

	for rows.Next() {
		var log models.IncidentLog
		if err := rows.Scan(&log.ID, &log.IncidentID, &log.UserID, &log.Action, &log.Notes, &log.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		history.Logs = append(history.Logs, log)

		// Create timeline entry
		entry := models.IncidentTimeline{
			CreatedAt: log.Timestamp,
			EventType: log.Action,
			Actor:     "USER", // Default actor for logs
		}
		if log.UserID != nil {
			entry.Actor = log.UserID.String()[:8]
		}
		history.Timeline = append(history.Timeline, entry)
	}

	// Fetch timeline events
	timeline, err := r.GetTimeline(ctx, incidentID)
	if err == nil {
		history.Timeline = append(history.Timeline, timeline...)
	}

	return history, nil
}

func (r *emergencyRepository) CreateEscalation(ctx context.Context, escalation *models.EmergencyEscalation) error {
	query := `
		INSERT INTO emergency_escalations (id, incident_id, contact_id, notification_channel, status, attempted_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, escalation.ID, escalation.IncidentID, escalation.ContactID, escalation.NotificationChannel, escalation.Status, escalation.AttemptedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) UpdateEscalationStatus(ctx context.Context, id uuid.UUID, status string, respondedAt *time.Time) error {
	query := `UPDATE emergency_escalations SET status = $1, responded_at = $2 WHERE id = $3`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, status, respondedAt, id)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) CreateContact(ctx context.Context, contact *models.EmergencyContact) error {
	query := `
		INSERT INTO emergency_contacts (id, org_id, name, role, phone, email, priority, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, contact.ID, contact.OrgID, contact.Name, contact.Role, contact.Phone, contact.Email, contact.Priority, contact.IsActive, contact.CreatedAt, contact.UpdatedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) ListContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	query := `SELECT id, org_id, name, role, phone, email, priority, is_active, created_at, updated_at FROM emergency_contacts WHERE org_id = $1 ORDER BY priority ASC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.EmergencyContact
	for rows.Next() {
		var c models.EmergencyContact
		if err := rows.Scan(&c.ID, &c.OrgID, &c.Name, &c.Role, &c.Phone, &c.Email, &c.Priority, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}

func (r *emergencyRepository) CreateResponseUnit(ctx context.Context, unit *models.EmergencyResponseUnit) error {
	query := `
		INSERT INTO emergency_response_units (id, org_id, unit_type, name, phone, coverage_city, coverage_state, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, unit.ID, unit.OrgID, unit.UnitType, unit.Name, unit.Phone, unit.CoverageCity, unit.CoverageState, unit.IsActive)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) ListResponseUnits(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyResponseUnit, error) {
	query := `SELECT id, org_id, unit_type, name, phone, coverage_city, coverage_state, is_active FROM emergency_response_units WHERE org_id = $1 AND is_active = TRUE`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []models.EmergencyResponseUnit
	for rows.Next() {
		var u models.EmergencyResponseUnit
		if err := rows.Scan(&u.ID, &u.OrgID, &u.UnitType, &u.Name, &u.Phone, &u.CoverageCity, &u.CoverageState, &u.IsActive); err != nil {
			return nil, err
		}
		// Map back for AIS-140 compliance logic in service
		u.ContactNumber = u.Phone
		u.Type = u.UnitType
		units = append(units, u)
	}
	return units, nil
}

func (r *emergencyRepository) UpdateVehicleHaltState(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, halted bool) error {
	query := `
		INSERT INTO emergency_vehicle_state (vehicle_id, org_id, halted, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (vehicle_id) DO UPDATE SET halted = $3, updated_at = $4
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, vehicleID, orgID, halted, time.Now())
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) GetVehicleHaltState(ctx context.Context, vehicleID uuid.UUID) (bool, error) {
	var halted bool
	query := `SELECT halted FROM emergency_vehicle_state WHERE vehicle_id = $1`
	start := time.Now()
	err := r.db.QueryRow(ctx, query, vehicleID).Scan(&halted)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return halted, err
}

func (r *emergencyRepository) LogTimelineEvent(ctx context.Context, event *models.IncidentTimeline) error {
	query := `
		INSERT INTO incident_timeline (id, incident_id, event_type, actor, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, event.ID, event.IncidentID, event.EventType, event.Actor, event.Metadata, event.CreatedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentTimeline, error) {
	query := `
		SELECT id, incident_id, event_type, actor, metadata, created_at
		FROM incident_timeline
		WHERE incident_id = $1
		ORDER BY created_at ASC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, incidentID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.IncidentTimeline
	for rows.Next() {
		var e models.IncidentTimeline
		if err := rows.Scan(&e.ID, &e.IncidentID, &e.EventType, &e.Actor, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *emergencyRepository) CreateMedia(ctx context.Context, media *models.IncidentMedia) error {
	query := `
		INSERT INTO incident_media (id, incident_id, media_type, file_url, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, media.ID, media.IncidentID, media.MediaType, media.FileURL, media.UploadedBy, media.CreatedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) GetMediaByIncident(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentMedia, error) {
	query := `SELECT id, incident_id, media_type, file_url, uploaded_by, created_at FROM incident_media WHERE incident_id = $1`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, incidentID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var mediaList []models.IncidentMedia
	for rows.Next() {
		var m models.IncidentMedia
		if err := rows.Scan(&m.ID, &m.IncidentID, &m.MediaType, &m.FileURL, &m.UploadedBy, &m.CreatedAt); err != nil {
			return nil, err
		}
		mediaList = append(mediaList, m)
	}
	return mediaList, nil
}

func (r *emergencyRepository) CreateVehicleSnapshot(ctx context.Context, snapshot *models.IncidentVehicleSnapshot) error {
	query := `
		INSERT INTO incident_vehicle_snapshot (incident_id, speed, heading, engine_status, fuel_level, battery_voltage, gps_accuracy, device_signal_strength, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, snapshot.IncidentID, snapshot.Speed, snapshot.Heading, snapshot.EngineStatus, snapshot.FuelLevel, snapshot.BatteryVoltage, snapshot.GPSAccuracy, snapshot.DeviceSignalStrength, snapshot.CreatedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) GetVehicleSnapshot(ctx context.Context, incidentID uuid.UUID) (*models.IncidentVehicleSnapshot, error) {
	query := `SELECT incident_id, speed, heading, engine_status, fuel_level, battery_voltage, gps_accuracy, device_signal_strength, created_at FROM incident_vehicle_snapshot WHERE incident_id = $1`
	start := time.Now()
	var s models.IncidentVehicleSnapshot
	err := r.db.QueryRow(ctx, query, incidentID).Scan(&s.IncidentID, &s.Speed, &s.Heading, &s.EngineStatus, &s.FuelLevel, &s.BatteryVoltage, &s.GPSAccuracy, &s.DeviceSignalStrength, &s.CreatedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *emergencyRepository) GetIncidentsByDay(ctx context.Context, orgID uuid.UUID, days int) ([]models.DailyIncidentCount, error) {
	query := `
		SELECT TO_CHAR(timestamp, 'YYYY-MM-DD') as date, COUNT(*) as count
		FROM sos_incidents
		WHERE org_id = $1 AND timestamp >= NOW() - ($2 || ' days')::INTERVAL
		GROUP BY date
		ORDER BY date ASC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID, days)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.DailyIncidentCount
	for rows.Next() {
		var d models.DailyIncidentCount
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	return results, nil
}

func (r *emergencyRepository) GetIncidentTypeDistribution(ctx context.Context, orgID uuid.UUID) ([]models.IncidentTypeCount, error) {
	query := `
		SELECT incident_type as type, COUNT(*) as count
		FROM sos_incidents
		WHERE org_id = $1
		GROUP BY type`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.IncidentTypeCount
	for rows.Next() {
		var t models.IncidentTypeCount
		if err := rows.Scan(&t.Type, &t.Count); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (r *emergencyRepository) GetVehicleSafetyScores(ctx context.Context, orgID uuid.UUID) ([]models.VehicleSafetyScore, error) {
	query := `
		SELECT v.id, v.registration_number, COUNT(i.id) as incident_count,
			   GREATEST(0, (100 - COUNT(i.id) * 5)::NUMERIC) as safety_score
		FROM vehicles v
		LEFT JOIN sos_incidents i ON v.id = i.vehicle_id
		WHERE v.org_id = $1
		GROUP BY v.id, v.registration_number
		ORDER BY safety_score ASC
		LIMIT 10`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.VehicleSafetyScore
	for rows.Next() {
		var s models.VehicleSafetyScore
		if err := rows.Scan(&s.VehicleID, &s.VehicleRegNumber, &s.IncidentCount, &s.SafetyScore); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

func (r *emergencyRepository) ArchiveOldSnapshots(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM incident_vehicle_snapshot WHERE created_at < $1`
	start := time.Now()
	res, err := r.db.Exec(ctx, query, olderThan)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *emergencyRepository) GetCriticalIncidentsForSnapshot(ctx context.Context, since time.Time) ([]models.SOSIncident, error) {
	query := `SELECT * FROM sos_incidents WHERE severity = 'CRITICAL' AND created_at >= $1`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, since)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.SOSIncident
	for rows.Next() {
		var i models.SOSIncident
		err := rows.Scan(
			&i.ID, &i.VehicleID, &i.DriverID, &i.TripID, &i.Timestamp, &i.Latitude, &i.Longitude,
			&i.LocationAddress, &i.Status, &i.Severity, &i.Metadata, &i.CreatedAt, &i.UpdatedAt,
			&i.IncidentType, &i.PanicSource, &i.Speed, &i.Heading, &i.AcknowledgedBy, &i.ResolvedBy,
			&i.OrgID, &i.HighRiskLocation, &i.SimulationMode, &i.DriverFeedback, &i.FeedbackNotes,
			&i.AcknowledgedAt, &i.ResolvedAt, &i.DeviceID,
		)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, i)
	}
	return incidents, nil
}

func (r *emergencyRepository) CreateIncidentSnapshot(ctx context.Context, snapshot *models.IncidentSnapshot) error {
	query := `
		INSERT INTO emergency_incident_snapshots (incident_id, vehicle_id, org_id, incident_type, severity, status, latitude, longitude, snapshot_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		snapshot.IncidentID, snapshot.VehicleID, snapshot.OrgID,
		snapshot.IncidentType, snapshot.Severity, snapshot.Status,
		snapshot.Latitude, snapshot.Longitude, snapshot.SnapshotData)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) ExportIncidents(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]models.SOSIncident, error) {
	query := `
		SELECT
			id, vehicle_id, driver_id, trip_id, timestamp, latitude, longitude, location_address,
			status, severity, incident_type, panic_source, speed, heading, metadata,
			acknowledged_by, resolved_by, org_id, high_risk_location, simulation_mode,
			driver_feedback, feedback_notes, acknowledged_at, resolved_at, created_at, updated_at, device_id
		FROM sos_incidents
		WHERE org_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC
	`
	queryStart := time.Now()
	rows, err := r.db.Query(ctx, query, orgID, start, end)
	sharedmetrics.RecordDBQuery(query, time.Since(queryStart))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.SOSIncident
	for rows.Next() {
		var i models.SOSIncident
		err := rows.Scan(
			&i.ID, &i.VehicleID, &i.DriverID, &i.TripID, &i.Timestamp, &i.Latitude, &i.Longitude,
			&i.LocationAddress, &i.Status, &i.Severity, &i.IncidentType, &i.PanicSource, &i.Speed, &i.Heading, &i.Metadata,
			&i.AcknowledgedBy, &i.ResolvedBy, &i.OrgID, &i.HighRiskLocation, &i.SimulationMode,
			&i.DriverFeedback, &i.FeedbackNotes, &i.AcknowledgedAt, &i.ResolvedAt, &i.CreatedAt, &i.UpdatedAt, &i.DeviceID,
		)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, i)
	}
	return incidents, nil
}

func (r *emergencyRepository) ArchiveOldIncidents(ctx context.Context, olderThan time.Time) (int64, error) {
	// Move to archive table and delete from main table
	moveQuery := `
		INSERT INTO emergency_incidents_archive (
			id, vehicle_id, driver_id, trip_id, timestamp, latitude, longitude, location_address,
			status, severity, incident_type, panic_source, speed, heading, metadata,
			acknowledged_by, resolved_by, org_id, high_risk_location, simulation_mode,
			driver_feedback, feedback_notes, acknowledged_at, resolved_at, created_at, updated_at, device_id
		)
		SELECT
			id, vehicle_id, driver_id, trip_id, timestamp, latitude, longitude, location_address,
			status, severity, incident_type, panic_source, speed, heading, metadata,
			acknowledged_by, resolved_by, org_id, high_risk_location, simulation_mode,
			driver_feedback, feedback_notes, acknowledged_at, resolved_at, created_at, updated_at, device_id
		FROM sos_incidents
		WHERE (status = 'RESOLVED' OR status = 'FALSE_ALARM') AND updated_at < $1
	`
	start := time.Now()
	res, err := r.db.Exec(ctx, moveQuery, olderThan)
	sharedmetrics.RecordDBQuery(moveQuery, time.Since(start))
	if err != nil {
		return 0, err
	}

	count := res.RowsAffected()
	if count > 0 {
		deleteQuery := `DELETE FROM sos_incidents WHERE (status = 'RESOLVED' OR status = 'FALSE_ALARM') AND updated_at < $1`
		start2 := time.Now()
		_, err = r.db.Exec(ctx, deleteQuery, olderThan)
		sharedmetrics.RecordDBQuery(deleteQuery, time.Since(start2))
		if err != nil {
			return count, fmt.Errorf("archive succeeded but delete failed: %w", err)
		}
	}

	return count, nil
}
func (r *emergencyRepository) UpdateDriverFeedback(ctx context.Context, incidentID uuid.UUID, feedback string, notes string) error {
	query := `UPDATE sos_incidents SET driver_feedback = $2, feedback_notes = $3, updated_at = $4 WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, incidentID, feedback, notes, time.Now())
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) LogSLAViolation(ctx context.Context, v *models.SLAViolation) error {
	query := `
		INSERT INTO incident_sla_violations (id, incident_id, severity, breach_type, breach_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, v.ID, v.IncidentID, v.Severity, v.BreachType, v.BreachTime, v.CreatedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) ArchiveOldMedia(ctx context.Context, olderThan time.Time) (int64, error) {
	// 1 year retention for media as per enterprise requirement
	query := `DELETE FROM incident_media WHERE created_at < $1`
	start := time.Now()
	res, err := r.db.Exec(ctx, query, olderThan)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *emergencyRepository) RecordContactNotification(ctx context.Context, incidentID, contactID uuid.UUID) error {
	query := `
		INSERT INTO emergency_contact_responses (incident_id, contact_id, response_status, created_at)
		VALUES ($1, $2, 'NOTIFIED', NOW())
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, incidentID, contactID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) UpdateContactResponse(ctx context.Context, incidentID, contactID uuid.UUID, status string, notes string) error {
	query := `
		UPDATE emergency_contact_responses
		SET response_status = $1, responded_at = NOW(), notes = $2
		WHERE incident_id = $3 AND contact_id = $4
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, status, notes, incidentID, contactID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *emergencyRepository) GetTopLocations(ctx context.Context, orgID uuid.UUID, limit int) ([]string, error) {
	query := `
		SELECT location_address
		FROM sos_incidents
		WHERE org_id = $1 AND location_address IS NOT NULL AND location_address != ''
		GROUP BY location_address
		ORDER BY COUNT(*) DESC
		LIMIT $2
	`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID, limit)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []string
	for rows.Next() {
		var loc string
		if err := rows.Scan(&loc); err == nil {
			locations = append(locations, loc)
		}
	}
	return locations, nil
}

func (r *emergencyRepository) GetIncidentsCount(ctx context.Context, orgID uuid.UUID, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM sos_incidents WHERE org_id = $1 AND created_at >= $2`
	start := time.Now()
	var count int
	err := r.db.QueryRow(ctx, query, orgID, since).Scan(&count)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return count, err
}

// Phase 4: GetStalledIncidents returns TRIGGERED incidents not updated within stalledFor duration.

// CreateTimelineEvent creates a new timeline event
func (r *emergencyRepository) CreateTimelineEvent(ctx context.Context, event *models.TimelineEvent) error {
	query := `
		INSERT INTO timeline_events (id, incident_id, event_type, event_data, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, event.ID, event.IncidentID, event.EventType, event.EventData, event.CreatedAt)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

// GetTimelineEvents retrieves timeline events for an incident
func (r *emergencyRepository) GetTimelineEvents(ctx context.Context, incidentID uuid.UUID) ([]models.TimelineEvent, error) {
	query := `
		SELECT id, incident_id, event_type, event_data, created_at
		FROM timeline_events
		WHERE incident_id = $1
		ORDER BY created_at ASC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, incidentID)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.TimelineEvent
	for rows.Next() {
		var e models.TimelineEvent
		if err := rows.Scan(&e.ID, &e.IncidentID, &e.EventType, &e.EventData, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// EnqueueDeviceEvent queues an offline device event for retry
func (r *emergencyRepository) EnqueueDeviceEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload []byte) error {
	event := &models.DeviceEventQueueItem{
		ID:          uuid.New(),
		VehicleID:   vehicleID,
		EventType:   eventType,
		Payload:     payload,
		RetryCount:  0,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
		NextRetryAt: time.Now(),
	}
	query := `
		INSERT INTO device_event_queue (id, vehicle_id, event_type, payload, retry_count, status, created_at, next_retry_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		event.ID, event.VehicleID, event.EventType, event.Payload,
		event.RetryCount, event.Status, event.CreatedAt, event.NextRetryAt,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

// GetPendingDeviceEvents retrieves pending device events
func (r *emergencyRepository) GetPendingDeviceEvents(ctx context.Context, limit int) ([]models.DeviceEventQueueItem, error) {
	query := `
		SELECT id, vehicle_id, event_type, payload, retry_count, status, created_at, next_retry_at
		FROM device_event_queue
		WHERE status = 'PENDING' AND next_retry_at <= NOW()
		ORDER BY created_at ASC
		LIMIT $1`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, limit)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.DeviceEventQueueItem
	for rows.Next() {
		var e models.DeviceEventQueueItem
		if err := rows.Scan(&e.ID, &e.VehicleID, &e.EventType, &e.Payload, &e.RetryCount, &e.Status, &e.CreatedAt, &e.NextRetryAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// MarkDeviceEventDelivered marks a device event as delivered
func (r *emergencyRepository) MarkDeviceEventDelivered(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE device_event_queue SET status = 'DELIVERED', updated_at = NOW() WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

// IncrementDeviceEventRetry increments retry count and sets next retry time
func (r *emergencyRepository) IncrementDeviceEventRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	query := `UPDATE device_event_queue SET retry_count = retry_count + 1, next_retry_at = $1, updated_at = NOW() WHERE id = $2`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, nextRetryAt, id)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

// GetStalledIncidents returns TRIGGERED incidents not updated within stalledFor duration
func (r *emergencyRepository) GetStalledIncidents(ctx context.Context, stalledFor time.Duration) ([]models.SOSIncident, error) {
	query := `
		SELECT i.id, i.vehicle_id, i.driver_id, i.trip_id, i.device_id, i.org_id, i.timestamp, i.latitude, i.longitude,
			   i.location_address, i.status, i.severity, i.incident_type, i.panic_source, i.speed, i.heading,
			   i.metadata, i.acknowledged_by, i.resolved_by, i.created_at, i.updated_at
		FROM sos_incidents i
		WHERE i.status IN ('TRIGGERED', 'ACKNOWLEDGED')
		  AND i.updated_at < NOW() - $1
		ORDER BY i.created_at ASC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, stalledFor)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.SOSIncident
	for rows.Next() {
		var i models.SOSIncident
		if err := rows.Scan(
			&i.ID, &i.VehicleID, &i.DriverID, &i.TripID, &i.DeviceID, &i.OrgID,
			&i.Timestamp, &i.Latitude, &i.Longitude, &i.LocationAddress,
			&i.Status, &i.Severity, &i.IncidentType, &i.PanicSource,
			&i.Speed, &i.Heading, &i.Metadata, &i.AcknowledgedBy, &i.ResolvedBy,
			&i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		incidents = append(incidents, i)
	}
	return incidents, nil
}

// GetAuthoritiesByLocation finds emergency response units that could cover the given location
func (r *emergencyRepository) GetAuthoritiesByLocation(ctx context.Context, lat, lng float64) ([]models.EmergencyResponseUnit, error) {
	// Simplified: return all active units. In production, this would use geospatial queries.
	query := `
		SELECT id, org_id, unit_type, name, phone, coverage_city, coverage_state, is_active, type, contact_number, api_endpoint, created_at
		FROM emergency_response_units
		WHERE is_active = true
		ORDER BY RANDOM()`
	start := time.Now()
	rows, err := r.db.Query(ctx, query)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []models.EmergencyResponseUnit
	for rows.Next() {
		var u models.EmergencyResponseUnit
		if err := rows.Scan(
			&u.ID, &u.OrgID, &u.UnitType, &u.Name, &u.Phone,
			&u.CoverageCity, &u.CoverageState, &u.IsActive, &u.Type, &u.ContactNumber, &u.APIEndpoint, &u.CreatedAt,
		); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

// LogAuthorityDelivery records an attempt to notify an emergency authority
func (r *emergencyRepository) LogAuthorityDelivery(ctx context.Context, log *models.AuthorityDeliveryLog) error {
	query := `
		INSERT INTO authority_delivery_logs (id, incident_id, authority_type, delivery_status, response_code, response_body, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	start := time.Now()
	now := time.Now()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}
	if log.UpdatedAt.IsZero() {
		log.UpdatedAt = now
	}
	_, err := r.db.Exec(ctx, query,
		log.ID, log.IncidentID, log.AuthorityType, log.DeliveryStatus, log.ResponseCode,
		log.ResponseBody, log.RetryCount, log.CreatedAt, log.UpdatedAt,
	)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}

// UpdateAuthorityDelivery updates the status of an authority delivery attempt
func (r *emergencyRepository) UpdateAuthorityDelivery(ctx context.Context, id uuid.UUID, status string, responseCode int, responseBody string) error {
	query := `
		UPDATE authority_delivery_logs
		SET delivery_status = $1, response_code = $2, response_body = $3, updated_at = NOW()
		WHERE id = $4`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, status, responseCode, responseBody, id)
	sharedmetrics.RecordDBQuery(query, time.Since(start))
	return err
}
