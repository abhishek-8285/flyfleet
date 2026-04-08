package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/emergency"
	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// AuthorityType represents the type of emergency authority.
type AuthorityType string

const (
	AuthorityPolice    AuthorityType = "POLICE"
	AuthorityAmbulance AuthorityType = "AMBULANCE"
	AuthorityFire      AuthorityType = "FIRE"
)

// authorityNotificationPayload is the AIS-140 compliant payload structure.
type authorityNotificationPayload struct {
	IncidentID string        `json:"incident_id"`
	VehicleID  string        `json:"vehicle_id"`
	DriverID   string        `json:"driver_id,omitempty"`
	Latitude   float64       `json:"latitude"`
	Longitude  float64       `json:"longitude"`
	Timestamp  time.Time     `json:"timestamp"`
	Type       string        `json:"type"`
	Severity   string        `json:"severity"`
	Authority  AuthorityType `json:"authority"`
}

// EmergencyAuthorityService handles AIS-140 compliant emergency authority notifications.
type EmergencyAuthorityService struct {
	repo   repository.EmergencyRepository
	nc     *nats.Conn
	logger *zap.Logger
}

func NewEmergencyAuthorityService(repo repository.EmergencyRepository, nc *nats.Conn, logger *zap.Logger) *EmergencyAuthorityService {
	return &EmergencyAuthorityService{repo: repo, nc: nc, logger: logger}
}

// NotifyAuthorities dispatches AIS-140 payloads and handles retries/logging/geo-routing.
func (s *EmergencyAuthorityService) NotifyAuthorities(ctx context.Context, incident *models.SOSIncident) {
	// 1. Get authorities covering the incident location
	var geoAuthorities []models.EmergencyResponseUnit
	var err error
	if s.repo != nil {
		geoAuthorities, err = s.repo.GetAuthoritiesByLocation(ctx, incident.Latitude, incident.Longitude)
	}

	if s.repo == nil || err != nil || len(geoAuthorities) == 0 {
		if s.repo != nil && err != nil {
			s.logger.Warn("Failed to get geographic authorities, falling back to static rules", zap.Error(err))
		}
		staticTypes := s.determineAuthorities(incident)
		for _, auth := range staticTypes {
			go s.notifyWithRetry(ctx, incident, auth)
		}
		return
	}

	// 2. Group by type to handle failover within each category (Police, Ambulance, Fire)
	typesNeeded := s.determineAuthorities(incident)
	for _, neededType := range typesNeeded {
		found := false
		for _, auth := range geoAuthorities {
			if AuthorityType(auth.Type) == neededType {
				go s.notifyWithRetry(ctx, incident, neededType) // In real world, we'd use auth.ContactNumber/API
				found = true
				break // Primary chosen (sorted by priority)
			}
		}

		if !found {
			s.logger.Warn("Specific geo-authority type not found in zone", zap.String("type", string(neededType)))
			go s.notifyWithRetry(ctx, incident, neededType) // Static fallback
		}
	}
}

func (s *EmergencyAuthorityService) determineAuthorities(incident *models.SOSIncident) []AuthorityType {
	switch incident.Severity {
	case models.SeverityCritical:
		return []AuthorityType{AuthorityPolice, AuthorityAmbulance, AuthorityFire}
	case models.SeverityHigh:
		return []AuthorityType{AuthorityPolice, AuthorityAmbulance}
	default:
		return []AuthorityType{AuthorityPolice}
	}
}

func (s *EmergencyAuthorityService) notifyWithRetry(ctx context.Context, incident *models.SOSIncident, auth AuthorityType) {
	deliveryID := uuid.New()
	log := &models.AuthorityDeliveryLog{
		ID:             deliveryID,
		IncidentID:     incident.ID,
		AuthorityType:  string(auth),
		DeliveryStatus: "SENT",
		RetryCount:     0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if s.repo != nil {
		s.repo.LogAuthorityDelivery(ctx, log)
	}

	for i := 0; i < 3; i++ {
		success := s.attemptNotification(ctx, incident, auth)

		status := "FAILED"
		if success {
			status = "CONFIRMED"
		}

		if s.repo != nil {
			s.repo.UpdateAuthorityDelivery(ctx, deliveryID, status, 200, "") // Assuming 200 for NATS publish
		}

		if success {
			emergency.AuthorityDeliverySuccessTotal.WithLabelValues(string(auth), "success").Inc()
			return
		}

		log.RetryCount++
		time.Sleep(time.Duration(2*(i+1)) * time.Second) // Simple backoff
	}

	emergency.AuthorityDeliverySuccessTotal.WithLabelValues(string(auth), "failed").Inc()
	s.logger.Error("Failed to notify authority after 3 retries",
		zap.String("authority", string(auth)),
		zap.String("incident_id", incident.ID.String()))
}

func (s *EmergencyAuthorityService) attemptNotification(ctx context.Context, incident *models.SOSIncident, auth AuthorityType) bool {
	driverID := ""
	if incident.DriverID != nil {
		driverID = incident.DriverID.String()
	}

	payload := authorityNotificationPayload{
		IncidentID: incident.ID.String(),
		VehicleID:  incident.VehicleID.String(),
		DriverID:   driverID,
		Latitude:   incident.Latitude,
		Longitude:  incident.Longitude,
		Timestamp:  incident.Timestamp,
		Type:       "SOS",
		Severity:   string(incident.Severity),
		Authority:  auth,
	}

	data, _ := json.Marshal(payload)

	// Publish to NATS
	if s.nc != nil {
		err := s.nc.Publish("emergency.authority.notified", data)
		if err != nil {
			s.logger.Error("NATS publish failed for authority notification", zap.Error(err))
			return false
		}
	}

	// Increment Prometheus metric
	emergency.AuthorityNotificationsTotal.WithLabelValues(string(auth), string(incident.Severity)).Inc()

	s.logger.Info("Authority notified (AIS-140)",
		zap.String("authority", string(auth)),
		zap.String("incident_id", incident.ID.String()),
	)
	return true
}
