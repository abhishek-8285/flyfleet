package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	trackingRepo "github.com/flyfleet/fleet-core-service/internal/domains/tracking/repository"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
)

const (
	// Duplicate SOS prevention window (minutes)
	DuplicateSOSWindow = 5
	// Escalation timeout (minutes)
	EscalationTimeout = 2 * time.Minute
)

type EmergencyService struct {
	repo          repository.EmergencyRepository
	trackingRepo  trackingRepo.LiveCacheRepository
	nc            *nats.Conn
	redis         *redis.Client
	log           *zap.Logger
	retry         *NotificationRetryService
	geo            *EmergencyGeoService
	authority      *EmergencyAuthorityService
	severityEngine *EmergencySeverityEngine
	fallback       *ProviderFallbackManager
}

func (s *EmergencyService) ValidateConfig() error {
	requiredEnv := []string{
		"EMERGENCY_SMS_PROVIDER",
		"EMERGENCY_CALL_PROVIDER",
		"NATS_URL",
		"REDIS_URL",
		"POSTGRES_URL",
	}

	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			return fmt.Errorf("missing required environment variable: %s", env)
		}
	}
	return nil
}

func (s *EmergencyService) CheckHealth(ctx context.Context) (map[string]string, error) {
	health := make(map[string]string)
	
	// Providers
	if s.fallback != nil {
		providerHealth := s.fallback.CheckProvidersHealth(ctx)
		for k, v := range providerHealth {
			health["provider_"+k] = v
		}
	} else {
		health["providers"] = "not_configured"
	}

	// Database
	if err := s.repo.Ping(ctx); err != nil {
		health["database"] = "unhealthy: " + err.Error()
	} else {
		health["database"] = "healthy"
	}

	// Redis
	if s.redis != nil {
		if err := s.redis.Ping(ctx).Err(); err != nil {
			health["redis"] = "unhealthy: " + err.Error()
		} else {
			health["redis"] = "healthy"
		}
	} else {
		health["redis"] = "not_configured"
	}

	// NATS
	if s.nc != nil {
		if s.nc.Status() != nats.CONNECTED {
			health["nats"] = "unhealthy: status " + s.nc.Status().String()
		} else {
			health["nats"] = "healthy"
		}
	} else {
		health["nats"] = "not_configured"
	}

	return health, nil
}

func NewEmergencyService(
	repo repository.EmergencyRepository,
	trackingRepo trackingRepo.LiveCacheRepository,
	nc *nats.Conn,
	rdb *redis.Client,
	log *zap.Logger,
	retry *NotificationRetryService,
	geo *EmergencyGeoService,
	authority *EmergencyAuthorityService,
	severityEngine *EmergencySeverityEngine,
	fallback *ProviderFallbackManager,
) *EmergencyService {
	return &EmergencyService{
		repo:           repo,
		trackingRepo:   trackingRepo,
		nc:             nc,
		redis:          rdb,
		log:            log,
		retry:          retry,
		geo:            geo,
		authority:      authority,
		severityEngine: severityEngine,
		fallback:       fallback,
	}
}

