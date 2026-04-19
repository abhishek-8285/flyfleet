package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/fleet-platform/shared/pkg/models"
	"github.com/flyfleet/fleet-core/internal/domains/tracking"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/processor"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// GPSPayload from device
type GPSPayload struct {
	DeviceID   string          `json:"device_id" validate:"required"`
	Timestamp  int64           `json:"timestamp" validate:"required"`
	Latitude   float64         `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude  float64         `json:"longitude" validate:"required,min=-180,max=180"`
	Altitude   *float64        `json:"altitude"`
	Speed      float64         `json:"speed"`
	Heading    float64         `json:"heading"`
	Accuracy   float64         `json:"accuracy"`
	Ignition   bool            `json:"ignition"`
	Odometer   *float64        `json:"odometer"`
	Battery    *int            `json:"battery_level"`
	Satellites *int            `json:"satellites"`
	HDOP       *float64        `json:"hdop"`
	RawData    json.RawMessage `json:"raw_data,omitempty"`
}

type TelemetryHandler struct {
	db          database.Database
	mb          messaging.MessageBroker
	redis       *redis.Client
	logger      *zap.Logger
	validator   *validator.Validate
	binaryCodec *processor.BinaryTelemetryCodec
	metrics     *trackingv2.Metrics
}

func NewTelemetryHandler(db database.Database, mb messaging.MessageBroker, rc *redis.Client, logger *zap.Logger, metrics *trackingv2.Metrics) *TelemetryHandler {
	return &TelemetryHandler{
		db:          db,
		mb:          mb,
		redis:       rc,
		logger:      logger,
		validator:   validator.New(),
		binaryCodec: processor.NewBinaryTelemetryCodec(),
		metrics:     metrics,
	}
}

func (h *TelemetryHandler) RegisterRoutes(r chi.Router) {
	r.Route("/device/telemetry", func(r chi.Router) {
		r.Post("/gps", h.IngestGPS)
		r.Post("/obd", h.IngestOBD)
		r.Post("/binary", h.IngestBinary)
	})
	r.Get("/vehicles/{id}/history", h.GetVehicleHistory)

	// Legacy endpoints for backward compatibility
	r.Post("/telemetry/gps", h.IngestGPS)
	r.Post("/telemetry/obd", h.IngestOBD)
}

// Helper methods for JSON responses
func (h *TelemetryHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", zap.Error(err))
	}
}

func (h *TelemetryHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

// IngestGPS - High-throughput GPS ingestion endpoint (JetStream)
// Legacy in-process storage removed. Now persistent via NATS JetStream.
func (h *TelemetryHandler) IngestGPS(w http.ResponseWriter, r *http.Request) {
	// SECURITY FIX: Limit request body size to prevent OOM attacks
	const maxBodySize = 10 * 1024 // 10KB max for GPS payload
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var payload GPSPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(payload); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	var jti string
	var orgIDStr string
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		// SECURITY FIX: Use proper JWT signature validation instead of ParseUnverified
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			// Reject tokens if no secret configured - don't fall back to unverified parsing
			h.logger.Error("JWT_SECRET not configured, rejecting authentication")
			h.errorJSON(w, "Server authentication misconfiguration", http.StatusInternalServerError)
			return
		}
		token, err := new(jwt.Parser).Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if jtiClaim, ok := claims["jti"].(string); ok {
					jti = jtiClaim
				}
				if orgClaim, ok := claims["tenant_id"].(string); ok {
					orgIDStr = orgClaim
				}
			}
		} else {
			h.logger.Warn("Invalid JWT token", zap.Error(err))
			h.errorJSON(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}
	} else if authHeader != "" {
		h.errorJSON(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	if h.redis != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Millisecond)
		defer cancel()

		// 1. Revocation Check (Zombie Kill Switch)
		isRevoked, err := h.redis.SIsMember(ctx, "auth:revoked_devices", payload.DeviceID).Result()
		if err == nil && isRevoked {
			if h.metrics != nil {
				h.metrics.TelemetryDroppedTotal.WithLabelValues("blacklisted", payload.DeviceID).Inc()
			}
			h.errorJSON(w, "Session Revoked - Device Blacklisted", http.StatusUnauthorized)
			return
		}

		sessionID := r.Header.Get("X-Session-ID")
		if sessionID != "" {
			sessionRevoked, err := h.redis.Exists(ctx, "auth:revoked_session:"+sessionID).Result()
			if err == nil && sessionRevoked > 0 {
				if h.metrics != nil {
					h.metrics.TelemetryDroppedTotal.WithLabelValues("session_revoked", payload.DeviceID).Inc()
				}
				h.errorJSON(w, "Session Revoked - Token Blacklisted", http.StatusUnauthorized)
				return
			}
		}

		if jti != "" {
			jtiRevoked, err := h.redis.Exists(ctx, "blacklist:jti:"+jti).Result()
			if err == nil && jtiRevoked > 0 {
				if h.metrics != nil {
					h.metrics.TelemetryZombiePingsTotal.WithLabelValues(orgIDStr).Inc()
				}
				h.logger.Warn("Blocked Zombie Ping", zap.String("jti", jti), zap.String("device_id", payload.DeviceID))
				h.errorJSON(w, "Session Revoked - JTI Blacklisted", http.StatusUnauthorized)
				return
			}
		}

		// 2. Deduplication Check (Zero-Defect Scaling)
		dedupeKey := fmt.Sprintf("dedupe:telemetry:%s:%d", payload.DeviceID, payload.Timestamp)
		// nx: true means only set if it does not exist
		added, err := h.redis.SetNX(ctx, dedupeKey, "1", 5*time.Minute).Result()
		if err != nil {
			h.logger.Warn("Redis dedupe check failed, continuing without deduplication", zap.Error(err))
		} else if !added {
			h.logger.Debug("Duplicate telemetry detected", zap.String("device_id", payload.DeviceID))
			h.json(w, map[string]interface{}{"status": "duplicate_suppressed"}, http.StatusAccepted)
			return
		}
	}

	var vehicleID, orgID uuid.UUID

	// Fast Path: Redis Cache
	if h.redis != nil {
		cacheKey := "device:resolve:" + payload.DeviceID
		cachedVal, err := h.redis.Get(r.Context(), cacheKey).Result()
		if err == nil {
			parts := strings.Split(cachedVal, "|")
			if len(parts) == 2 {
				vehicleID, _ = uuid.Parse(parts[0])
				orgID, _ = uuid.Parse(parts[1])
			}
		}
	}

	if vehicleID == uuid.Nil {
		row := h.db.QueryRow(r.Context(),
			`SELECT vehicle_id, tenant_id FROM devices WHERE device_id = $1 AND status = 'active'`,
			payload.DeviceID,
		)
		err := row.Scan(&vehicleID, &orgID)

		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				h.logger.Warn("Device not found or inactive", zap.String("device_id", payload.DeviceID))
				h.errorJSON(w, "Device not registered", http.StatusNotFound)
			} else {
				h.logger.Error("Database error during device resolution", zap.String("device_id", payload.DeviceID), zap.Error(err))
				h.errorJSON(w, "Service Unavailable: Database exhausted", http.StatusServiceUnavailable)
			}
			return
		}

		if h.redis != nil {
			cacheKey := "device:resolve:" + payload.DeviceID
			h.redis.Set(r.Context(), cacheKey, vehicleID.String()+"|"+orgID.String(), 15*time.Minute)
		}
	}

	timestamp := time.Unix(payload.Timestamp, 0)

	// Compliance check: Insurance status
	complianceViolation := false
	if h.redis != nil {
		complianceKey := fmt.Sprintf("compliance:expired_insurance:vehicle:%s", vehicleID.String())
		if exists, _ := h.redis.Exists(r.Context(), complianceKey).Result(); exists > 0 {
			complianceViolation = true
		}
	}

	// Pillar 4: Persistent Ingestion via JetStream
	// Storage handled by dedicated ingestion-service
	eventData := struct {
		Type      string              `json:"type"`
		Timestamp string              `json:"timestamp"`
		OrgID     string              `json:"tenant_id"`
		VehicleID string              `json:"vehicle_id"`
		Data      models.GPSTelemetry `json:"data"`
	}{
		Type:      "gps",
		Timestamp: timestamp.Format(time.RFC3339),
		OrgID:     orgID.String(),
		VehicleID: vehicleID.String(),
		Data: models.GPSTelemetry{
			Time:                timestamp,
			VehicleID:           vehicleID,
			OrgID:               orgID,
			Latitude:            payload.Latitude,
			Longitude:           payload.Longitude,
			Altitude:            payload.Altitude,
			Speed:               &payload.Speed,
			Heading:             &payload.Heading,
			Accuracy:            &payload.Accuracy,
			Ignition:            &payload.Ignition,
			Odometer:            payload.Odometer,
			Battery:             payload.Battery,
			Satellites:          payload.Satellites,
			ComplianceViolation: complianceViolation,
			RawData:             nil, // Map json.RawMessage if needed
		},
	}

	if h.mb != nil {
		if err := h.mb.PublishJSWithContext(r.Context(), "telemetry.gps", eventData); err != nil {
			h.logger.Error("Failed to ingest GPS data to JetStream", zap.Error(err))
		}
	} else {
		h.logger.Warn("Skipping GPS publish: message broker not configured")
	}

	// Immediate response to device
	h.json(w, map[string]interface{}{
		"status":     "accepted",
		"vehicle_id": vehicleID.String(),
		"timestamp":  timestamp.Unix(),
	}, http.StatusAccepted)
}

// IngestOBD - OBD-II data ingestion (JetStream)
func (h *TelemetryHandler) IngestOBD(w http.ResponseWriter, r *http.Request) {
	// SECURITY FIX: Limit request body size to prevent OOM attacks
	const maxBodySize = 10 * 1024 // 10KB max for OBD payload
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var payload struct {
		DeviceID  string          `json:"device_id" validate:"required"`
		Timestamp int64           `json:"timestamp" validate:"required"`
		EngineRPM *int            `json:"engine_rpm"`
		FuelLevel *int            `json:"fuel_level"`
		RawData   json.RawMessage `json:"raw_data,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	var jti string
	var orgIDStr string
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		// SECURITY FIX: Use proper JWT signature validation instead of ParseUnverified
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			// Reject tokens if no secret configured - don't fall back to unverified parsing
			h.logger.Error("JWT_SECRET not configured, rejecting authentication")
			h.errorJSON(w, "Server authentication misconfiguration", http.StatusInternalServerError)
			return
		}
		token, err := new(jwt.Parser).Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if jtiClaim, ok := claims["jti"].(string); ok {
					jti = jtiClaim
				}
				if orgClaim, ok := claims["tenant_id"].(string); ok {
					orgIDStr = orgClaim
				}
			}
		} else {
			h.logger.Warn("Invalid JWT token", zap.Error(err))
			h.errorJSON(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}
	} else if authHeader != "" {
		h.errorJSON(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	if h.redis != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Millisecond)
		defer cancel()
		isRevoked, err := h.redis.SIsMember(ctx, "auth:revoked_devices", payload.DeviceID).Result()
		if err == nil && isRevoked {
			if h.metrics != nil {
				h.metrics.TelemetryDroppedTotal.WithLabelValues("blacklisted", payload.DeviceID).Inc()
			}
			h.errorJSON(w, "Session Revoked - Device Blacklisted", http.StatusUnauthorized)
			return
		}

		if jti != "" {
			jtiRevoked, err := h.redis.Exists(ctx, "blacklist:jti:"+jti).Result()
			if err == nil && jtiRevoked > 0 {
				if h.metrics != nil {
					h.metrics.TelemetryZombiePingsTotal.WithLabelValues(orgIDStr).Inc()
				}
				h.logger.Warn("Blocked Zombie Ping", zap.String("jti", jti), zap.String("device_id", payload.DeviceID))
				h.errorJSON(w, "Session Revoked - JTI Blacklisted", http.StatusUnauthorized)
				return
			}
		}
	}

	var vehicleID, orgID uuid.UUID
	row := h.db.QueryRow(r.Context(),
		`SELECT vehicle_id, tenant_id FROM devices WHERE device_id = $1 AND status = 'active'`,
		payload.DeviceID,
	)
	err := row.Scan(&vehicleID, &orgID)
	if err != nil {
		h.errorJSON(w, "Device not registered", http.StatusNotFound)
		return
	}

	timestamp := time.Unix(payload.Timestamp, 0)

	// Pillar 4: Persistent Ingestion via JetStream
	eventData := struct {
		Type      string              `json:"type"`
		Timestamp string              `json:"timestamp"`
		OrgID     string              `json:"tenant_id"`
		VehicleID string              `json:"vehicle_id"`
		Data      models.OBDTelemetry `json:"data"`
	}{
		Type:      "obd",
		Timestamp: timestamp.Format(time.RFC3339),
		OrgID:     orgID.String(),
		VehicleID: vehicleID.String(),
		Data: models.OBDTelemetry{
			Time:      timestamp,
			VehicleID: vehicleID,
			OrgID:     orgID,
			RPM:       payload.EngineRPM,
			FuelLevel: nil, // Map correctly from payload if needed
		},
	}

	if h.mb != nil {
		if err := h.mb.PublishJSWithContext(r.Context(), "telemetry.obd", eventData); err != nil {
			h.logger.Error("Failed to ingest OBD data to JetStream", zap.Error(err))
		}
	} else {
		h.logger.Warn("Skipping OBD publish: message broker not configured")
	}

	h.json(w, map[string]string{"status": "accepted"}, http.StatusAccepted)
}

