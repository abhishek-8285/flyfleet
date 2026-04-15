# 📡 Device Lifecycle & Firmware Integrity Analysis

**Author**: Lead System Architect  
**Date**: March 15, 2026  
**Scope**: Version drift, digital twin shadow state, OTA failure recovery

---

## Executive Summary

This document provides a comprehensive analysis of FlyFleet's device lifecycle management and firmware integrity strategy, covering:

1. **Version Drift**: Backward compatibility for 10,000 legacy devices running Firmware v1.0
2. **Shadow State Logic**: Digital Twin implementation for desired vs reported state reconciliation
3. **OTA Failure**: Safe-mode rollback logic to prevent bricking devices

---

## 1️⃣ The Version Drift: Backward Compatibility Strategy

### Current Device Fleet Composition

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DEVICE FLEET COMPOSITION                             │
└─────────────────────────────────────────────────────────────────────────┘

Firmware Version Distribution:
┌──────────────────────────────────────────────────────────────────────┐
│ v1.0 (Legacy)     ████████████████████████████████████  10,000 devices │
│ v1.5              ████████████████                       4,000 devices │
│ v2.0              ██████████                             2,500 devices │
│ v2.1 (Current)    ███████                                1,800 devices │
│ v2.2 (Beta)       ██                                       500 devices │
└──────────────────────────────────────────────────────────────────────┘
Total: 18,800 devices across 5 firmware versions
```

### Schema Evolution Challenge

**Scenario**: `tracking_v2` domain updates ingestion schema from v1 → v2

```go
// OLD SCHEMA (v1.0 devices send this)
type GPSTelemetryV1 struct {
    DeviceID   string  `json:"device_id"`
    Timestamp  int64   `json:"timestamp"`  // Unix seconds
    Latitude   float64 `json:"latitude"`
    Longitude  float64 `json:"longitude"`
    Speed      float64 `json:"speed"`
}

// NEW SCHEMA (v2.0+ devices send this)
type GPSTelemetryV2 struct {
    DeviceID   string     `json:"device_id"`
    Timestamp  time.Time  `json:"timestamp"`  // RFC3339 format
    Latitude   float64    `json:"latitude"`
    Longitude  float64    `json:"longitude"`
    Speed      float64    `json:"speed"`
    Altitude   *float64   `json:"altitude"`   // NEW field
    Heading    *float64   `json:"heading"`    // NEW field
    Satellites *int       `json:"satellites"` // NEW field
    HDOP       *float64   `json:"hdop"`       // NEW field
}
```

### ⚠️ CRITICAL GAP: No Explicit Version Negotiation

**Current behavior** (from `telemetry_handler.go`):

```go
// CURRENT IMPLEMENTATION (No version handling)
func (h *TelemetryHandler) IngestGPS(w http.ResponseWriter, r *http.Request) {
    var payload GPSPayload  // Single schema, no version negotiation
    
    if err := json.Unmarshal(msg.Data, &payload); err != nil {
        h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
        return  // ← Legacy v1.0 devices get 400 error
    }
    
    // Process with v2 schema...
}
```

**Problem**: When schema changes:
1. Legacy v1.0 devices send old format
2. Monolith expects new format
3. JSON unmarshal fails → 400 Bad Request
4. **10,000 devices suddenly broken**

### Recommended: Multi-Version Ingestion with Adapter Pattern

```go
// internal/domains/tracking_v2/handlers/telemetry_handler.go (ENHANCED)
package handlers

// TelemetryVersion identifies the schema version
type TelemetryVersion string

const (
    VersionV1 TelemetryVersion = "v1"
    VersionV2 TelemetryVersion = "v2"
    VersionV3 TelemetryVersion = "v3"
)

// GPSPayloadV1 - Legacy schema (10,000 devices)
type GPSPayloadV1 struct {
    DeviceID  string  `json:"device_id" validate:"required"`
    Timestamp int64   `json:"timestamp" validate:"required"`
    Latitude  float64 `json:"latitude" validate:"required"`
    Longitude float64 `json:"longitude" validate:"required"`
    Speed     float64 `json:"speed"`
}

// GPSPayloadV2 - Current schema (6,300 devices)
type GPSPayloadV2 struct {
    DeviceID   string     `json:"device_id" validate:"required"`
    Timestamp  time.Time  `json:"timestamp" validate:"required"`
    Latitude   float64    `json:"latitude" validate:"required"`
    Longitude  float64    `json:"longitude" validate:"required"`
    Speed      float64    `json:"speed"`
    Altitude   *float64   `json:"altitude"`
    Heading    *float64   `json:"heading"`
    Satellites *int       `json:"satellites"`
}

// GPSPayloadV3 - Next-gen schema (500 devices)
type GPSPayloadV3 struct {
    GPSPayloadV2           // Embed v2 for backward compat
    HDOP          *float64 `json:"hdop"`
    AccuracyMeters *float64 `json:"accuracy_meters"`
    Source        string   `json:"source"` // gps, glonass, galileo
}

// TelemetryAdapter converts between versions
type TelemetryAdapter struct {
    logger *zap.Logger
}

// NormalizeToV3 converts any version to canonical v3 format
func (a *TelemetryAdapter) NormalizeToV3(data []byte, version TelemetryVersion) (*GPSPayloadV3, error) {
    switch version {
    case VersionV1:
        var v1 GPSPayloadV1
        if err := json.Unmarshal(data, &v1); err != nil {
            return nil, err
        }
        
        // ADAPTER: Convert v1 → v3
        return &GPSPayloadV3{
            GPSPayloadV2: GPSPayloadV2{
                DeviceID:  v1.DeviceID,
                Timestamp: time.Unix(v1.Timestamp, 0), // Convert unix → RFC3339
                Latitude:  v1.Latitude,
                Longitude: v1.Longitude,
                Speed:     v1.Speed,
                // v1 doesn't have these - set defaults
                Altitude:   nil,
                Heading:    nil,
                Satellites: nil,
            },
            HDOP:           nil,
            AccuracyMeters: nil,
            Source:         "gps",
        }, nil
        
    case VersionV2:
        var v2 GPSPayloadV2
        if err := json.Unmarshal(data, &v2); err != nil {
            return nil, err
        }
        
        // ADAPTER: Convert v2 → v3
        return &GPSPayloadV3{
            GPSPayloadV2:   v2,
            HDOP:           nil,
            AccuracyMeters: nil,
            Source:         "gps",
        }, nil
        
    case VersionV3:
        var v3 GPSPayloadV3
        if err := json.Unmarshal(data, &v3); err != nil {
            return nil, err
        }
        return &v3, nil
        
    default:
        return nil, fmt.Errorf("unsupported telemetry version: %s", version)
    }
}