func (s *EmergencyService) logEnhancedTimelineEvent(ctx context.Context, incidentID uuid.UUID, eventType string, data interface{}) {
	eventData, _ := json.Marshal(data)
	event := &models.TimelineEvent{
		ID:         uuid.New(),
		IncidentID: incidentID,
		EventType:  eventType,
		EventData:  eventData,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateTimelineEvent(ctx, event); err != nil {
		s.log.Error("Failed to log enhanced timeline event", zap.Error(err))
	}
	emergency.TimelineEventsTotal.WithLabelValues(eventType).Inc()
}

func (s *EmergencyService) CalculateSeverity(incidentType models.IncidentType) models.SOSSeverity {
	switch incidentType {
	case models.TypePanic, models.TypeVehicleAccident, models.TypeMedicalEmergency, models.TypeFireAlert, models.TypeDriverAssault:
		return models.SeverityCritical
	case models.TypeSecurityThreat, models.TypeVehicleTheft:
		return models.SeverityHigh
	case models.TypeVehicleBreakdown:
		return models.SeverityMedium
	default:
		return models.SeverityLow
	}
}

func (s *EmergencyService) TriggerSOS(ctx context.Context, req models.CreateSOSRequest) (*models.SOSIncident, error) {
	// Validate vehicle ID
	if req.VehicleID == "" {
		return nil, ErrMissingVehicleID
	}

	vehicleID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle_id: %w", err)
	}

	// Validate coordinates
	if err := validateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}

	// Default incident type if empty
	if req.IncidentType == "" {
		req.IncidentType = models.TypePanic
	}

	// Auto-calculate severity for standardization
	// 1.3 Determine severity automatically (Phase 3)
	collisionDetected := false // Check if req.Metadata indicates collision
	panicButtonPressed := true 
	driverHealthSignal := true
	severity := s.severityEngine.Classify(float64(req.Speed), collisionDetected, panicButtonPressed, driverHealthSignal)
	
	// Record severity metric
	emergency.IncidentSeverityTotal.WithLabelValues(string(severity)).Inc()

	// Phase 3: Check for duplicate SOS within window using Redis
	if s.redis != nil {
		dedupKey := fmt.Sprintf("sos_dedup:%s", vehicleID)
		set, err := s.redis.SetNX(ctx, dedupKey, "triggered", 5*time.Minute).Result()
		if err != nil {
			s.log.Error("Redis deduplication check failed", zap.Error(err))
		} else if !set {
			return nil, ErrDuplicateSOS
		}
	} else {
		s.log.Warn("Redis not initialized, skipping deduplication check")
	}

	// SOS Rate limiting (5 incidents / 10 mins)
	limited, err := s.isRateLimited(ctx, vehicleID)
	if err != nil {
		s.log.Error("Rate limit check failed", zap.Error(err))
	} else if limited {
		return nil, fmt.Errorf("SOS rate limit exceeded for vehicle")
	}

	// 1. Create Incident Model
	// 1. Create Incident Model
	incident := &models.SOSIncident{
		ID:              uuid.New(),
		VehicleID:       vehicleID,
		OrgID:           req.OrgID,
		Status:          models.StatusTriggered,
		Severity:        severity,
		IncidentType:    req.IncidentType,
		PanicSource:     req.PanicSource,
		Timestamp:       time.Now(),
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		LocationAddress: &req.LocationAddress,
		Speed:           req.Speed,
		Heading:         req.Heading,
		Metadata:        req.Metadata,
		SimulationMode:  req.SimulationMode,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Phase 5: Geo Context Enrichment
	if s.geo != nil {
		geoContext := s.geo.EnrichWithNearbyResources(ctx, incident.Latitude, incident.Longitude)
		// Merge with existing metadata if any
		meta := make(map[string]interface{})
		if len(incident.Metadata) > 0 {
			_ = json.Unmarshal(incident.Metadata, &meta)
			if meta == nil {
				meta = make(map[string]interface{})
			}
		}
		for k, v := range geoContext {
			meta[k] = v
		}
		data, _ := json.Marshal(meta)
		incident.Metadata = data
	}

	// 1.5 Geo Risk Tagging (Rule-based Enterprise logic)
	if s.isHighRiskArea(incident.Latitude, incident.Longitude) {
		incident.HighRiskLocation = true
		s.log.Warn("Incident in High Risk Zone detected", zap.Float64("lat", incident.Latitude))
	}

	// Dynamic associations
	if req.DeviceID != "" {
		if id, err := uuid.Parse(req.DeviceID); err == nil {
			incident.DeviceID = &id
		}
	}
	if req.DriverID != "" {
		if id, err := uuid.Parse(req.DriverID); err == nil {
			incident.DriverID = &id
		}
	}
	if req.TripID != "" {
		if id, err := uuid.Parse(req.TripID); err == nil {
			incident.TripID = &id
		}
	}

	// Default Panic Source
	if incident.PanicSource == "" {
		incident.PanicSource = models.SourceMobile
	}

	// 2. Persist to DB
	if err := s.repo.CreateIncident(ctx, incident); err != nil {
		s.log.Error("Failed to create SOS incident", zap.Error(err))
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	// 3. Log to action log (legacy) and timeline (new)
	s.logAction(ctx, incident.ID, nil, "TRIGGERED", fmt.Sprintf("Incident %s triggered via %s", incident.IncidentType, incident.PanicSource))
	s.logTimelineEvent(ctx, incident.ID, "TRIGGERED", "SYSTEM", map[string]interface{}{
		"source":   incident.PanicSource,
		"severity": incident.Severity,
		"type":     incident.IncidentType,
	})
	s.logEnhancedTimelineEvent(ctx, incident.ID, "SOS_TRIGGERED", map[string]interface{}{
		"source":   incident.PanicSource,
		"severity": incident.Severity,
		"type":     incident.IncidentType,
		"location": map[string]float64{"lat": incident.Latitude, "lng": incident.Longitude},
	})

	// 4. Metrics
	emergency.IncidentsTotal.WithLabelValues(incident.OrgID.String(), string(incident.IncidentType), string(incident.Severity)).Inc()
	emergency.ActiveIncidents.WithLabelValues(incident.OrgID.String()).Inc()

	// 4.5 Capture vehicle snapshot (best effort)
	if s.trackingRepo != nil {
		if liveLoc, err := s.trackingRepo.GetLiveLocation(ctx, vehicleID); err == nil && liveLoc != nil {
			snapshot := &models.IncidentVehicleSnapshot{
				IncidentID:   incident.ID,
				Speed:        int(liveLoc.Speed),
				Heading:      liveLoc.Heading,
				EngineStatus: liveLoc.Ignition,
				CreatedAt:    time.Now(),
			}
			if err := s.repo.CreateVehicleSnapshot(ctx, snapshot); err != nil {
				s.log.Error("Failed to save vehicle snapshot", zap.Error(err), zap.String("incident_id", incident.ID.String()))
			}
		}
	}

	// 5. Publish to NATS (Real-time Alert)
	s.publishEvent("emergency.sos.triggered", incident)

	// 6. Start escalation chain
	go s.startEscalationChain(context.Background(), incident)

	// 7. Initial Mobile Feedback
	s.notifyDriverApp(ctx, incident.DriverID, &incident.VehicleID, "help_on_way", "Distress signal received. Local emergency services have been alerted.")

	return incident, nil
}

func (s *EmergencyService) startEscalationChain(ctx context.Context, incident *models.SOSIncident) {
	s.logTimelineEvent(ctx, incident.ID, "ESCALATION_STARTED", "SYSTEM", map[string]interface{}{
		"timeout": EscalationTimeout.Seconds(),
	})
	s.logEnhancedTimelineEvent(ctx, incident.ID, "ESCALATION_STARTED", map[string]interface{}{
		"timeout": EscalationTimeout.Seconds(),
	})

	contacts, err := s.repo.GetEmergencyContacts(ctx, incident.OrgID)
	if err != nil {
		s.log.Error("Failed to get emergency contacts for escalation", zap.Error(err))
		return
	}

	if len(contacts) == 0 {
		s.log.Warn("No emergency contacts found for organization", zap.String("org_id", incident.OrgID.String()))
		s.logTimelineEvent(ctx, incident.ID, "ESCALATION_FAILED", "SYSTEM", map[string]string{"reason": "no_contacts"})

		// Fallback for critical
		if incident.Severity == models.SeverityCritical {
			s.routeToAuthorities(ctx, incident)
		}
		return
	}

	for _, contact := range contacts {
		// Check current status
		inc, err := s.repo.GetIncidentByID(ctx, incident.ID)
		if err != nil || inc == nil || (inc.Status != models.StatusTriggered && inc.Status != models.StatusAcknowledged) {
			return // Resolved or cancelled
		}

		// Phase 1: Record that we are notifying this contact
		s.repo.RecordContactNotification(ctx, incident.ID, contact.ID)

		escalation := &models.EmergencyEscalation{
			ID:                  uuid.New(),
			IncidentID:          incident.ID,
			ContactID:           contact.ID,
			NotificationChannel: "SMS",
			Status:              string(models.EscalationNotified),
			AttemptedAt:         time.Now(),
		}
		s.repo.CreateEscalation(ctx, escalation)

		// Phase 2: Use NotificationRetryService
		if s.retry != nil {
			s.retry.SendWithRetry(ctx, NotificationRequest{
				IncidentID: incident.ID,
				OrgID:      incident.OrgID,
				ContactID:  contact.ID,
				Type:       "SMS",
				Target:     contact.Phone,
				Content:    fmt.Sprintf("EMERGENCY: SOS triggered for Vehicle %s. Please acknowledge.", incident.VehicleID),
			})
		}

		// Cost Monitoring
		emergency.SMSSentTotal.WithLabelValues(incident.OrgID.String()).Inc()

		s.logTimelineEvent(ctx, incident.ID, "CONTACT_NOTIFIED", "SYSTEM", map[string]interface{}{
			"contact_name": contact.Name,
			"priority":     contact.Priority,
			"channel":      "SMS",
		})
		s.logEnhancedTimelineEvent(ctx, incident.ID, "CONTACT_NOTIFIED", map[string]interface{}{
			"contact_name": contact.Name,
			"priority":     contact.Priority,
			"channel":      "SMS",
		})

		s.publishEvent("emergency.sos.notified", map[string]interface{}{
			"incident_id": incident.ID,
			"contact":     contact.Name,
		})

		// Wait for response
		time.Sleep(EscalationTimeout)

		// Re-check after timeout
		currentInc, _ := s.repo.GetIncidentByID(ctx, incident.ID)
		if currentInc != nil && currentInc.Status == models.StatusAcknowledged {
			s.logTimelineEvent(ctx, incident.ID, "ESCALATION_RESPONDED", contact.Name, nil)
			return // Escalation successful (responded)
		}

		s.logTimelineEvent(ctx, incident.ID, "ESCALATION_NOT_RESPONDED", "SYSTEM", map[string]string{
			"contact": contact.Name,
		})
	}

	// All contacts failed to respond
	s.logTimelineEvent(ctx, incident.ID, "ESCALATION_ESCALATED", "SYSTEM", nil)
	s.publishEvent("emergency.sos.escalated", incident)

	// If still triggered after all contacts, notify authorities if critical
	if incident.Severity == models.SeverityCritical {
		s.routeToAuthorities(ctx, incident)
		// Phase 2: AIS-140 authority service – skip for drills/simulations
		if s.authority != nil && !incident.SimulationMode {
			s.authority.NotifyAuthorities(ctx, incident)
		}
	}
}

func (s *EmergencyService) routeToAuthorities(ctx context.Context, incident *models.SOSIncident) {
	s.log.Warn("Escalating to Indian Emergency Response Units", zap.String("incident", incident.ID.String()))

	units, err := s.repo.ListResponseUnits(ctx, incident.OrgID)
	if err != nil {
		s.log.Error("Failed to list response units", zap.Error(err))
		return
	}

	for _, unit := range units {
		shouldNotify := false
		switch incident.Severity {
		case models.SeverityCritical:
			if unit.Type == "POLICE" || unit.Type == "AMBULANCE" || unit.Type == "FIRE" {
				shouldNotify = true
			}
		case models.SeverityHigh:
			if unit.Type == "POLICE" || unit.Type == "ROAD_ASSISTANCE" {
				shouldNotify = true
			}
		case models.SeverityMedium:
			if incident.IncidentType == models.TypeVehicleBreakdown && unit.Type == "ROAD_ASSISTANCE" {
				shouldNotify = true
			}
		}

		if shouldNotify {
			s.log.Info("Notifying authority unit", zap.String("unit", unit.Name), zap.String("type", unit.Type))
			s.logTimelineEvent(ctx, incident.ID, "AUTHORITY_NOTIFIED", "SYSTEM", map[string]interface{}{
				"unit_name": unit.Name,
				"unit_type": unit.Type,
				"contact":   unit.ContactNumber,
			})
			s.logEnhancedTimelineEvent(ctx, incident.ID, "AUTHORITY_ALERTED", map[string]interface{}{
				"unit_name": unit.Name,
				"unit_type": unit.Type,
				"contact":   unit.ContactNumber,
			})
			// In production: trigger actual API call to ERU system or SMS gateway
		}
	}
}

// validateCoordinates validates GPS coordinates
func validateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return NewCoordinateError(ErrInvalidLatitude.Error())
	}
	if lon < -180 || lon > 180 {
		return NewCoordinateError(ErrInvalidLongitude.Error())
	}
	return nil
}

