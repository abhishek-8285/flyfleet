package com.flyfleet.vehicledevice.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

/**
 * Device Registration Request
 */
@Serializable
data class DeviceRegisterRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("hardware_attestation")
    val hardwareAttestation: String,
    
    @SerialName("android_version")
    val androidVersion: String,
    
    @SerialName("device_model")
    val deviceModel: String,
    
    @SerialName("app_version")
    val appVersion: String,
    
    @SerialName("organization_id")
    val organizationId: String? = null,
    
    @SerialName("vehicle_id")
    val vehicleId: String? = null
)

/**
 * Device Registration Response
 */
@Serializable
data class DeviceRegisterResponse(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("jwt_token")
    val jwtToken: String,
    
    @SerialName("refresh_token")
    val refreshToken: String,
    
    @SerialName("expires_at")
    val expiresAt: Long
)

/**
 * Heartbeat Request
 */
@Serializable
data class HeartbeatRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("status")
    val status: String, // ok, degraded, critical
    
    @SerialName("uptime")
    val uptime: Long, // Seconds since boot
    
    @SerialName("module_health")
    val moduleHealth: ModuleHealth,
    
    @SerialName("telemetry_summary")
    val telemetrySummary: TelemetrySummary? = null
)

/**
 * Module Health Status
 */
@Serializable
data class ModuleHealth(
    @SerialName("camera")
    val camera: ModuleStatus,
    
    @SerialName("ai")
    val ai: ModuleStatus,
    
    @SerialName("gps")
    val gps: ModuleStatus,
    
    @SerialName("can_bus")
    val canBus: ModuleStatus,
    
    @SerialName("storage")
    val storage: ModuleStatus,
    
    @SerialName("network")
    val network: ModuleStatus
)

/**
 * Single Module Status
 */
@Serializable
data class ModuleStatus(
    @SerialName("status")
    val status: String, // ok, degraded, failed
    
    @SerialName("temperature")
    val temperature: Double, // Celsius
    
    @SerialName("memory_usage")
    val memoryUsage: Double, // Percentage 0-100
    
    @SerialName("error_count")
    val errorCount: Int
)

/**
 * Telemetry Summary
 */
@Serializable
data class TelemetrySummary(
    @SerialName("gps_points_count")
    val gpsPointsCount: Int,
    
    @SerialName("can_bus_messages_count")
    val canBusMessagesCount: Int,
    
    @SerialName("average_speed_kph")
    val averageSpeedKph: Double,
    
    @SerialName("max_speed_kph")
    val maxSpeedKph: Double,
    
    @SerialName("distance_km")
    val distanceKm: Double
)

/**
 * Heartbeat Response
 */
@Serializable
data class HeartbeatResponse(
    @SerialName("status")
    val status: String,
    
    @SerialName("config_version")
    val configVersion: String,
    
    @SerialName("pending_commands")
    val pendingCommands: List<Command> = emptyList()
)

/**
 * Config Response
 */
@Serializable
data class ConfigResponse(
    @SerialName("config")
    val config: DeviceRuntimeConfig,
    
    @SerialName("signature")
    val signature: String,
    
    @SerialName("version")
    val version: String,
    
    @SerialName("expires_at")
    val expiresAt: Long
)

/**
 * Device Runtime Configuration
 */
@Serializable
data class DeviceRuntimeConfig(
    @SerialName("schema_version")
    val schemaVersion: String,
    
    @SerialName("heartbeat_seconds")
    val heartbeatSeconds: Int,
    
    @SerialName("ring_buffer_minutes")
    val ringBufferMinutes: Int,
    
    @SerialName("gps")
    val gps: GPSConfig,
    
    @SerialName("ai")
    val ai: AIConfig,
    
    @SerialName("retention")
    val retention: RetentionConfig,
    
    @SerialName("feature_flags")
    val featureFlags: FeatureFlags
)

@Serializable
data class GPSConfig(
    @SerialName("active_interval_seconds")
    val activeIntervalSeconds: Int,
    
    @SerialName("idle_interval_seconds")
    val idleIntervalSeconds: Int
)

@Serializable
data class AIConfig(
    @SerialName("enabled")
    val enabled: Boolean,
    
    @SerialName("frame_sample_stride")
    val frameSampleStride: Int,
    
    @SerialName("severity_thresholds")
    val severityThresholds: SeverityThresholds
)

@Serializable
data class SeverityThresholds(
    @SerialName("warn_min")
    val warnMin: Double,
    
    @SerialName("critical_min")
    val criticalMin: Double
)

@Serializable
data class RetentionConfig(
    @SerialName("critical_clip_days")
    val criticalClipDays: Int,
    
    @SerialName("min_storage_reserve_mb")
    val minStorageReserveMb: Int,
    
    @SerialName("legal_hold")
    val legalHold: Boolean
)