// DetectVersion auto-detects schema version from payload
func (a *TelemetryAdapter) DetectVersion(data []byte) TelemetryVersion {
    // Try to unmarshal into v3 first (most fields)
    var v3 GPSPayloadV3
    if err := json.Unmarshal(data, &v3); err == nil {
        if v3.HDOP != nil || v3.AccuracyMeters != nil {
            return VersionV3
        }
        if v3.Altitude != nil || v3.Heading != nil {
            return VersionV2
        }
    }
    
    // Try v2
    var v2 GPSPayloadV2
    if err := json.Unmarshal(data, &v2); err == nil {
        return VersionV2
    }
    
    // Fallback to v1 (legacy)
    return VersionV1
}

// Enhanced handler with version negotiation
func (h *TelemetryHandler) IngestGPS(w http.ResponseWriter, r *http.Request) {
    var rawPayload json.RawMessage
    if err := json.NewDecoder(r.Body).Decode(&rawPayload); err != nil {
        h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
        return
    }
    
    // Auto-detect version from payload structure
    adapter := NewTelemetryAdapter(h.logger)
    version := adapter.DetectVersion(rawPayload)
    
    h.logger.Debug("Detected telemetry version",
        zap.String("version", string(version)),
        zap.String("device_id", extractDeviceID(rawPayload)))
    
    // Normalize to canonical v3 format
    normalized, err := adapter.NormalizeToV3(rawPayload, version)
    if err != nil {
        h.errorJSON(w, fmt.Sprintf("Invalid payload for version %s", version), http.StatusBadRequest)
        return
    }
    
    // Process normalized payload (all business logic uses v3)
    h.processNormalizedGPS(r.Context(), normalized, version)
    
    // Track version distribution for analytics
    h.metrics.RecordTelemetryVersion(version)
}

func (h *TelemetryHandler) processNormalizedGPS(ctx context.Context, payload *GPSPayloadV3, originalVersion TelemetryVersion) {
    // All business logic now works with normalized v3 format
    // regardless of what version the device sent
    
    // Resolve device to vehicle + org
    var vehicleID, orgID uuid.UUID
    row := h.db.QueryRow(ctx,
        `SELECT vehicle_id, org_id FROM devices 
         WHERE device_id = $1 AND status = 'active'`,
        payload.DeviceID,
    )
    row.Scan(&vehicleID, &orgID)
    
    // Store in database (schema supports all v3 fields)
    eventData := struct {
        Type      string              `json:"type"`
        Timestamp string              `json:"timestamp"`
        OrgID     string              `json:"org_id"`
        VehicleID string              `json:"vehicle_id"`
        Data      models.GPSTelemetry `json:"data"`
    }{
        Type:      "gps",
        Timestamp: payload.Timestamp.Format(time.RFC3339),
        OrgID:     orgID.String(),
        VehicleID: vehicleID.String(),
        Data: models.GPSTelemetry{
            Time:       payload.Timestamp,
            VehicleID:  vehicleID,
            OrgID:      orgID,
            Latitude:   payload.Latitude,
            Longitude:  payload.Longitude,
            Altitude:   payload.Altitude,
            Speed:      &payload.Speed,
            Heading:    payload.Heading,
            Satellites: payload.Satellites,
        },
    }
    
    // Publish to NATS for ingestion
    h.mb.PublishJS("telemetry.gps", eventData)
    
    // Respond with version-specific acknowledgment
    h.json(w, map[string]interface{}{
        "status":      "accepted",
        "version":     string(originalVersion),
        "normalized":  true,
        "vehicle_id":  vehicleID.String(),
        "timestamp":   payload.Timestamp.Unix(),
    })
}
```

### Version Negotiation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    VERSION NEGOTIATION FLOW                             │
└─────────────────────────────────────────────────────────────────────────┘

Device (v1.0) sends:
  {"device_id": "DEV001", "timestamp": 1710518400, "latitude": 28.6139, "longitude": 77.2090, "speed": 45}
                    ↑ Unix timestamp, no altitude/heading

Monolith receives:
  1. DetectVersion() → "v1" (no altitude/heading fields)
  2. NormalizeToV3() → Convert to canonical format
     - timestamp: 1710518400 → "2024-03-15T12:00:00Z" (RFC3339)
     - altitude: null (v1 doesn't have)
     - heading: null (v1 doesn't have)
  3. Process with v3 business logic
  4. Store in database (all fields, nulls for missing)

Response to device:
  {
    "status": "accepted",
    "version": "v1",
    "normalized": true,
    "vehicle_id": "550e8400-e29b-41d4-a716-446655440000"
  }

Device (v2.2) sends:
  {"device_id": "DEV002", "timestamp": "2024-03-15T12:00:00Z", "latitude": 28.6139, "longitude": 77.2090, "speed": 60, "altitude": 220, "heading": 135, "satellites": 12, "hdop": 1.2}

Monolith receives:
  1. DetectVersion() → "v3" (has hdop field)
  2. NormalizeToV3() → Already v3, pass through
  3. Process with v3 business logic
  4. Store in database (all fields populated)
```

### Database Schema for Multi-Version Support

```sql
-- migrations/072_tracking_enterprise.sql (ENHANCED)
CREATE TABLE gps_locations (
    time TIMESTAMPTZ NOT NULL,
    vehicle_id UUID NOT NULL,
    org_id UUID REFERENCES organizations(id),
    
    -- Core fields (all versions)
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    speed DECIMAL(6,2),
    
    -- v2+ fields (nullable for v1 compat)
    altitude DECIMAL(8,2),
    heading DECIMAL(5,2),
    satellites INTEGER,
    
    -- v3 fields (nullable for v1/v2 compat)
    hdop DECIMAL(4,2),
    accuracy_meters DECIMAL(6,2),
    source VARCHAR(20) DEFAULT 'gps',
    
    -- Version tracking for analytics
    schema_version VARCHAR(10) DEFAULT 'v3',
    raw_payload JSONB,  -- Store original for debugging
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for version-based queries
CREATE INDEX idx_gps_version ON gps_locations(schema_version);
CREATE INDEX idx_gps_vehicle_time ON gps_locations(vehicle_id, time DESC);
```