// notifyEmergencyContacts sends notifications to emergency contacts
func (s *EmergencyService) notifyEmergencyContacts(ctx context.Context, vehicleID uuid.UUID, incident *models.SOSIncident) {
	// Get organization ID from vehicle (would need vehicle repository)
	// For now, this is a placeholder
	s.log.Info("Emergency contact notification triggered",
		zap.String("vehicle_id", vehicleID.String()),
		zap.String("incident_id", incident.ID.String()),
	)

	// In production:
	// 1. Get organization from vehicle
	// 2. Get emergency contacts for organization
	// 3. Send SMS/call to contacts in priority order
	// 4. Log notification attempts
}

func (s *EmergencyService) GetIncident(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.SOSIncident, error) {
	incident, err := s.repo.GetIncidentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, ErrIncidentNotFound
	}
	if incident.OrgID != orgID {
		return nil, ErrUnauthorized
	}
	return incident, nil
}

func (s *EmergencyService) ListActiveIncidents(ctx context.Context, orgID uuid.UUID) ([]models.SOSIncident, error) {
	return s.repo.GetActiveIncidentsByOrg(ctx, orgID)
}

func (s *EmergencyService) GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	return s.repo.GetEmergencyContacts(ctx, orgID)
}

func (s *EmergencyService) GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentTimeline, error) {
	return s.repo.GetTimeline(ctx, incidentID)
}