@Serializable
data class FeatureFlags(
    @SerialName("camera_enabled")
    val cameraEnabled: Boolean,
    
    @SerialName("audio_event_checks_enabled")
    val audioEventChecksEnabled: Boolean,
    
    @SerialName("ai_enabled")
    val aiEnabled: Boolean,
    
    @SerialName("can_bus_enabled")
    val canBusEnabled: Boolean,
    
    @SerialName("driver_lockout_enabled")
    val driverLockoutEnabled: Boolean
)

/**
 * Telemetry Batch Request
 */
@Serializable
data class TelemetryBatchRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("batch_id")
    val batchId: String,
    
    @SerialName("telemetry")
    val telemetry: List<TelemetryPoint>,
    
    @SerialName("timestamp")
    val timestamp: Long
)

/**
 * Single Telemetry Point
 */
@Serializable
data class TelemetryPoint(
    @SerialName("timestamp")
    val timestamp: Long,
    
    @SerialName("gps")
    val gps: GPSData,
    
    @SerialName("can_bus")
    val canBus: CANBusData? = null,
    
    @SerialName("speed_kph")
    val speedKph: Double,
    
    @SerialName("heading")
    val heading: Double,
    
    @SerialName("altitude")
    val altitude: Double,
    
    @SerialName("accuracy")
    val accuracy: Double
)

/**
 * GPS Data
 */
@Serializable
data class GPSData(
    @SerialName("latitude")
    val latitude: Double,
    
    @SerialName("longitude")
    val longitude: Double,
    
    @SerialName("accuracy")
    val accuracy: Double,
    
    @SerialName("satellites")
    val satellites: Int
)

/**
 * CAN Bus Data
 */
@Serializable
data class CANBusData(
    @SerialName("engine_rpm")
    val engineRpm: Int,
    
    @SerialName("speed_kph")
    val speedKph: Double,
    
    @SerialName("brake_pressed")
    val brakePressed: Boolean,
    
    @SerialName("turn_signal_on")
    val turnSignalOn: Boolean,
    
    @SerialName("fuel_level")
    val fuelLevel: Double,
    
    @SerialName("coolant_temp")
    val coolantTemp: Double,
    
    @SerialName("dtc_count")
    val dtcCount: Int,
    
    @SerialName("odometer_km")
    val odometerKm: Double
)

/**
 * Telemetry Response
 */
@Serializable
data class TelemetryResponse(
    @SerialName("status")
    val status: String,
    
    @SerialName("received_count")
    val receivedCount: Int,
    
    @SerialName("next_batch_id")
    val nextBatchId: String
)

/**
 * Upload URL Request
 */
@Serializable
data class UploadUrlRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("clip_id")
    val clipId: String,
    
    @SerialName("size_bytes")
    val sizeBytes: Long,
    
    @SerialName("content_type")
    val contentType: String,
    
    @SerialName("duration_seconds")
    val durationSeconds: Int,
    
    @SerialName("event_type")
    val eventType: String? = null,
    
    @SerialName("storage_mode")
    val storageMode: String = "local"
)

/**
 * Upload URL Response
 */
@Serializable
data class UploadUrlResponse(
    @SerialName("upload_url")
    val uploadUrl: String? = null,
    
    @SerialName("local_path")
    val localPath: String? = null,
    
    @SerialName("expires_at")
    val expiresAt: Long? = null,
    
    @SerialName("object_key")
    val objectKey: String,
    
    @SerialName("storage_mode")
    val storageMode: String,
    
    @SerialName("max_local_size_mb")
    val maxLocalSizeMb: Long
)

/**
 * Confirm Upload Request
 */
@Serializable
data class ConfirmUploadRequest(
    @SerialName("clip_id")
    val clipId: String,
    
    @SerialName("checksum_sha256")
    val checksumSha256: String,
    
    @SerialName("duration_seconds")
    val durationSeconds: Int,
    
    @SerialName("uploaded_bytes")
    val uploadedBytes: Long,
    
    @SerialName("storage_mode")
    val storageMode: String,
    
    @SerialName("local_path")
    val localPath: String? = null
)

/**
 * Confirm Upload Response
 */
@Serializable
data class ConfirmUploadResponse(
    @SerialName("status")
    val status: String,
    
    @SerialName("storage_location")
    val storageLocation: String,
    
    @SerialName("storage_mode")
    val storageMode: String,
    
    @SerialName("sync_status")
    val syncStatus: String
)

/**
 * Attestation Request
 */
@Serializable
data class AttestationRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("evidence_hash")
    val evidenceHash: String,
    
    @SerialName("chain_hash")
    val chainHash: String,
    
    @SerialName("timestamp")
    val timestamp: Long
)

/**
 * Attestation Response
 */