### Deprecation Strategy for Legacy Devices

```go
// internal/domains/tracking_v2/service/device_lifecycle.go (NEW)
type DeviceLifecycleService struct {
    repo   DeviceRepository
    logger *zap.Logger
}

// DeprecationNotice sent to legacy devices
type DeprecationNotice struct {
    CurrentVersion     string    `json:"current_version"`
    MinimumSupported   string    `json:"minimum_supported"`
    SunsetDate         time.Time `json:"sunset_date"`
    UpdateRequired     bool      `json:"update_required"`
    UpdateInstructions string    `json:"update_instructions"`
}

// CheckAndNotifyLegacy checks if device needs update
func (s *DeviceLifecycleService) CheckAndNotifyLegacy(ctx context.Context, deviceID string) (*DeprecationNotice, error) {
    device, err := s.repo.GetByDeviceID(ctx, deviceID)
    if err != nil {
        return nil, err
    }
    
    // Parse firmware version
    currentVersion := parseVersion(device.FirmwareVersion)
    minimumSupported := parseVersion("v1.5")  // v1.0 deprecated after Q2 2026
    
    if currentVersion.LessThan(minimumSupported) {
        notice := &DeprecationNotice{
            CurrentVersion:   device.FirmwareVersion,
            MinimumSupported: "v1.5",
            SunsetDate:       time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
            UpdateRequired:   true,
            UpdateInstructions: "Contact support for OTA update or visit service center",
        }
        
        s.logger.Warn("Legacy device detected",
            zap.String("device_id", deviceID),
            zap.String("firmware", device.FirmwareVersion),
            zap.Time("sunset_date", notice.SunsetDate))
        
        return notice, nil
    }
    
    return nil, nil
}
```

### Version Support Timeline

| Quarter | v1.0 | v1.5 | v2.0 | v2.1 | v2.2 |
|---------|------|------|------|------|------|
| **Q1 2026** | ✅ Supported | ✅ Supported | ✅ Supported | ✅ Supported | ✅ Beta |
| **Q2 2026** | ⚠️ Deprecated | ✅ Supported | ✅ Supported | ✅ Supported | ✅ Supported |
| **Q3 2026** | ❌ Unsupported | ✅ Supported | ✅ Supported | ✅ Supported | ✅ Supported |
| **Q4 2026** | ❌ Unsupported | ⚠️ Deprecated | ✅ Supported | ✅ Supported | ✅ Supported |

---

## 2️⃣ Shadow State Logic: Digital Twin Implementation

### Current Digital Twin Architecture

**Location**: `internal/domains/diagnostics_v2/service/device_twin_service.go`

```go
// CURRENT IMPLEMENTATION (Partial)
type DeviceTwin struct {
    DeviceID          uuid.UUID              `json:"device_id"`
    OrgID             uuid.UUID              `json:"org_id"`
    DesiredState      map[string]interface{} `json:"desired_state,omitempty"`
    ReportedState     map[string]interface{} `json:"reported_state,omitempty"`
    LastSeen          *time.Time             `json:"last_seen,omitempty"`
    DriftStatus       string                 `json:"drift_status"`  // unknown, in_sync, drifted
    DriftReason       *string                `json:"drift_reason,omitempty"`
    LastCommandID     *uuid.UUID             `json:"last_command_id,omitempty"`
    LastCommandStatus *string                `json:"last_command_status,omitempty"`
}

// CURRENT: Unimplemented methods
func (s *DeviceTwinService) UpdateDesiredState(ctx context.Context, deviceID uuid.UUID, state json.RawMessage) error {
    return fmt.Errorf("unimplemented: requires extending repository.go")
}

func (s *DeviceTwinService) SyncReportedState(ctx context.Context, deviceID uuid.UUID, state json.RawMessage) error {
    return fmt.Errorf("unimplemented: requires extending repository.go")
}

func (s *DeviceTwinService) DetectConfigurationDrift(ctx context.Context, deviceID uuid.UUID) (bool, error) {
    return false, nil  // ← Always returns false (no drift detection)
}
```

### ⚠️ CRITICAL GAP: Digital Twin Logic Incomplete

**Problem**: 
- `UpdateDesiredState()` → Unimplemented
- `SyncReportedState()` → Unimplemented  
- `DetectConfigurationDrift()` → Always returns false

**Impact**: No actual shadow state reconciliation happening.

### Recommended: Complete Digital Twin Implementation

