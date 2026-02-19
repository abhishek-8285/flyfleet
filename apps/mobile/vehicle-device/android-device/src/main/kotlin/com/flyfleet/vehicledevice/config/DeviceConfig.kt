package com.flyfleet.vehicledevice.config

import android.content.Context
import android.content.SharedPreferences
import com.flyfleet.vehicledevice.BuildConfig

/**
 * Device Configuration Manager
 * 
 * Centralized configuration for:
 * - Backend URLs
 * - PostHog analytics
 * - Device settings
 * - Feature flags
 */
class DeviceConfig(private val context: Context) {

    private val prefs: SharedPreferences =
        context.getSharedPreferences("flyfleet_config", Context.MODE_PRIVATE)

    companion object {
        // Backend Configuration
        private const val KEY_BACKEND_URL = "backend_url"
        private const val DEFAULT_BACKEND_URL = "https://api.flyfleet.com"

        // PostHog Configuration
        private const val KEY_POSTHOG_API_KEY = "posthog_api_key"
        private const val KEY_POSTHOG_HOST = "posthog_host"
        private const val DEFAULT_POSTHOG_HOST = "https://app.posthog.com"

        // JWT Secret (for local token validation)
        private const val KEY_JWT_SECRET = "jwt_secret"

        // Device Settings
        private const val KEY_DEVICE_ID = "device_id"
        private const val KEY_VEHICLE_ID = "vehicle_id"
        private const val KEY_ORGANIZATION_ID = "organization_id"

        // Feature Flags
        private const val KEY_CAMERA_ENABLED = "camera_enabled"
        private const val KEY_AI_ENABLED = "ai_enabled"
        private const val KEY_TELEMETRY_INTERVAL = "telemetry_interval"

        fun getBackendUrl(context: Context): String {
            return context.getSharedPreferences("flyfleet_config", Context.MODE_PRIVATE)
                .getString(KEY_BACKEND_URL, DEFAULT_BACKEND_URL) ?: DEFAULT_BACKEND_URL
        }

        fun getPostHogApiKey(context: Context): String {
            // Use BuildConfig value (set in build.gradle)
            return BuildConfig.POSTHOG_API_KEY
        }

        fun getPostHogHost(context: Context): String {
            // Use BuildConfig value (set in build.gradle)
            return BuildConfig.POSTHOG_HOST
        }

        fun getJwtSecret(context: Context): String {
            return context.getSharedPreferences("flyfleet_config", Context.MODE_PRIVATE)
                .getString(KEY_JWT_SECRET, "default-jwt-secret-change-in-production") ?: ""
        }
    }

    // Backend URL
    var backendUrl: String
        get() = prefs.getString(KEY_BACKEND_URL, DEFAULT_BACKEND_URL) ?: DEFAULT_BACKEND_URL
        set(value) = prefs.edit().putString(KEY_BACKEND_URL, value).apply()

    // PostHog Configuration
    var postHogApiKey: String
        get() = prefs.getString(KEY_POSTHOG_API_KEY, "") ?: ""
        set(value) = prefs.edit().putString(KEY_POSTHOG_API_KEY, value).apply()

    var postHogHost: String
        get() = prefs.getString(KEY_POSTHOG_HOST, DEFAULT_POSTHOG_HOST) ?: DEFAULT_POSTHOG_HOST
        set(value) = prefs.edit().putString(KEY_POSTHOG_HOST, value).apply()

    // Device Identifiers
    var deviceId: String?
        get() = prefs.getString(KEY_DEVICE_ID, null)
        set(value) = prefs.edit().putString(KEY_DEVICE_ID, value).apply()

    var vehicleId: String?
        get() = prefs.getString(KEY_VEHICLE_ID, null)
        set(value) = prefs.edit().putString(KEY_VEHICLE_ID, value).apply()

    var organizationId: String?
        get() = prefs.getString(KEY_ORGANIZATION_ID, null)
        set(value) = prefs.edit().putString(KEY_ORGANIZATION_ID, value).apply()

    // Feature Flags
    var cameraEnabled: Boolean
        get() = prefs.getBoolean(KEY_CAMERA_ENABLED, true)
        set(value) = prefs.edit().putBoolean(KEY_CAMERA_ENABLED, value).apply()

    var aiEnabled: Boolean
        get() = prefs.getBoolean(KEY_AI_ENABLED, false)
        set(value) = prefs.edit().putBoolean(KEY_AI_ENABLED, value).apply()

    var telemetryIntervalMs: Long
        get() = prefs.getLong(KEY_TELEMETRY_INTERVAL, 60000L) // Default 60 seconds
        set(value) = prefs.edit().putLong(KEY_TELEMETRY_INTERVAL, value).apply()

    /**
     * Clear all configuration (for device wipe)
     */
    fun clearAll() {
        prefs.edit().clear().apply()
    }

    /**
     * Get camera configuration
     */
    fun getCameraConfig(): CameraConfig {
        return CameraConfig(
            enabled = cameraEnabled,
            ringBufferMinutes = prefs.getInt("camera_ring_buffer_minutes", 10),
            preEventWindowSeconds = prefs.getInt("camera_pre_event_seconds", 15),
            postEventWindowSeconds = prefs.getInt("camera_post_event_seconds", 30)
        )
    }

    /**
     * Get telemetry configuration
     */
    fun getTelemetryConfig(): TelemetryConfig {
        return TelemetryConfig(
            batchSize = prefs.getInt("telemetry_batch_size", 60),
            uploadIntervalMs = telemetryIntervalMs
        )
    }

    /**
     * Data classes for configuration
     */
    data class CameraConfig(
        val enabled: Boolean,
        val ringBufferMinutes: Int,
        val preEventWindowSeconds: Int,
        val postEventWindowSeconds: Int
    )

    data class TelemetryConfig(
        val batchSize: Int,
        val uploadIntervalMs: Long
    )
}