func (s *EmergencyService) logTimelineEvent(ctx context.Context, incidentID uuid.UUID, eventType, actor string, metadata interface{}) {
	metaJSON, _ := json.Marshal(metadata)
	event := &models.IncidentTimeline{
		ID:         uuid.New(),
		IncidentID: incidentID,
		EventType:  eventType,
		Actor:      actor,
		Metadata:   metaJSON,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.LogTimelineEvent(ctx, event); err != nil {
		s.log.Error("Failed to log timeline event", zap.Error(err), zap.String("eventType", eventType))
	}
}

func (s *EmergencyService) AcknowledgeIncident(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID, notes string) error {
	incident, err := s.repo.GetIncidentByID(ctx, id)
	if err != nil {
		return err
	}
	if incident == nil {
		return ErrIncidentNotFound
	}
	if incident.OrgID != orgID {
		return ErrUnauthorized
	}

	if incident.Status != models.StatusTriggered {
		return ErrIncidentAlreadyAcknowledged
	}

	if err := s.repo.UpdateIncidentStatus(ctx, id, models.StatusAcknowledged, &userID); err != nil {
		return err
	}

	s.logAction(ctx, id, &userID, "ACKNOWLEDGED", notes)
	s.logTimelineEvent(ctx, id, "ACKNOWLEDGED", userID.String(), map[string]string{"notes": notes})
	s.logEnhancedTimelineEvent(ctx, id, "ACKNOWLEDGED", map[string]interface{}{
		"user_id": userID,
		"notes":   notes,
	})

	// Publish update
	incident.Status = models.StatusAcknowledged
	s.publishEvent("emergency.sos.acknowledged", incident)

	// 4. Notify Driver App (Enterprise Requirement)
	s.notifyDriverApp(ctx, incident.DriverID, &incident.VehicleID, "incident_acknowledged", "Help is on the way. Incident acknowledged.")

	return nil
}

func (s *EmergencyService) ResolveIncident(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID, notes string, status models.SOSStatus) error {
	if status != models.StatusResolved && status != models.StatusFalseAlarm {
		return ErrInvalidStatusTransition
	}

	incident, err := s.repo.GetIncidentByID(ctx, id)
	if err != nil {
		return err
	}
	if incident == nil {
		return ErrIncidentNotFound
	}
	if incident.OrgID != orgID {
		return ErrUnauthorized
	}

	if incident.Status == models.StatusResolved || incident.Status == models.StatusFalseAlarm {
		return ErrIncidentAlreadyResolved
	}

	if err := s.repo.UpdateIncidentStatus(ctx, id, status, &userID); err != nil {
		return err
	}

	s.logAction(ctx, id, &userID, string(status), notes)
	s.logTimelineEvent(ctx, id, string(status), userID.String(), map[string]string{"notes": notes})
	s.logEnhancedTimelineEvent(ctx, id, string(status), map[string]interface{}{
		"user_id": userID,
		"notes":   notes,
	})

	// Publish update
	incident.Status = status
	s.publishEvent("emergency.sos.resolved", incident)

	// 4. Notify Driver App (Enterprise Requirement)
	s.notifyDriverApp(ctx, incident.DriverID, &incident.VehicleID, "incident_resolved", "Incident has been resolved. Safe travels.")

	return nil
}

func (s *EmergencyService) FleetHalt(ctx context.Context, orgID uuid.UUID, reason string, userName string) error {
	// 1. Get all active vehicles for org (Simplified: we'd usually use a vehicle service)
	// For production, we publish a global NATS command that all edge devices/gateways listen to
	cmd := map[string]interface{}{
		"org_id":    orgID,
		"command":   "HALT",
		"reason":    reason,
		"issued_by": userName,
		"timestamp": time.Now(),
	}

	s.publishEvent("fleet.cmd.halt", cmd)
	return nil
}

func (s *EmergencyService) FleetResume(ctx context.Context, orgID uuid.UUID, userName string) error {
	cmd := map[string]interface{}{
		"org_id":    orgID,
		"command":   "RESUME",
		"issued_by": userName,
		"timestamp": time.Now(),
	}

	s.publishEvent("fleet.cmd.resume", cmd)
	return nil
}

func (s *EmergencyService) GetIncidentMetrics(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*repository.IncidentMetrics, error) {
	return s.repo.GetIncidentMetrics(ctx, orgID, startDate, endDate)
}

func (s *EmergencyService) logAction(ctx context.Context, incidentID uuid.UUID, userID *uuid.UUID, action, notes string) {
	log := &models.IncidentLog{
		ID:         uuid.New(),
		IncidentID: incidentID,
		UserID:     userID,
		Action:     action,
		Notes:      notes,
		Timestamp:  time.Now(),
	}
	if err := s.repo.LogIncidentAction(ctx, log); err != nil {
		s.log.Error("Failed to log incident action", zap.Error(err))
	}
}

func (s *EmergencyService) publishEvent(subject string, data interface{}) {
	payload, _ := json.Marshal(data)

	// Primary: NATS
	if s.nc != nil {
		if err := s.nc.Publish(subject, payload); err == nil {
			return
		} else {
			s.log.Error("Failed to publish NATS event, falling back to Redis", zap.String("subject", subject), zap.Error(err))
		}
	}

	// Fallback: Redis (LPUSH to a queue for later processing)
	if s.redis != nil {
		if err := s.redis.LPush(context.Background(), "emergency:events:retry_queue", payload).Err(); err != nil {
			s.log.Error("Failed to publish Redis fallback event", zap.String("subject", subject), zap.Error(err))
		}
	}
}

func (s *EmergencyService) isRateLimited(ctx context.Context, vehicleID uuid.UUID) (bool, error) {
	if s.redis == nil {
		return false, nil
	}
	key := fmt.Sprintf("emergency:ratelimit:%s", vehicleID)
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		s.redis.Expire(ctx, key, 10*time.Minute)
	}
	return count > 5, nil
}

