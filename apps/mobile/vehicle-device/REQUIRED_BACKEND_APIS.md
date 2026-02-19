# Vehicle Device - Required Backend APIs Analysis

**Analysis Date:** 2026-02-15
**Purpose:** Identify ONLY the essential backend APIs needed for the FlyFleet Vehicle Device Android App

---

## Executive Summary

After analyzing the full FlyFleet backend API surface (535+ endpoints across 16 services), the Vehicle Device app requires only **21 critical endpoints** from **4 backend services**.

**Key Finding:** 96% of backend APIs are NOT needed for vehicle device operation.

---

## Required APIs by Priority

### 🔴 CRITICAL (Must Have - Device Cannot Function Without)

| # | Endpoint | Method | Service | Purpose | Request | Response |
|---|----------|--------|---------|---------|---------|----------|
| 1 | `/device/register` | POST | tracking-service | Device provisioning with attestation | `{ device_id, hardware_attestation, android_version }` | `{ device_id, jwt_token, refresh_token }` |
| 2 | `/device/heartbeat` | POST | tracking-service | Health heartbeat every 30-60s | `{ device_id, status, telemetry, module_health }` | `{ status: "ok", config_version }` |
| 3 | `/device/config` | GET | tracking-service | Fetch signed runtime config | `{ device_id, config_version }` | `{ config: {...}, signature, version }` |
| 4 | `/telemetry/batch` | POST | tracking-service | GPS + CAN bus telemetry upload | `{ device_id, telemetry: [...], timestamp }` | `{ status: "accepted", received_count }` |
| 5 | `/events` | POST | alert-service | AI detections, rule triggers, panic events | `{ device_id, event_type, severity, confidence, metadata }` | `{ event_id, status: "accepted" }` |
| 6 | `/media/upload-url` | POST | tracking-service | Get pre-signed URL for video clip upload | `{ device_id, clip_id, size_bytes, content_type }` | `{ upload_url, expires_at }` |
| 7 | `/media/confirm` | POST | tracking-service | Confirm upload completion | `{ clip_id, checksum, duration_seconds }` | `{ status: "confirmed" }` |
| 8 | `/driver/login` | POST | auth-service | Driver RFID/NFC login with HOS tracking | `{ device_id, driver_id, rfid_token, photo_hash }` | `{ status: "success", driver_profile, hos_status }` |
| 9 | `/driver/logout` | POST | auth-service | Driver logout | `{ device_id, driver_id, shift_duration }` | `{ status: "success", hos_summary }` |
| 10 | `/driver/hos` | POST | auth-service | HOS duty status updates | `{ device_id, driver_id, duty_status, timestamp }` | `{ status: "accepted", violations: [] }` |

### 🟡 HIGH (Should Have - Core Functionality)

| # | Endpoint | Method | Service | Purpose | Request | Response |
|---|----------|--------|---------|---------|---------|----------|
| 11 | `/device/diagnostics` | POST | tracking-service | Remote diagnostics (on-demand) | `{ device_id, logs, metrics }` | `{ status: "received", commands: [] }` |
| 12 | `/forensics/attest` | POST | tracking-service | Evidence integrity attestation | `{ device_id, evidence_hash, chain_hash }` | `{ attestation_id, signature, timestamp }` |
| 13 | `/alerts/acknowledge` | POST | alert-service | Acknowledge backend-generated alerts | `{ device_id, alert_id, status }` | `{ status: "acknowledged" }` |
| 14 | `/device/commands` | GET | tracking-service | Fetch pending remote commands | `{ device_id, last_command_id }` | `{ commands: [{id, type, params, expires_at}] }` |

### 🟢 MEDIUM (Nice to Have - Enhanced Features)

| # | Endpoint | Method | Service | Purpose | Request | Response |
|---|----------|--------|---------|---------|---------|----------|
| 15 | `/device/firmware` | GET | tracking-service | Check for firmware/app updates | `{ device_id, current_version }` | `{ update_available, version, download_url }` |
| 16 | `/geofences` | GET | fleet-core-service | Fetch geofence definitions | `{ device_id, org_id }` | `{ geofences: [{id, name, geometry, type}] }` |
| 17 | `/trips/active` | GET | fleet-core-service | Get active trip for current driver | `{ device_id, driver_id }` | `{ trip: {id, route, waypoints} }` |
| 18 | `/trips/start` | POST | fleet-core-service | Start new trip | `{ device_id, driver_id, vehicle_id }` | `{ trip_id, status: "started" }` |
| 19 | `/trips/end` | POST | fleet-core-service | End current trip | `{ device_id, trip_id, end_location }` | `{ trip_id, status: "ended", summary }` |

### ⚪ LOW (Optional - Future Features)

| # | Endpoint | Method | Service | Purpose |
|---|----------|--------|---------|---------|
| 20 | `/analytics/driver-score` | GET | analytics-service | Fetch driver scorecard for coaching |
| 21 | `/maintenance/schedule` | GET | maintenance-service | Fetch upcoming maintenance schedule |