```go
// internal/domains/diagnostics_v2/service/device_twin_service.go (COMPLETE)
package service

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/flyfleet/fleet-core/internal/domains/diagnostics_v2/repository"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

type DeviceTwinService struct {
    repo       *repository.DiagnosticsRepository
    nats       messaging.MessageBroker
    logger     *zap.Logger
}

// DesiredState represents configuration commands from backend
type DesiredState struct {
    Version       int                    `json:"version"`
    ConfigVersion string                 `json:"config_version"`
    Parameters    map[string]interface{} `json:"parameters"`
    CommandID     uuid.UUID              `json:"command_id"`
    IssuedAt      time.Time              `json:"issued_at"`
    ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
}

// ReportedState represents actual device state
type ReportedState struct {
    Version       int                    `json:"version"`
    ConfigVersion string                 `json:"config_version"`
    Parameters    map[string]interface{} `json:"parameters"`
    LastSeen      time.Time              `json:"last_seen"`
    Firmware      string                 `json:"firmware"`
    BatteryLevel  *int                   `json:"battery_level,omitempty"`
    SignalStrength *int                  `json:"signal_strength,omitempty"`
}

// DriftDetection represents the result of comparing desired vs reported
type DriftDetection struct {
    IsDrifted     bool              `json:"is_drifted"`
    DriftedFields []string          `json:"drifted_fields"`
    DriftSeverity string            `json:"drift_severity"`  // minor, major, critical
    Details       map[string]DriftDetail `json:"details"`
}

type DriftDetail struct {
    Field       string      `json:"field"`
    Desired     interface{} `json:"desired"`
    Reported    interface{} `json:"reported"`
    IsCritical  bool        `json:"is_critical"`
}

// UpdateDesiredState sets the target configuration for a device
func (s *DeviceTwinService) UpdateDesiredState(
    ctx context.Context,
    deviceID uuid.UUID,
    orgID uuid.UUID,
    params map[string]interface{},
    commandID *uuid.UUID,
) error {
    
    // 1. Get current twin
    twin, err := s.repo.GetDeviceTwin(ctx, deviceID, orgID)
    if err != nil {
        return fmt.Errorf("failed to get device twin: %w", err)
    }
    
    // 2. Build desired state
    desired := DesiredState{
        Version:       1,
        ConfigVersion: "v2.0",
        Parameters:    params,
        CommandID:     *commandID,
        IssuedAt:      time.Now(),
    }
    
    // 3. Set expiration for time-sensitive commands (e.g., speed limit)
    if params["expires_in_seconds"] != nil {
        if expiresIn, ok := params["expires_in_seconds"].(float64); ok {
            expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
            desired.ExpiresAt = &expiresAt
        }
    }
    
    // 4. Upsert desired state
    desiredJSON, _ := json.Marshal(desired)
    var desiredMap map[string]interface{}
    json.Unmarshal(desiredJSON, &desiredMap)
    
    err = s.repo.UpsertDesiredState(ctx, deviceID, orgID, desiredMap, commandID, nil)
    if err != nil {
        return fmt.Errorf("failed to upsert desired state: %w", err)
    }
    
    // 5. Publish to NATS for device sync
    syncEvent := struct {
        Type       string    `json:"type"`
        DeviceID   string    `json:"device_id"`
        Desired    DesiredState `json:"desired"`
        Timestamp  time.Time `json:"timestamp"`
    }{
        Type:      "twin.desired_updated",
        DeviceID:  deviceID.String(),
        Desired:   desired,
        Timestamp: time.Now(),
    }
    
    if s.nats != nil {
        s.nats.Publish("device.sync."+deviceID.String(), syncEvent)
    }
    
    s.logger.Info("Desired state updated",
        zap.String("device_id", deviceID.String()),
        zap.Any("params", params))
    
    return nil
}

// SyncReportedState updates the actual device state from heartbeat
func (s *DeviceTwinService) SyncReportedState(
    ctx context.Context,
    deviceID uuid.UUID,
    orgID uuid.UUID,
    params map[string]interface{},
    lastSeen *time.Time,
) error {
    
    // 1. Build reported state
    reported := ReportedState{
        Version:       1,
        ConfigVersion: extractConfigVersion(params),
        Parameters:    params,
        LastSeen:      *lastSeen,
        Firmware:      extractFirmwareVersion(params),
        BatteryLevel:  extractBattery(params),
        SignalStrength: extractSignal(params),
    }
    
    // 2. Upsert reported state
    reportedJSON, _ := json.Marshal(reported)
    var reportedMap map[string]interface{}
    json.Unmarshal(reportedJSON, &reportedMap)
    
    err := s.repo.UpsertReportedState(ctx, deviceID, orgID, reportedMap, lastSeen)
    if err != nil {
        return fmt.Errorf("failed to upsert reported state: %w", err)
    }
    
    // 3. Check for drift after updating reported state
    drift, err := s.DetectConfigurationDrift(ctx, deviceID, orgID)
    if err != nil {
        s.logger.Error("Failed to detect drift", zap.Error(err))
    }
    
    if drift.IsDrifted {
        s.logger.Warn("Configuration drift detected",
            zap.String("device_id", deviceID.String()),
            zap.Strings("drifted_fields", drift.DriftedFields),
            zap.String("severity", drift.DriftSeverity))
        
        // 4. Emit drift alert if critical
        if drift.DriftSeverity == "critical" {
            s.emitDriftAlert(ctx, deviceID, orgID, drift)
        }
    }
    
    return nil
}

// DetectConfigurationDrift compares desired vs reported state
func (s *DeviceTwinService) DetectConfigurationDrift(
    ctx context.Context,
    deviceID uuid.UUID,
    orgID uuid.UUID,
) (*DriftDetection, error) {
    
    // 1. Get twin
    twin, err := s.repo.GetDeviceTwin(ctx, deviceID, orgID)
    if err != nil {
        return nil, err
    }
    
    // 2. Handle missing states
    if len(twin.DesiredState) == 0 || len(twin.ReportedState) == 0 {
        return &DriftDetection{
            IsDrifted:     false,
            DriftedFields: nil,
            DriftSeverity: "unknown",
            Details:       nil,
        }, nil
    }
    
    // 3. Extract parameters for comparison
    desiredParams := extractParams(twin.DesiredState)
    reportedParams := extractParams(twin.ReportedState)
    
    // 4. Compare field by field
    drift := &DriftDetection{
        IsDrifted:     false,
        DriftedFields: []string{},
        DriftSeverity: "minor",
        Details:       make(map[string]DriftDetail),
    }
    
    criticalFields := map[string]bool{
        "speed_limit":        true,
        "geofence_enabled":   true,
        "emergency_contact":  true,
        "firmware_version":   true,
    }
    
    for field, desiredValue := range desiredParams {
        reportedValue, exists := reportedParams[field]
        
        if !exists {
            // Field missing in reported state
            drift.IsDrifted = true
            drift.DriftedFields = append(drift.DriftedFields, field)
            drift.Details[field] = DriftDetail{
                Field:      field,
                Desired:    desiredValue,
                Reported:   nil,
                IsCritical: criticalFields[field],
            }
            continue
        }
        
        // Compare values (handle JSON number vs float64)
        if !valuesEqual(desiredValue, reportedValue) {
            drift.IsDrifted = true
            drift.DriftedFields = append(drift.DriftedFields, field)
            drift.Details[field] = DriftDetail{
                Field:      field,
                Desired:    desiredValue,
                Reported:   reportedValue,
                IsCritical: criticalFields[field],
            }
        }
    }
    
    // 5. Determine severity
    hasCriticalDrift := false
    for _, detail := range drift.Details {
        if detail.IsCritical {
            hasCriticalDrift = true
            break
        }
    }
    
    if hasCriticalDrift {
        drift.DriftSeverity = "critical"
    } else if len(drift.DriftedFields) > 3 {
        drift.DriftSeverity = "major"
    } else if len(drift.DriftedFields) > 0 {
        drift.DriftSeverity = "minor"
    } else {
        drift.DriftSeverity = "none"
    }
    
    return drift, nil
}

// ReconcileOfflineDevice handles devices that have been offline
func (s *DeviceTwinService) ReconcileOfflineDevice(
    ctx context.Context,
    deviceID uuid.UUID,
    orgID uuid.UUID,
    offlineDuration time.Duration,
) error {
    
    twin, err := s.repo.GetDeviceTwin(ctx, deviceID, orgID)
    if err != nil {
        return err
    }
    
    // Check if desired state has expired
    if twin.DesiredState["expires_at"] != nil {
        expiresAt, ok := twin.DesiredState["expires_at"].(time.Time)
        if ok && time.Now().After(expiresAt) {
            // Desired state expired while device was offline
            s.logger.Info("Desired state expired during offline period",
                zap.String("device_id", deviceID.String()),
                zap.Duration("offline_duration", offlineDuration),
                zap.Time("expired_at", expiresAt))
            
            // Clear expired desired state
            s.repo.UpsertDesiredState(ctx, deviceID, orgID, map[string]interface{}{}, nil, nil)
            return nil
        }
    }
    
    // If offline > 48 hours, mark for resync
    if offlineDuration > 48*time.Hour {
        s.logger.Warn("Device offline > 48 hours, marking for resync",
            zap.String("device_id", deviceID.String()),
            zap.Duration("offline_duration", offlineDuration))
        
        // Queue resync command
        resyncCmd := &models.DeviceCommand{
            DeviceID: deviceID,
            OrgID:    orgID,
            Command:  "RESYNC_CONFIG",
            Category: "CONFIG",
            Params: map[string]interface{}{
                "reason": "offline_extended",
                "duration_hours": int(offlineDuration.Hours()),
            },
            Status: "queued",
        }
        
        s.repo.CreateCommand(ctx, resyncCmd)
    }
    
    return nil
}

// Helper functions
func extractParams(state map[string]interface{}) map[string]interface{} {
    if params, ok := state["parameters"].(map[string]interface{}); ok {
        return params
    }
    return state
}

func valuesEqual(a, b interface{}) bool {
    // Handle JSON number comparison (float64 vs int)
    if aFloat, ok := a.(float64); ok {
        if bFloat, ok := b.(float64); ok {
            return aFloat == bFloat
        }
        if bInt, ok := b.(int); ok {
            return aFloat == float64(bInt)
        }
    }
    
    // Handle string comparison
    if aStr, ok := a.(string); ok {
        if bStr, ok := b.(string); ok {
            return aStr == bStr
        }
    }
    
    // Handle boolean comparison
    if aBool, ok := a.(bool); ok {
        if bBool, ok := b.(bool); ok {
            return aBool == bBool
        }
    }
    
    // Fallback to JSON comparison
    aJSON, _ := json.Marshal(a)
    bJSON, _ := json.Marshal(b)
    return bytes.Equal(aJSON, bJSON)
}

func extractConfigVersion(params map[string]interface{}) string {
    if v, ok := params["config_version"].(string); ok {
        return v
    }
    return "unknown"
}

func extractFirmwareVersion(params map[string]interface{}) string {
    if v, ok := params["firmware"].(string); ok {
        return v
    }
    return "unknown"
}

func extractBattery(params map[string]interface{}) *int {
    if v, ok := params["battery_level"].(float64); ok {
        battery := int(v)
        return &battery
    }
    return nil
}

func extractSignal(params map[string]interface{}) *int {
    if v, ok := params["signal_strength"].(float64); ok {
        signal := int(v)
        return &signal
    }
    return nil
}
```