func (s *EmergencyService) AddMedia(ctx context.Context, media models.IncidentMedia, orgID uuid.UUID) error {
	// Security check: ensure incident belongs to org
	incident, err := s.repo.GetIncidentByID(ctx, media.IncidentID)
	if err != nil || incident == nil {
		return ErrIncidentNotFound
	}
	if incident.OrgID != orgID {
		return ErrUnauthorized
	}

	if err := s.repo.CreateMedia(ctx, &media); err != nil {
		return fmt.Errorf("failed to save media: %w", err)
	}

	s.logTimelineEvent(ctx, media.IncidentID, "MEDIA_UPLOADED", "USER", map[string]interface{}{
		"media_type": media.MediaType,
		"file_url":   media.FileURL,
	})

	return nil
}

func (s *EmergencyService) TriggerSimulatedSOS(ctx context.Context, req models.CreateSOSRequest) (*models.SOSIncident, error) {
	req.SimulationMode = true
	meta := make(map[string]interface{})
	if len(req.Metadata) > 0 {
		json.Unmarshal(req.Metadata, &meta)
	}
	meta["simulated"] = true
	req.Metadata, _ = json.Marshal(meta)
	return s.TriggerSOS(ctx, req)
}

// TriggerDrill creates a drill incident: simulation_mode=true, metrics recorded, authority alerts suppressed.
func (s *EmergencyService) TriggerDrill(ctx context.Context, req models.CreateSOSRequest) (*models.SOSIncident, error) {
	req.SimulationMode = true

	meta := make(map[string]interface{})
	if len(req.Metadata) > 0 {
		json.Unmarshal(req.Metadata, &meta)
	}
	meta["drill"] = true
	meta["alerts_suppressed"] = true
	req.Metadata, _ = json.Marshal(meta)

	// Use TriggerSOS but the authority service will check simulation_mode before dispatching
	incident, err := s.TriggerSOS(ctx, req)
	if err != nil {
		return nil, err
	}

	// Phase 8: Record drill metric
	emergency.DrillEventsTotal.WithLabelValues(incident.OrgID.String()).Inc()

	s.log.Info("Emergency drill completed",
		zap.String("incident_id", incident.ID.String()),
		zap.String("org_id", incident.OrgID.String()),
	)
	return incident, nil
}