@Serializable
data class AttestationResponse(
    @SerialName("attestation_id")
    val attestationId: String,
    
    @SerialName("signature")
    val signature: String,
    
    @SerialName("timestamp")
    val timestamp: Long,
    
    @SerialName("blockchain_tx")
    val blockchainTx: String? = null
)

/**
 * Remote Command
 */
@Serializable
data class Command(
    @SerialName("id")
    val id: String,
    
    @SerialName("type")
    val type: String,
    
    @SerialName("params")
    val params: Map<String, JsonElement> = emptyMap(),
    
    @SerialName("expires_at")
    val expiresAt: Long,
    
    @SerialName("priority")
    val priority: Int,
    
    @SerialName("created_at")
    val createdAt: Long
)

/**
 * Driver Login Request
 */
@Serializable
data class DriverLoginRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("driver_id")
    val driverId: String,
    
    @SerialName("rfid_token")
    val rfidToken: String? = null,
    
    @SerialName("photo_hash")
    val photoHash: String? = null,
    
    @SerialName("duty_status")
    val dutyStatus: String
)

/**
 * Driver Login Response
 */
@Serializable
data class DriverLoginResponse(
    @SerialName("status")
    val status: String,
    
    @SerialName("driver_profile")
    val driverProfile: DriverProfile?,
    
    @SerialName("hos_status")
    val hosStatus: HOSStatus?,
    
    @SerialName("shift_start")
    val shiftStart: Long
)

/**
 * Driver Profile
 */
@Serializable
data class DriverProfile(
    @SerialName("driver_id")
    val driverId: String,
    
    @SerialName("name")
    val name: String,
    
    @SerialName("license_number")
    val licenseNumber: String,
    
    @SerialName("hos_limits")
    val hosLimits: HOSLimits
)

/**
 * HOS Limits
 */
@Serializable
data class HOSLimits(
    @SerialName("max_driving_hours")
    val maxDrivingHours: Double,
    
    @SerialName("max_on_duty_hours")
    val maxOnDutyHours: Double,
    
    @SerialName("min_break_after_hours")
    val minBreakAfterHours: Double,
    
    @SerialName("min_break_duration")
    val minBreakDuration: Double,
    
    @SerialName("max_on_duty_days")
    val maxOnDutyDays: Int
)

/**
 * HOS Status
 */
@Serializable
data class HOSStatus(
    @SerialName("duty_status")
    val dutyStatus: String,
    
    @SerialName("driving_time_remaining")
    val drivingTimeRemaining: Double,
    
    @SerialName("on_duty_time_remaining")
    val onDutyTimeRemaining: Double,
    
    @SerialName("break_required")
    val breakRequired: Boolean,
    
    @SerialName("violations")
    val violations: List<String>
)

/**
 * Driver Logout Request
 */
@Serializable
data class DriverLogoutRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("driver_id")
    val driverId: String,
    
    @SerialName("shift_duration")
    val shiftDuration: Double,
    
    @SerialName("end_duty_status")
    val endDutyStatus: String
)

/**
 * Driver Logout Response
 */
@Serializable
data class DriverLogoutResponse(
    @SerialName("status")
    val status: String,
    
    @SerialName("hos_summary")
    val hosSummary: HOSSummary,
    
    @SerialName("violations")
    val violations: List<String>
)

/**
 * HOS Summary
 */
@Serializable
data class HOSSummary(
    @SerialName("shift_start")
    val shiftStart: Long,
    
    @SerialName("shift_end")
    val shiftEnd: Long,
    
    @SerialName("total_driving")
    val totalDriving: Double,
    
    @SerialName("total_on_duty")
    val totalOnDuty: Double,
    
    @SerialName("total_breaks")
    val totalBreaks: Double,
    
    @SerialName("distance_km")
    val distanceKm: Double
)

/**
 * HOS Update Request
 */
@Serializable
data class HOSUpdateRequest(
    @SerialName("device_id")
    val deviceId: String,
    
    @SerialName("driver_id")
    val driverId: String,
    
    @SerialName("duty_status")
    val dutyStatus: String,
    
    @SerialName("timestamp")
    val timestamp: Long,
    
    @SerialName("location")
    val location: Location? = null
)

/**
 * Location
 */
@Serializable
data class Location(
    @SerialName("latitude")
    val latitude: Double,
    
    @SerialName("longitude")
    val longitude: Double,
    
    @SerialName("address")
    val address: String? = null
)

/**
 * HOS Update Response
 */
@Serializable
data class HOSUpdateResponse(
    @SerialName("status")
    val status: String,
    
    @SerialName("violations")
    val violations: List<String>,
    
    @SerialName("remaining_drive_time")
    val remainingDriveTime: Double,
    
    @SerialName("remaining_on_duty_time")
    val remainingOnDutyTime: Double
)