### 48-Hour Offline Reconciliation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    OFFLINE DEVICE RECONCILIATION                        │
└─────────────────────────────────────────────────────────────────────────┘

T0: Device goes offline
    → Last heartbeat: 2024-03-13 12:00:00
    → Desired state: {speed_limit: 60, geofence: enabled}
    → Reported state: {speed_limit: 60, geofence: enabled}
    → Drift status: in_sync

T1: Backend updates desired state (T0 + 24 hours)
    → New desired: {speed_limit: 50, geofence: enabled}
    → Device still offline, can't receive update
    → Twin: desired ≠ reported (drift exists, but device doesn't know)

T2: Device comes online (T0 + 48 hours)
    → Heartbeat received
    → Reported state: {speed_limit: 60, geofence: enabled} (OLD)
    
    Reconciliation logic:
    1. Check if desired state expired
       → expires_at: null (no expiration)
       → State still valid
    
    2. Detect drift
       → desired.speed_limit: 50
       → reported.speed_limit: 60
       → DRIFT DETECTED (critical field)
    
    3. Queue RESYNC_CONFIG command
       → Command: "Update speed_limit to 50"
       → Status: queued
       → Priority: high
    
    4. Mark twin for resync
       → drift_status: drifted
       → drift_reason: "device_offline_48h"

T3: Device receives command
    → Apply new config: speed_limit = 50
    → Send acknowledgment
    → Update reported state

T4: Backend receives ack
    → Update reported state: {speed_limit: 50, ...}
    → Re-check drift
    → drift_status: in_sync
```

### State Reconciliation Matrix

| Scenario | Desired State | Reported State | Action |
|----------|---------------|----------------|--------|
| **Normal sync** | `{speed: 50}` | `{speed: 50}` | ✅ No action (in_sync) |
| **Drift detected** | `{speed: 50}` | `{speed: 60}` | 🔴 Queue config update command |
| **Expired desired** | `{speed: 50, expires: T0}` | `{speed: 60}` | ✅ Clear expired, accept reported |
| **Offline > 48h** | `{speed: 50}` | `{speed: 60}` | 🔴 Queue RESYNC_CONFIG |
| **Missing reported** | `{speed: 50}` | `{}` | ⚠️ Wait for heartbeat |
| **Missing desired** | `{}` | `{speed: 60}` | ✅ Use reported as baseline |

---

## 3️⃣ The OTA Failure: Safe-Mode Rollback Logic

### Current OTA Rollout Implementation

**Location**: `internal/domains/diagnostics/models/models.go`

```go
// CURRENT ROLLOUT MODEL
type FirmwareRollout struct {
    ID               uuid.UUID
    FirmwareID       uuid.UUID
    Status           string  // scheduled, in_progress, paused, completed, cancelled
    RolloutStrategy  string  // gradual, canary, full
    TotalDevices     int
    CompletedDevices int
    FailedDevices    int
    // ...
}

// CURRENT: No device-level rollback state tracking
type FirmwareRolloutEvent struct {
    ID          uuid.UUID
    RolloutID   uuid.UUID
    EventType   string  // rollout.note, rollout.transition, rollout.created
    FromStatus  *string
    ToStatus    *string
    // ...
}
```

### ⚠️ CRITICAL GAP: No Device-Side Rollback Logic

**Problem**:
- No A/B partition tracking in database schema
- No boot failure detection mechanism
- No automatic rollback trigger
- No safe-mode state definition

**Impact**: If OTA fails at 90%, device becomes a brick with no recovery path.

### Recommended: A/B Partition Rollback Implementation

```go
// internal/domains/diagnostics/models/models.go (ENHANCED)
package models

// DeviceFirmwareState tracks OTA state per device
type DeviceFirmwareState struct {
    DeviceID        uuid.UUID  `json:"device_id" db:"device_id"`
    RolloutID       uuid.UUID  `json:"rollout_id" db:"rollout_id"`
    
    // Partition tracking (A/B)
    ActiveSlot      string     `json:"active_slot" db:"active_slot"`  // A or B
    PreviousSlot    string     `json:"previous_slot" db:"previous_slot"`
    
    // OTA state machine
    OTAState        string     `json:"ota_state" db:"ota_state"`
    // scheduled → downloading → downloaded → verifying → installing → rebooting → validating → success|failed|rolled_back
    
    // Progress tracking
    DownloadPercent int        `json:"download_percent" db:"download_percent"`  // 0-100
    InstallPhase    string     `json:"install_phase" db:"install_phase"`  // unpack, write, verify
    
    // Failure tracking
    FailureReason   *string    `json:"failure_reason,omitempty" db:"failure_reason"`
    FailurePhase    *string    `json:"failure_phase,omitempty" db:"failure_phase"`
    RetryCount      int        `json:"retry_count" db:"retry_count"`
    
    // Rollback tracking
    RollbackAvailable bool     `json:"rollback_available" db:"rollback_available"`
    RollbackTriggered *time.Time `json:"rollback_triggered,omitempty" db:"rollback_triggered"`
    RollbackReason    *string    `json:"rollback_reason,omitempty" db:"rollback_reason"`
    
    // Health check
    LastHeartbeat   *time.Time `json:"last_heartbeat,omitempty" db:"last_heartbeat"`
    PostBootHealth  *string    `json:"post_boot_health,omitempty" db:"post_boot_health"`  // pass, fail, timeout
    
    UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// OTAState defines valid OTA states
type OTAState string

const (
    OTAStateScheduled    OTAState = "scheduled"
    OTAStateDownloading  OTAState = "downloading"
    OTAStateDownloaded   OTAState = "downloaded"
    OTAStateVerifying    OTAState = "verifying"
    OTAStateInstalling   OTAState = "installing"
    OTAStateRebooting    OTAState = "rebooting"
    OTAStateValidating   OTAState = "validating"
    OTAStateSuccess      OTAState = "success"
    OTAStateFailed       OTAState = "failed"
    OTAStateRolledBack   OTAState = "rolled_back"
)

// Valid OTA state transitions
var validOTATransitions = map[OTAState][]OTAState{
    OTAStateScheduled:   {OTAStateDownloading, OTAStateFailed},
    OTAStateDownloading: {OTAStateDownloaded, OTAStateFailed},
    OTAStateDownloaded:  {OTAStateVerifying, OTAStateFailed},
    OTAStateVerifying:   {OTAStateInstalling, OTAStateFailed},
    OTAStateInstalling:  {OTAStateRebooting, OTAStateFailed},
    OTAStateRebooting:   {OTAStateValidating, OTAStateFailed},
    OTAStateValidating:  {OTAStateSuccess, OTAStateFailed, OTAStateRolledBack},
    OTAStateSuccess:     {},  // Terminal
    OTAStateFailed:      {},  // Terminal
    OTAStateRolledBack:  {},  // Terminal
}

// IsValidOTATransition checks if state transition is allowed
func IsValidOTATransition(from, to OTAState) bool {
    allowed, exists := validOTATransitions[from]
    if !exists {
        return false
    }
    for _, state := range allowed {
        if state == to {
            return true
        }
    }
    return false
}
```

### Device-Side Safe-Mode Implementation

```c
// firmware/src/ota/ota_manager.c (DEVICE-SIDE IMPLEMENTATION)

#include "ota_manager.h"
#include "bootloader.h"
#include "partition.h"
#include "watchdog.h"

// A/B Partition layout
#define SLOT_A_ADDRESS    0x08000000
#define SLOT_B_ADDRESS    0x08100000
#define SLOT_SIZE         0x00100000  // 1MB each

// OTA State machine
typedef enum {
    OTA_STATE_IDLE,
    OTA_STATE_DOWNLOADING,
    OTA_STATE_DOWNLOAD_COMPLETE,
    OTA_STATE_VERIFYING,
    OTA_STATE_INSTALLING,
    OTA_STATE_REBOOTING,
    OTA_STATE_VALIDATING,
    OTA_STATE_SUCCESS,
    OTA_STATE_FAILED,
    OTA_STATE_ROLLED_BACK
} OTAState_t;

// OTA Context
typedef struct {
    OTAState_t state;
    uint8_t active_slot;  // 0 = A, 1 = B
    uint8_t previous_slot;
    uint32_t download_size;
    uint32_t download_progress;
    char failure_reason[64];
    char failure_phase[32];
    uint8_t retry_count;
    bool rollback_available;
} OTAContext_t;

static OTAContext_t ota_ctx = {0};

/**
 * @brief Start OTA update process
 * @param firmware_url URL to download firmware from
 * @param expected_hash SHA256 hash of expected firmware
 * @return OTA_ERROR_NONE on success
 */
OTAError_t OTA_StartUpdate(const char* firmware_url, const uint8_t* expected_hash) {
    // 1. Check prerequisites
    if (BSP_GetBatteryLevel() < 20) {
        snprintf(ota_ctx.failure_reason, sizeof(ota_ctx.failure_reason), 
                 "Battery too low: %d%%", BSP_GetBatteryLevel());
        strcpy(ota_ctx.failure_phase, "prerequisite_check");
        return OTA_ERROR_LOW_BATTERY;
    }
    
    if (PARTITION_GetFreeSpace() < FIRMWARE_IMAGE_SIZE) {
        strcpy(ota_ctx.failure_reason, "Insufficient storage");
        strcpy(ota_ctx.failure_phase, "prerequisite_check");
        return OTA_ERROR_NO_SPACE;
    }
    
    // 2. Determine target slot (opposite of active)
    uint8_t target_slot = (ota_ctx.active_slot == 0) ? 1 : 0;
    ota_ctx.previous_slot = ota_ctx.active_slot;
    
    // 3. Start download
    ota_ctx.state = OTA_STATE_DOWNLOADING;
    OTA_PersistState();  // Save to flash for recovery
    
    OTAError_t err = OTA_DownloadFirmware(firmware_url, target_slot);
    if (err != OTA_ERROR_NONE) {
        ota_ctx.state = OTA_STATE_FAILED;
        strcpy(ota_ctx.failure_phase, "download");
        OTA_PersistState();
        return err;
    }
    
    // 4. Verify signature and hash
    ota_ctx.state = OTA_STATE_VERIFYING;
    OTA_PersistState();
    
    err = OTA_VerifyFirmware(target_slot, expected_hash);
    if (err != OTA_ERROR_NONE) {
        ota_ctx.state = OTA_STATE_FAILED;
        strcpy(ota_ctx.failure_phase, "verification");
        snprintf(ota_ctx.failure_reason, sizeof(ota_ctx.failure_reason),
                 "Signature/hash mismatch");
        OTA_PersistState();
        return err;
    }
    
    // 5. Mark target slot as valid
    PARTITION_MarkSlotValid(target_slot);
    
    // 6. Schedule reboot
    ota_ctx.state = OTA_STATE_REBOOTING;
    OTA_PersistState();
    
    // 7. Set watchdog for boot validation
    WATCHDOG_SetTimeout(300000);  // 5 minutes for boot validation
    WATCHDOG_Enable();
    
    // 8. Reboot into new slot
    BOOTLOADER_SetActiveSlot(target_slot);
    NVIC_SystemReset();
    
    return OTA_ERROR_NONE;
}

/**
 * @brief Boot-time validation (called from main.c after boot)
 * @return true if boot successful
 */
bool OTA_ValidateBoot(void) {
    ota_ctx.state = OTA_STATE_VALIDATING;
    OTA_PersistState();
    
    // 1. Run self-tests
    if (!OTA_RunSelfTests()) {
        strcpy(ota_ctx.failure_reason, "Self-test failed");
        strcpy(ota_ctx.failure_phase, "post_boot_validation");
        ota_ctx.state = OTA_STATE_FAILED;
        OTA_PersistState();
        return false;
    }
    
    // 2. Check critical peripherals
    if (!OTA_CheckPeripherals()) {
        strcpy(ota_ctx.failure_reason, "Peripheral check failed");
        strcpy(ota_ctx.failure_phase, "post_boot_validation");
        ota_ctx.state = OTA_STATE_FAILED;
        OTA_PersistState();
        return false;
    }
    
    // 3. Connect to backend and report success
    if (!OTA_ReportSuccess()) {
        // Non-fatal, continue
    }
    
    // 4. Boot validation passed
    ota_ctx.state = OTA_STATE_SUCCESS;
    ota_ctx.rollback_available = true;  // Can still rollback if needed
    OTA_PersistState();
    
    // 5. Disable watchdog (boot successful)
    WATCHDOG_Disable();
    
    return true;
}

/**
 * @brief Automatic rollback on boot failure
 * @return true if rollback successful
 */
bool OTA_AutoRollback(void) {
    // 1. Check if rollback is available
    if (!ota_ctx.rollback_available) {
        return false;
    }
    
    // 2. Switch back to previous slot
    uint8_t rollback_slot = ota_ctx.active_slot;  // Current active is the failed one
    uint8_t previous_slot = ota_ctx.previous_slot;
    
    // 3. Mark failed slot as invalid
    PARTITION_MarkSlotInvalid(rollback_slot);
    
    // 4. Switch to previous slot
    BOOTLOADER_SetActiveSlot(previous_slot);
    
    // 5. Update state
    ota_ctx.state = OTA_STATE_ROLLED_BACK;
    ota_ctx.active_slot = previous_slot;
    ota_ctx.rollback_triggered = time(NULL);
    snprintf(ota_ctx.rollback_reason, sizeof(ota_ctx.rollback_reason),
             "Auto-rollback: boot validation failed");
    OTA_PersistState();
    
    // 6. Reboot into safe slot
    NVIC_SystemReset();
    
    return true;
}

/**
 * @brief Manual rollback trigger (from backend command)
 * @return true if rollback initiated
 */
bool OTA_TriggerManualRollback(void) {
    if (!ota_ctx.rollback_available) {
        return false;
    }
    
    // 1. Mark current slot as failed
    PARTITION_MarkSlotInvalid(ota_ctx.active_slot);
    
    // 2. Switch to previous slot
    BOOTLOADER_SetActiveSlot(ota_ctx.previous_slot);
    
    // 3. Update state
    ota_ctx.state = OTA_STATE_ROLLED_BACK;
    ota_ctx.rollback_triggered = time(NULL);
    strcpy(ota_ctx.rollback_reason, "Manual rollback from backend");
    OTA_PersistState();
    
    // 4. Reboot
    NVIC_SystemReset();
    
    return true;
}

/**
 * @brief Persist OTA state to flash (survives reboot)
 */
void OTA_PersistState(void) {
    // Write to dedicated flash sector (preserved across reboots)
    FLASH_Write(OTA_STATE_ADDRESS, &ota_ctx, sizeof(OTAContext_t));
}

/**
 * @brief Load OTA state from flash (called at boot)
 */
void OTA_LoadState(void) {
    FLASH_Read(OTA_STATE_ADDRESS, &ota_ctx, sizeof(OTAContext_t));
    
    // Check if we need to validate boot
    if (ota_ctx.state == OTA_STATE_REBOOTING) {
        // Just rebooted after OTA install
        if (!OTA_ValidateBoot()) {
            // Boot validation failed - trigger auto-rollback
            OTA_AutoRollback();
        }
    }
}
```

### 90% Failure Rollback Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    OTA FAILURE AT 90% - ROLLBACK FLOW                   │
└─────────────────────────────────────────────────────────────────────────┘

T0: OTA Update Initiated
    → Current firmware: v2.1 (Slot A - active)
    → Target firmware: v2.2 (Slot B - inactive)
    → State: scheduled

T1: Download Phase (0% → 100%)
    → Download progress: 0% → 90% → 100%
    → State: downloading → downloaded
    → ✅ Download successful

T2: Verification Phase
    → Verify SHA256 hash
    → Verify signature (ed25519)
    → ✅ Verification successful
    → State: verifying → installing

T3: Installation Phase (0% → 90% → FAILURE)
    → Unpack firmware image to Slot B
    → Write to flash: 0% → 50% → 90%
    → ⚠️ FLASH WRITE ERROR at 90%
    → Error: "Block 0x080F0000 failed to program"
    
    Device-side handling:
    1. Stop installation
    2. Set state: FAILED
    3. Set failure_phase: "installation"
    4. Set failure_reason: "Flash write error at 90%"
    5. Persist state to flash
    6. Rollback available: true (Slot A still valid)
    7. Report failure to backend

T4: Backend Receives Failure
    → Update rollout_device_status:
       - device_id: DEV001
       - status: failed
       - failure_phase: installation
       - failure_reason: Flash write error at 90%
       - rollback_available: true
    
    → Backend decision:
       - Retry? NO (flash error unlikely to succeed)
       - Rollback? YES (previous slot valid)
    
    → Send ROLLBACK command to device

T5: Device Receives Rollback Command
    → Validate command signature
    → Check rollback_available: true
    → Execute OTA_TriggerManualRollback():
       1. Mark Slot B as invalid
       2. Set active slot = Slot A (previous)
       3. Update state: rolled_back
       4. Persist to flash
       5. Reboot

T6: Reboot into Safe Slot
    → Bootloader reads active slot = A
    → Boot from Slot A (v2.1 - known good)
    → Run self-tests: ✅ PASS
    → Run peripheral checks: ✅ PASS
    → Report success to backend
    
    → State: rolled_back (terminal)
    → Device operational on v2.1

T7: Backend Updates Rollout
    → Update firmware_rollouts:
       - failed_devices: +1
       - rolled_back_devices: +1
    
    → Create rollout_event:
       - event_type: rollout.transition
       - from_status: in_progress
       - to_status: paused (if failure rate > threshold)
       - message: "Device DEV001 rolled back after install failure"
    
    → Alert engineering team if failure rate > 5%
```

### Database Schema for OTA Tracking

```sql
-- migrations/065_rollout_device_status.sql (ENHANCED)
CREATE TABLE rollout_device_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rollout_id UUID REFERENCES firmware_rollouts(id),
    device_id UUID NOT NULL,
    org_id UUID REFERENCES organizations(id),
    
    -- OTA state machine
    ota_state VARCHAR(50) DEFAULT 'scheduled',
    -- scheduled, downloading, downloaded, verifying, installing, rebooting, validating, success, failed, rolled_back
    
    -- Progress tracking
    download_percent INTEGER DEFAULT 0,
    install_phase VARCHAR(50),
    
    -- A/B partition tracking
    active_slot VARCHAR(10),  -- A or B
    previous_slot VARCHAR(10),
    
    -- Failure tracking
    failure_reason TEXT,
    failure_phase VARCHAR(50),
    retry_count INTEGER DEFAULT 0,
    
    -- Rollback tracking
    rollback_available BOOLEAN DEFAULT false,
    rollback_triggered_at TIMESTAMPTZ,
    rollback_reason TEXT,
    
    -- Health check
    post_boot_health VARCHAR(20),  -- pass, fail, timeout
    last_heartbeat TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_rollout_device_status_rollout ON rollout_device_status(rollout_id);
CREATE INDEX idx_rollout_device_status_device ON rollout_device_status(device_id);
CREATE INDEX idx_rollout_device_status_state ON rollout_device_status(rollout_id, ota_state);
CREATE INDEX idx_rollout_device_status_failed ON rollout_device_status(rollout_id) WHERE ota_state = 'failed';
```

### OTA State Machine Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    FIRMWARE OTA STATE MACHINE                           │
└─────────────────────────────────────────────────────────────────────────┘

                    ┌─────────────┐
                    │  SCHEDULED  │
                    └──────┬──────┘
                           │ Start OTA
                           ▼
                    ┌─────────────┐
                    │ DOWNLOADING │──────┐
                    └──────┬──────┘      │ Error
                           │ Complete    ▼
                    ┌─────────────┐  ┌─────────────┐
                    │ DOWNLOADED  │  │   FAILED    │
                    └──────┬──────┘  └──────┬──────┘
                           │ Verify         │
                           ▼                │
                    ┌─────────────┐         │
                    │  VERIFYING  │─────────┤
                    └──────┬──────┘         │
                           │ Valid          │
                           ▼                │
                    ┌─────────────┐         │
                    │ INSTALLING  │─────────┤
                    └──────┬──────┘         │
                           │ Write          │
                           ▼                │
                    ┌─────────────┐         │
                    │  REBOOTING  │─────────┤
                    └──────┬──────┘         │
                           │ Boot           │
                           ▼                │
                    ┌─────────────┐         │
                    │ VALIDATING  │─────────┤
                    └──────┬──────┘         │
                           │                │
            ┌──────────────┼──────────────┐ │
            │              │              │ │
            ▼              ▼              ▼ ▼
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │   SUCCESS   │ │   FAILED    │ │  ROLLED     │
    │  (terminal) │ │  (terminal) │ │   BACK      │
    └─────────────┘ └─────────────┘ │  (terminal) │
                                    └─────────────┘
```

---

## Summary: Production-Ready Configuration

### 1. Version Drift Mitigation

```bash
# Telemetry version support
TELEMETRY_ADAPTER_ENABLED=true
LEGACY_V1_SUNSET_DATE=2026-06-30
MINIMUM_SUPPORTED_VERSION=v1.5
```

### 2. Digital Twin Reconciliation

```bash
# Twin sync settings
TWIN_DRIFT_CHECK_INTERVAL=60s
TWIN_OFFLINE_THRESHOLD=48h
TWIN_RESYNC_COMMAND_PRIORITY=high
```

### 3. OTA Rollback Safety

```bash
# OTA safety settings
OTA_AB_PARTITION_REQUIRED=true
OTA_BOOT_VALIDATION_TIMEOUT=300s
OTA_AUTO_ROLLBACK_ENABLED=true
OTA_MAX_RETRY_COUNT=2
```

---

## Conclusion

The current implementation has **three critical gaps**:

1. **No version negotiation** → 10,000 legacy devices will break on schema change
2. **Digital twin unimplemented** → No actual shadow state reconciliation
3. **No A/B partition rollback** → OTA failure at 90% = bricked device

**Recommended immediate actions**:

1. Implement telemetry adapter pattern for multi-version support
2. Complete digital twin service implementation
3. Add A/B partition tracking and auto-rollback logic

These changes ensure **zero-downtime firmware updates** and **backward compatibility** with legacy devices.