func (s *EmergencyService) UpdateDriverFeedback(ctx context.Context, incidentID uuid.UUID, orgID uuid.UUID, feedback string, notes string) error {
	// Security check: ensure incident belongs to org
	incident, err := s.repo.GetIncidentByID(ctx, incidentID)
	if err != nil || incident == nil {
		return ErrIncidentNotFound
	}
	if incident.OrgID != orgID {
		return ErrUnauthorized
	}

	if err := s.repo.UpdateDriverFeedback(ctx, incidentID, feedback, notes); err != nil {
		return fmt.Errorf("failed to update driver feedback: %w", err)
	}

	s.logTimelineEvent(ctx, incidentID, "DRIVER_FEEDBACK", "DRIVER", map[string]interface{}{
		"feedback": feedback,
		"notes":    notes,
	})

	return nil
}
func (s *EmergencyService) notifyDriverApp(ctx context.Context, driverID, vehicleID *uuid.UUID, eventType, message string) {
	// In a real system, this would push to Firebase (FCM) or a WebSocket hub.
	// We simulate this by publishing a specialized event.
	s.log.Info("Pushing notification to driver app",
		zap.String("event", eventType),
		zap.String("message", message))

	payload := map[string]interface{}{
		"event_type": eventType,
		"message":    message,
		"timestamp":  time.Now(),
	}
	if driverID != nil {
		payload["driver_id"] = *driverID
	}
	if vehicleID != nil {
		payload["vehicle_id"] = *vehicleID
	}

	s.publishEvent("mobile.push.emergency_update", payload)
}

func (s *EmergencyService) ExportIncidents(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]models.SOSIncident, error) {
	return s.repo.ExportIncidents(ctx, orgID, start, end)
}

func (s *EmergencyService) GetIncidentTimeline(ctx context.Context, incidentID, orgID uuid.UUID) ([]models.TimelineEvent, error) {
	// Security check
	incident, err := s.repo.GetIncidentByID(ctx, incidentID)
	if err != nil || incident == nil {
		return nil, ErrIncidentNotFound
	}
	if incident.OrgID != orgID {
		return nil, ErrUnauthorized
	}

	return s.repo.GetTimelineEvents(ctx, incidentID)
}

func (s *EmergencyService) isHighRiskArea(lat, lng float64) bool {
	// Simulated Geofencing rules for Highway/Industrial/Border zones
	// NH8 corridor (Gurgaon - Manesar)
	if lat > 28.3 && lat < 28.5 && lng > 76.8 && lng < 77.1 {
		return true
	}
	// Mumbai - Pune Expressway
	if lat > 18.5 && lat < 19.2 && lng > 73.0 && lng < 73.9 {
		return true
	}
	return false
}