// IngestBinary - High-performance bit-packed binary ingestion (< 50 bytes)
func (h *TelemetryHandler) IngestBinary(w http.ResponseWriter, r *http.Request) {
	// 1. Read binary payload
	data := make([]byte, 25) // Fixed size for GPS packet
	_, err := r.Body.Read(data)
	if err != nil {
		h.errorJSON(w, "Incomplete binary payload", http.StatusBadRequest)
		return
	}

	// 2. Decode bit-packed data
	telemetry, deviceID, err := h.binaryCodec.DecodeGPS(data)
	if err != nil {
		h.logger.Error("Binary decode failed", zap.Error(err))
		h.errorJSON(w, "Protocol Error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var jti string
	var orgIDStr string
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		// SECURITY FIX: Use proper JWT signature validation instead of ParseUnverified
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			// Reject tokens if no secret configured - don't fall back to unverified parsing
			h.logger.Error("JWT_SECRET not configured, rejecting authentication")
			h.errorJSON(w, "Server authentication misconfiguration", http.StatusInternalServerError)
			return
		}
		token, err := new(jwt.Parser).Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if jtiClaim, ok := claims["jti"].(string); ok {
					jti = jtiClaim
				}
				if orgClaim, ok := claims["tenant_id"].(string); ok {
					orgIDStr = orgClaim
				}
			}
		} else {
			h.logger.Warn("Invalid JWT token", zap.Error(err))
			h.errorJSON(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}
	} else if authHeader != "" {
		h.errorJSON(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	if h.redis != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Millisecond)
		defer cancel()
		isRevoked, resErr := h.redis.SIsMember(ctx, "auth:revoked_devices", deviceID).Result()
		if resErr == nil && isRevoked {
			if h.metrics != nil {
				h.metrics.TelemetryDroppedTotal.WithLabelValues("blacklisted", deviceID).Inc()
			}
			h.errorJSON(w, "Session Revoked - Device Blacklisted", http.StatusUnauthorized)
			return
		}

		if jti != "" {
			jtiRevoked, err := h.redis.Exists(ctx, "blacklist:jti:"+jti).Result()
			if err == nil && jtiRevoked > 0 {
				if h.metrics != nil {
					h.metrics.TelemetryZombiePingsTotal.WithLabelValues(orgIDStr).Inc()
				}
				h.logger.Warn("Blocked Zombie Ping", zap.String("jti", jti), zap.String("device_id", deviceID))
				h.errorJSON(w, "Session Revoked - JTI Blacklisted", http.StatusUnauthorized)
				return
			}
		}
	}

	// 3. Resolve Device and Vechicle (Cached path would be better but keeping simple for MVP)
	var vehicleID, orgID uuid.UUID
	row := h.db.QueryRow(r.Context(),
		`SELECT vehicle_id, tenant_id FROM devices WHERE device_id = $1 AND status = 'active'`,
		deviceID,
	)
	err = row.Scan(&vehicleID, &orgID)
	if err != nil {
		h.errorJSON(w, "Device not registered", http.StatusNotFound)
		return
	}

	// 4. Enrich and Publish to JetStream
	eventData := struct {
		Type      string              `json:"type"`
		Timestamp string              `json:"timestamp"`
		OrgID     string              `json:"tenant_id"`
		VehicleID string              `json:"vehicle_id"`
		Data      models.GPSTelemetry `json:"data"`
	}{
		Type:      "gps_binary",
		Timestamp: telemetry.Time.Format(time.RFC3339),
		OrgID:     orgID.String(),
		VehicleID: vehicleID.String(),
		Data:      *telemetry,
	}

	if h.mb != nil {
		h.mb.PublishJSWithContext(r.Context(), "telemetry.gps", eventData)
	}

	h.json(w, map[string]string{"status": "accepted_binary"}, http.StatusAccepted)
}

func (h *TelemetryHandler) GetVehicleHistory(w http.ResponseWriter, r *http.Request) {
	vehicleID := chi.URLParam(r, "id")

	rows, err := h.db.Query(r.Context(), `
        SELECT time, ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng, speed
        FROM gps_locations
        WHERE vehicle_id = $1
        ORDER BY time ASC LIMIT 100`,
		vehicleID,
	)

	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, rows, http.StatusOK)
}
