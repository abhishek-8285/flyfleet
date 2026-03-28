package video

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// cameraHeartbeatThreshold: cameras not heard from in this duration are considered offline.
	cameraHeartbeatThreshold = 5 * time.Minute
	// tamperAlertCooldown: minimum time between tamper alerts for the same camera.
	tamperAlertCooldown = 15 * time.Minute
)

// DeviceHealthMonitor tracks dashcam heartbeats, firmware versions, and
// generates tamper/offline alerts via structured logging.
// In production, extend to publish to the alerts module via gRPC.
type DeviceHealthMonitor struct {
	repo       Repository
	logger     *zap.Logger
	lastAlerts map[uuid.UUID]time.Time
	mu         sync.RWMutex // SECURITY FIX: Protect concurrent map access
}

func NewDeviceHealthMonitor(repo Repository, logger *zap.Logger) *DeviceHealthMonitor {
	return &DeviceHealthMonitor{
		repo:       repo,
		logger:     logger,
		lastAlerts: make(map[uuid.UUID]time.Time),
	}
}

// Run starts the health monitor loop. Checks every minute.
func (m *DeviceHealthMonitor) Run(ctx context.Context) {
	m.logger.Info("Device health monitor started")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Device health monitor stopped")
			return
		case <-ticker.C:
			m.checkAll(ctx)
		}
	}
}

// RecordHeartbeat updates the camera's last_heartbeat timestamp.
func (m *DeviceHealthMonitor) RecordHeartbeat(ctx context.Context, cameraID uuid.UUID) error {
	return m.repo.TouchCameraHeartbeat(ctx, cameraID)
}

func (m *DeviceHealthMonitor) checkAll(ctx context.Context) {
	// In production: query all active cameras and check heartbeat times.
	// Stub: structure is in place; real implementation queries cameras table.
	m.logger.Debug("Device health check cycle complete")
}

// CheckCamera evaluates a specific camera and emits an alert if stale/offline.
func (m *DeviceHealthMonitor) CheckCamera(ctx context.Context, c *Camera, expectedFirmware string) {
	// 1. Check heartbeat
	if IsCameraStale(c.LastHeartbeat, cameraHeartbeatThreshold) {
		m.emitAlert(c.ID, EventCameraOffline, 4,
			"Camera offline: no heartbeat received",
			zap.String("camera_id", c.ID.String()),
			zap.String("vehicle_id", c.VehicleID.String()),
			zap.String("position", string(c.Position)))
		// Persist offline event
		event := &VideoEvent{
			ID:              uuid.New(),
			TenantID:        c.TenantID,
			VehicleID:       c.VehicleID,
			EventType:       EventCameraOffline,
			OccurredAt:      time.Now(),
			Severity:        4,
			ConfidenceScore: 1.0,
			Metadata: map[string]interface{}{
				"camera_position": string(c.Position),
				"serial_number":   c.SerialNumber,
			},
		}
		m.repo.CreateVideoEvent(ctx, event)
	}

	// 2. Check firmware version mismatch
	if expectedFirmware != "" && c.FirmwareVersion != expectedFirmware {
		m.logger.Warn("Camera firmware out of date",
			zap.String("camera_id", c.ID.String()),
			zap.String("current", c.FirmwareVersion),
			zap.String("expected", expectedFirmware))
	}
}

// ReportTamper is called when an edge device detects tampering (e.g., lens covered, SD removed).
func (m *DeviceHealthMonitor) ReportTamper(ctx context.Context, cameraID, vehicleID, tenantID uuid.UUID, detail string) {
	m.emitAlert(cameraID, EventCameraTamper, 5,
		"Camera tamper detected: "+detail,
		zap.String("camera_id", cameraID.String()),
		zap.String("vehicle_id", vehicleID.String()))

	event := &VideoEvent{
		ID:              uuid.New(),
		TenantID:        tenantID,
		VehicleID:       vehicleID,
		EventType:       EventCameraTamper,
		OccurredAt:      time.Now(),
		Severity:        5,
		ConfidenceScore: 1.0,
		Metadata:        map[string]interface{}{"detail": detail},
	}
	m.repo.CreateVideoEvent(ctx, event)
}

func (m *DeviceHealthMonitor) emitAlert(cameraID uuid.UUID, et EventType, severity int, msg string, fields ...zap.Field) {
	// Cooldown: don't alert too frequently for the same camera
	// SECURITY FIX: Use RWMutex to protect concurrent map access
	m.mu.Lock()
	last, ok := m.lastAlerts[cameraID]
	if ok && time.Since(last) < tamperAlertCooldown {
		m.mu.Unlock()
		return
	}
	m.lastAlerts[cameraID] = time.Now()
	m.mu.Unlock()

	m.logger.Warn(msg, append(fields,
		zap.String("event_type", string(et)),
		zap.Int("severity", severity))...)
}
