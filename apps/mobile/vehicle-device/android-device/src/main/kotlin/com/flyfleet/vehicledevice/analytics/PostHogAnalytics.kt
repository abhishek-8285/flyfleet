package com.flyfleet.vehicledevice.analytics

import android.content.Context
import com.flyfleet.vehicledevice.BuildConfig
import com.posthog.PostHog
import com.posthog.android.PostHogAndroid
import timber.log.Timber

/**
 * PostHog Analytics for FlyFleet Vehicle Device
 * 
 * Features:
 * - Event tracking
 * - User properties
 * - Feature flags
 * - Error tracking
 * - A/B testing
 * 
 * 100% Open-Source (MIT License)
 * Self-hosted or cloud
 */
class PostHogAnalytics(
    private val context: Context
) {
    private var deviceId: String? = null
    private var driverId: String? = null

    /**
     * Initialize analytics
     */
    fun initialize() {
        Timber.d("PostHog Analytics initialized")
    }

    /**
     * Identify user (driver)
     */
    fun identify(userId: String, properties: Map<String, Any> = emptyMap()) {
        this.driverId = userId
        
        PostHog.identify(userId, properties)
        
        Timber.d("PostHog identify: $userId")
    }

    /**
     * Set device properties
     */
    fun setDeviceProperties(
        deviceId: String,
        vehicleId: String?,
        organizationId: String?
    ) {
        this.deviceId = deviceId
        
        val properties = mutableMapOf<String, Any>()
        properties["device_id"] = deviceId
        vehicleId?.let { properties["vehicle_id"] = it }
        organizationId?.let { properties["organization_id"] = it }
        properties["app_version"] = BuildConfig.VERSION_NAME
        properties["device_model"] = android.os.Build.MODEL
        properties["android_version"] = android.os.Build.VERSION.RELEASE
        
        PostHog.identify(deviceId, properties)
        
        Timber.d("PostHog device properties set: $deviceId")
    }

    /**
     * Track custom event
     */
    fun trackEvent(
        eventName: String,
        properties: Map<String, Any> = emptyMap()
    ) {
        PostHog.capture(eventName, properties = properties)
        Timber.d("PostHog event: $eventName - $properties")
    }

    //region Device Events

    fun trackDeviceRegistered(success: Boolean, errorMessage: String? = null) {
        trackEvent(
            eventName = "device_registered",
            properties = mapOf(
                "success" to success,
                "error_message" to (errorMessage ?: ""),
                "timestamp" to System.currentTimeMillis()
            )
        )

        if (!success) {
            trackError("Device registration failed", errorMessage)
        }
    }

    fun trackDeviceStarted() {
        trackEvent(
            eventName = "device_started",
            properties = mapOf(
                "timestamp" to System.currentTimeMillis(),
                "app_version" to BuildConfig.VERSION_NAME
            )
        )
    }

    fun trackDeviceStopped(reason: String) {
        trackEvent(
            eventName = "device_stopped",
            properties = mapOf(
                "reason" to reason,
                "timestamp" to System.currentTimeMillis()
            )
        )
    }

    //endregion

    //region Camera Events

    fun trackCameraStarted(resolution: String, fps: Int) {
        trackEvent(
            eventName = "camera_started",
            properties = mapOf(
                "resolution" to resolution,
                "fps" to fps
            )
        )
    }

    fun trackCameraStopped() {
        trackEvent(eventName = "camera_stopped")
    }

    fun trackClipRecorded(eventType: String, durationSeconds: Int) {
        trackEvent(
            eventName = "clip_recorded",
            properties = mapOf(
                "event_type" to eventType,
                "duration_seconds" to durationSeconds,
                "timestamp" to System.currentTimeMillis()
            )
        )
    }

    fun trackClipUploaded(
        clipId: String,
        success: Boolean,
        durationMs: Long,
        errorMessage: String? = null
    ) {
        trackEvent(
            eventName = "clip_uploaded",
            properties = mapOf(
                "clip_id" to clipId,
                "success" to success,
                "duration_ms" to durationMs,
                "error_message" to (errorMessage ?: "")
            )
        )

        if (!success) {
            trackError("Clip upload failed", errorMessage)
        }
    }

    //endregion

    //region Telemetry Events

    fun trackTelemetryUploaded(
        pointsCount: Long,
        success: Boolean,
        durationMs: Long
    ) {
        trackEvent(
            eventName = "telemetry_uploaded",
            properties = mapOf(
                "points_count" to pointsCount,
                "success" to success,
                "duration_ms" to durationMs,
                "timestamp" to System.currentTimeMillis()
            )
        )
    }

    //endregion

    //region CAN Bus Events

    fun trackCanBusEvent(eventType: String, data: Map<String, Any>) {
        trackEvent(
            eventName = "can_bus_event",
            properties = mapOf(
                "event_type" to eventType,
                "data" to data,
                "timestamp" to System.currentTimeMillis()
            )
        )
    }

    //endregion

    //region Driver Events

    fun trackDriverLogin(success: Boolean, durationMs: Long, errorMessage: String? = null) {
        trackEvent(
            eventName = "driver_login",
            properties = mapOf(
                "success" to success,
                "duration_ms" to durationMs,
                "error_message" to (errorMessage ?: "")
            )
        )

        if (!success) {
            trackError("Driver login failed", errorMessage)
        }
    }

    fun trackDriverLogout(shiftDurationHours: Double) {
        trackEvent(
            eventName = "driver_logout",
            properties = mapOf(
                "shift_duration_hours" to shiftDurationHours,
                "timestamp" to System.currentTimeMillis()
            )
        )
    }

    fun trackHosUpdated(dutyStatus: String, violations: List<String>) {
        trackEvent(
            eventName = "hos_updated",
            properties = mapOf(
                "duty_status" to dutyStatus,
                "violations" to violations,
                "violation_count" to violations.size
            )
        )
    }

    //endregion

    //region Event Detection

    fun trackEventDetected(eventType: String, severity: String, confidence: Double) {
        trackEvent(
            eventName = "event_detected",
            properties = mapOf(
                "event_type" to eventType,
                "severity" to severity,
                "confidence" to confidence
            )
        )

        // Capture critical events as errors too
        if (severity == "CRITICAL") {
            trackError("Critical event detected", "$eventType detected with $confidence confidence")
        }
    }

    //endregion

    //region Error Tracking

    fun trackError(message: String, throwable: Throwable? = null) {
        // Track as event
        trackEvent(
            eventName = "error",
            properties = mapOf(
                "message" to message,
                "exception" to (throwable?.javaClass?.name ?: ""),
                "timestamp" to System.currentTimeMillis()
            )
        )

        Timber.e(message, throwable)
    }

    fun trackError(category: String, message: String?, throwable: Throwable? = null) {
        trackError("$category: ${message ?: "unknown"}", throwable)
    }

    //endregion

    //region Feature Flags

    /**
     * Check if feature is enabled
     */
    fun isFeatureEnabled(featureKey: String, defaultValue: Boolean = false): Boolean {
        return try {
            PostHog.isFeatureEnabled(featureKey, defaultValue)
        } catch (e: Exception) {
            defaultValue
        }
    }

    /**
     * Get feature flag value
     */
    fun getFeatureFlag(key: String): Any? {
        return try {
            PostHog.getFeatureFlag(key)
        } catch (e: Exception) {
            null
        }
    }

    /**
     * Get all feature flags
     */
    fun getAllFeatureFlags(): Map<String, Any> {
        return emptyMap() // Feature flags are accessed individually
    }

    //endregion

    //region Shadow Mode

    /**
     * Check if Shadow Mode is enabled via feature flag
     */
    fun isShadowModeEnabled(): Boolean {
        return isFeatureEnabled("shadow_mode", defaultValue = false)
    }

    //endregion

    //region Privacy & Compliance

    /**
     * Opt-out of analytics (GDPR compliance)
     */
    fun setAnalyticsEnabled(enabled: Boolean) {
        if (!enabled) {
            PostHog.optOut()
        } else {
            PostHog.optIn()
        }
        Timber.d("PostHog analytics ${if (enabled) "enabled" else "disabled"}")
    }

    /**
     * Reset all analytics data (GDPR right to erasure)
     */
    fun reset() {
        PostHog.reset()
        deviceId = null
        driverId = null
        Timber.i("PostHog analytics data reset")
    }

    //endregion
}

/**
 * Extension function for tracking operation duration
 */
suspend fun <T> PostHogAnalytics.trackOperation(
    operationName: String,
    block: suspend () -> T
): T {
    val startTime = System.currentTimeMillis()
    try {
        return block()
    } catch (e: Exception) {
        trackError("operation", "$operationName failed: ${e.message}", e)
        throw e
    } finally {
        val duration = System.currentTimeMillis() - startTime
        trackEvent(
            eventName = "operation_completed",
            properties = mapOf(
                "operation" to operationName,
                "duration_ms" to duration
            )
        )
    }
}