---

## APIs NOT Needed (96% of Backend)

The following backend services are **NOT required** for vehicle device operation:

### Auth Service - 60+ endpoints (Need only 2)
**NOT Needed:**
- ❌ `/api/v1/auth/login` (device uses token-based auth, not user login)
- ❌ `/api/v1/auth/register` (devices are provisioned, not self-registered)
- ❌ `/api/v1/auth/refresh-token` (device uses long-lived device tokens)
- ❌ `/api/v1/auth/profile` (devices don't have user profiles)
- ❌ `/api/v1/user/theme` (irrelevant for headless device)
- ❌ ALL MFA endpoints (`/api/v1/security/mfa/*`) - devices don't use MFA
- ❌ ALL admin endpoints (`/api/v1/admin/*`) - device is not an admin client
- ❌ ALL RBAC endpoints (`/api/v1/admin/rbac/*`) - device has fixed permissions
- ❌ ALL audit endpoints (`/api/v1/admin/audit/*`) - device sends logs, doesn't query audit

**ONLY Needed:**
- ✅ `/driver/login` (for HOS tracking)
- ✅ `/driver/logout` (for HOS tracking)
- ✅ `/driver/hos` (for HOS tracking)

### Fleet Core Service - 100+ endpoints (Need only 4)
**NOT Needed:**
- ❌ `/api/v1/vehicles` CRUD (devices don't manage vehicle registry)
- ❌ `/api/v1/drivers` CRUD (devices don't manage driver registry)
- ❌ `/api/v1/maintenance/*` (optional, not core device function)
- ❌ `/api/v1/work-orders/*` (maintenance management, not device function)
- ❌ `/api/v1/inventory/*` (parts management, not device function)
- ❌ `/api/v1/geofencing/*` (optional, only needed if geofence alerts required)
- ❌ `/api/v1/fuel/*` (fuel tracking, not core device function)
- ❌ `/api/v1/routes/*` (route optimization, not device function)
- ❌ `/api/v1/diagnostics/*` (device sends diagnostics, doesn't query)

**ONLY Needed:**
- ✅ `/geofences` GET (for geofence entry/exit alerts)
- ✅ `/trips/active` GET (to associate events with trip)
- ✅ `/trips/start` POST (optional, if device can start trips)
- ✅ `/trips/end` POST (optional, if device can end trips)

### Alert Service - 8 endpoints (Need only 2)
**NOT Needed:**
- ❌ `/api/v1/alerts` GET (device sends alerts, doesn't query)
- ❌ `/api/v1/alerts/stats/*` (analytics, not device function)

**ONLY Needed:**
- ✅ `/events` POST (for AI detections and rule triggers)
- ✅ `/alerts/acknowledge` POST (optional, for driver acknowledgment)

### Analytics Service - 43 endpoints (Need 0-1)
**NOT Needed:**
- ❌ ALL analytics endpoints (device sends raw data, backend computes analytics)

**OPTIONALLY Needed:**
- ⚪ `/analytics/driver-score` GET (for in-cab driver coaching display)

### Other Services - 300+ endpoints (Need 0)
**NOT Needed:**
- ❌ Billing Service (13 endpoints) - device doesn't handle billing
- ❌ Reports Service (11 endpoints) - device doesn't generate reports
- ❌ Governance Service (10 endpoints) - device doesn't manage governance
- ❌ Integrations Service (4 endpoints) - device doesn't manage integrations
- ❌ Settings Service (12 endpoints) - device fetches config, doesn't manage settings
- ❌ India Integration Service (FASTag, e-way bills) - backend handles this
- ❌ Video Service (video streams) - device uploads clips, doesn't manage streams
- ❌ Route Optimization Service - backend handles route planning
- ❌ Maintenance Service - device doesn't manage maintenance schedules

---

## Minimal API Client Implementation

Based on this analysis, the Vehicle Device API client needs only:

### 1. Device Service (tracking-service)
```kotlin
interface DeviceApi {
    @POST("/device/register")
    suspend fun register(@Body request: DeviceRegisterRequest): DeviceRegisterResponse
    
    @POST("/device/heartbeat")
    suspend fun heartbeat(@Body request: HeartbeatRequest): HeartbeatResponse
    
    @GET("/device/config")
    suspend fun getConfig(@Query("device_id") deviceId: String): ConfigResponse
    
    @POST("/telemetry/batch")
    suspend fun uploadTelemetry(@Body request: TelemetryBatchRequest): TelemetryResponse
    
    @POST("/media/upload-url")
    suspend fun getUploadUrl(@Body request: UploadUrlRequest): UploadUrlResponse
    
    @POST("/media/confirm")
    suspend fun confirmUpload(@Body request: ConfirmUploadRequest): ConfirmUploadResponse
    
    @POST("/device/diagnostics")
    suspend fun sendDiagnostics(@Body request: DiagnosticsRequest): DiagnosticsResponse
    
    @POST("/forensics/attest")
    suspend fun attestEvidence(@Body request: AttestationRequest): AttestationResponse
    
    @GET("/device/commands")
    suspend fun getCommands(@Query("device_id") deviceId: String): CommandsResponse
    
    @GET("/device/firmware")
    suspend fun checkFirmware(@Query("device_id") deviceId: String): FirmwareResponse
}
```

### 2. Driver Service (auth-service)
```kotlin
interface DriverApi {
    @POST("/driver/login")
    suspend fun login(@Body request: DriverLoginRequest): DriverLoginResponse
    
    @POST("/driver/logout")
    suspend fun logout(@Body request: DriverLogoutRequest): DriverLogoutResponse
    
    @POST("/driver/hos")
    suspend fun updateHos(@Body request: HosUpdateRequest): HosResponse
}
```

### 3. Event Service (alert-service)
```kotlin
interface EventApi {
    @POST("/events")
    suspend fun submitEvent(@Body request: EventRequest): EventResponse
    
    @POST("/alerts/acknowledge")
    suspend fun acknowledgeAlert(@Body request: AcknowledgeRequest): AcknowledgeResponse
}
```

### 4. Fleet Service (fleet-core-service) - OPTIONAL
```kotlin
interface FleetApi {
    @GET("/geofences")
    suspend fun getGeofences(@Query("device_id") deviceId: String): GeofencesResponse
    
    @GET("/trips/active")
    suspend fun getActiveTrip(@Query("driver_id") driverId: String): TripResponse
    
    @POST("/trips/start")
    suspend fun startTrip(@Body request: StartTripRequest): StartTripResponse
    
    @POST("/trips/end")
    suspend fun endTrip(@Body request: EndTripRequest): EndTripResponse
}
```

---

## Backend Implementation Gaps

### ❌ Missing Endpoints (Need to be created)

The following endpoints **DO NOT EXIST** in current backend and must be implemented:

1. **`POST /device/register`** - Device provisioning with hardware attestation
2. **`POST /device/heartbeat`** - Health heartbeat endpoint
3. **`GET /device/config`** - Signed config fetch
4. **`POST /telemetry/batch`** - Batch telemetry upload (GPS + CAN)
5. **`POST /events`** - Event submission (AI detections, rule triggers)
6. **`POST /media/upload-url`** - Pre-signed URL for clip upload
7. **`POST /media/confirm`** - Upload confirmation
8. **`POST /driver/login`** - Driver login with HOS tracking
9. **`POST /driver/logout`** - Driver logout
10. **`POST /driver/hos`** - HOS duty status updates
11. **`POST /forensics/attest`** - Evidence integrity attestation
12. **`GET /device/commands`** - Remote command queue

### ⚠️ Existing But Need Modification

1. **`POST /device/diagnostics`** - Exists in spec but may need enhancement for device-specific diagnostics
2. **`GET /geofences`** - Exists in fleet-core but may need device-specific filtering
3. **`GET /trips/active`** - May need to be created or modified

---

## Recommended Implementation Priority

### Phase 1 (Week 1-2): Core Device Operations
1. `POST /device/register` - Device provisioning
2. `POST /device/heartbeat` - Health monitoring
3. `GET /device/config` - Remote config
4. `POST /telemetry/batch` - Telemetry upload
5. `POST /events` - Event submission

### Phase 2 (Week 3-4): Media & Evidence
6. `POST /media/upload-url` - Clip upload URL
7. `POST /media/confirm` - Upload confirmation
8. `POST /forensics/attest` - Evidence attestation

### Phase 3 (Week 5-6): Driver & HOS
9. `POST /driver/login` - Driver login
10. `POST /driver/logout` - Driver logout
11. `POST /driver/hos` - HOS tracking

### Phase 4 (Week 7-8): Enhanced Features
12. `GET /device/commands` - Remote commands
13. `GET /geofences` - Geofence definitions
14. `GET /trips/active`, `POST /trips/start`, `POST /trips/end` - Trip management

---

## Conclusion

**Total APIs Required:** 21 endpoints
**Critical (Must Have):** 10 endpoints
**High (Should Have):** 4 endpoints  
**Medium (Nice to Have):** 5 endpoints
**Low (Optional):** 2 endpoints

**Backend Services Needed:** 4 out of 16 (25%)
- tracking-service (primary)
- auth-service (driver/HOS)
- alert-service (events)
- fleet-core-service (optional: geofences, trips)

**Development Effort:**
- New endpoints to create: 12
- Existing endpoints to modify: 2
- Total backend effort: 2-3 weeks for experienced Go developer

This focused API surface allows the Vehicle Device app to be **independent, resilient, and efficient** without coupling to unnecessary backend complexity.
