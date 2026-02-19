package com.flyfleet.vehicledevice.analytics

import android.content.Context
import android.os.Bundle
import timber.log.Timber

/**
 * Mobile Analytics Manager for FlyFleet Vehicle Device
 * 
 * Uses PostHog as the analytics backend (open-source).
 * Wraps PostHogAnalytics with high-level tracking methods.
 * 
 * Features:
 * - Custom event tracking
 * - User property tracking
 * - Error reporting
 * - Performance tracing
 * - Privacy-compliant data collection
 */
class MobileAnalytics(
    private val context: Context,
    private val postHogAnalytics: PostHogAnalytics
) {

    companion object {
        private const val TAG = "MobileAnalytics"
    }

    // Internal state
    private var analyticsEnabled = true
    private var deviceId: String? = null
    private var vehicleId: String? = null
    private var driverId: String? = null
    private var organizationId: String? = null
    private val activeTraces = mutableMapOf<String, Long>()

    /**
     * Initialize analytics
     */
    fun initialize() {
        if (!analyticsEnabled) {
            Timber.d("Analytics disabled - skipping initialization")
            return
        }

        postHogAnalytics.initialize()
        setDefaultUserProperties()

        Timber.i("MobileAnalytics initialized")
    }

    /**
     * Set device identifiers (call after device registration)
     */
    fun setDeviceIdentifiers(
        deviceId: String,
        vehicleId: String? = null,
        organizationId: String? = null
    ) {
        this.deviceId = deviceId
        this.vehicleId = vehicleId
        this.organizationId = organizationId

        postHogAnalytics.setDeviceProperties(deviceId, vehicleId, organizationId)

        Timber.d("Device identifiers set: $deviceId")
    }

    /**
     * Set driver identifier (call after driver login)
     */
    fun setDriverId(driverId: String) {
        this.driverId = driverId
        postHogAnalytics.identify(driverId, mapOf("role" to "driver"))
        Timber.d("Driver ID set: $driverId")
    }

    /**
     * Clear driver identifier (call after driver logout)
     */
    fun clearDriverId() {
        this.driverId = null
        Timber.d("Driver ID cleared")
    }

    //region Device Events

    /**
     * Track device registration
     */
    fun trackDeviceRegistered(success: Boolean, errorMessage: String? = null) {
        postHogAnalytics.trackDeviceRegistered(success, errorMessage)
    }

    /**
     * Track device start/stop
     */
    fun trackDeviceStarted() {
        postHogAnalytics.trackDeviceStarted()
    }

    fun trackDeviceStopped(reason: String) {
        postHogAnalytics.trackDeviceStopped(reason)
    }

    //endregion

    //region Camera Events

    /**
     * Track camera events
     */
    fun trackCameraStarted(resolution: String, fps: Int) {
        postHogAnalytics.trackCameraStarted(resolution, fps)
    }

    fun trackCameraStopped() {
        postHogAnalytics.trackCameraStopped()
    }

    /**
     * Track event clip recorded
     */
    fun trackClipRecorded(eventType: String, durationSeconds: Int) {
        postHogAnalytics.trackClipRecorded(eventType, durationSeconds)
    }

    /**
     * Track clip upload
     */
    fun trackClipUploaded(clipId: String, success: Boolean, durationMs: Long, errorMessage: String? = null) {
        postHogAnalytics.trackClipUploaded(clipId, success, durationMs, errorMessage)
    }

    //endregion

    //region Telemetry Events

    /**
     * Track telemetry upload
     */
    fun trackTelemetryUploaded(pointsCount: Long, success: Boolean, durationMs: Long) {
        postHogAnalytics.trackTelemetryUploaded(pointsCount, success, durationMs)
    }

    //endregion

    //region Driver Events

    /**
     * Track driver login/logout
     */
    fun trackDriverLogin(success: Boolean, durationMs: Long, errorMessage: String? = null) {
        postHogAnalytics.trackDriverLogin(success, durationMs, errorMessage)
    }

    fun trackDriverLogout(shiftDurationHours: Double) {
        postHogAnalytics.trackDriverLogout(shiftDurationHours)
    }

    /**
     * Track HOS update
     */
    fun trackHosUpdated(dutyStatus: String, violations: List<String>) {
        postHogAnalytics.trackHosUpdated(dutyStatus, violations)
    }

    //endregion

    //region Event Detection

    /**
     * Track event detected (AI/rule-based)
     */
    fun trackEventDetected(eventType: String, severity: String, confidence: Double) {
        postHogAnalytics.trackEventDetected(eventType, severity, confidence)
    }

    //endregion

    //region Storage

    /**
     * Track storage cleanup
     */
    fun trackStorageCleanup(clipsDeleted: Int, spaceFreedMB: Long) {
        postHogAnalytics.trackEvent(
            eventName = "storage_cleanup",
            properties = mapOf(
                "clips_deleted" to clipsDeleted,
                "space_freed_mb" to spaceFreedMB
            )
        )
    }

    //endregion

    //region Config

    /**
     * Track config update
     */
    fun trackConfigUpdated(configVersion: String, changes: List<String>) {
        postHogAnalytics.trackEvent(
            eventName = "config_updated",
            properties = mapOf(
                "config_version" to configVersion,
                "changes" to changes,
                "changes_count" to changes.size
            )
        )
    }

    //endregion

    //region Error Tracking

    /**
     * Track error
     */
    fun trackError(category: String, message: String, throwable: Throwable? = null) {
        postHogAnalytics.trackError(category, message, throwable)
    }

    //endregion

    //region Performance

    /**
     * Start performance trace
     */
    fun startTrace(traceName: String) {
        activeTraces[traceName] = System.currentTimeMillis()
        Timber.d("Trace started: $traceName")
    }

    /**
     * Track API call performance
     */
    fun trackApiCall(
        endpoint: String,
        success: Boolean,
        durationMs: Long,
        responseSizeBytes: Long? = null
    ) {
        val properties = mutableMapOf<String, Any>(
            "endpoint" to endpoint,
            "success" to success,
            "duration_ms" to durationMs
        )
        responseSizeBytes?.let { properties["response_size_bytes"] = it }

        postHogAnalytics.trackEvent(
            eventName = "api_call",
            properties = properties
        )
    }

    /**
     * Track network request
     */
    fun trackNetworkRequest(
        url: String,
        httpMethod: String,
        httpStatusCode: Int,
        responsePayloadSize: Long,
        durationMs: Long
    ) {
        postHogAnalytics.trackEvent(
            eventName = "network_request",
            properties = mapOf(
                "url" to url,
                "method" to httpMethod,
                "status_code" to httpStatusCode,
                "payload_size" to responsePayloadSize,
                "duration_ms" to durationMs
            )
        )
    }

    //endregion

    //region Privacy

    /**
     * Set analytics consent
     */
    fun setAnalyticsConsent(granted: Boolean) {
        analyticsEnabled = granted
        postHogAnalytics.setAnalyticsEnabled(granted)
    }

    /**
     * Reset analytics data (for GDPR right to erasure)
     */
    fun resetAnalyticsData() {
        postHogAnalytics.reset()
        Timber.i("Analytics data reset")
    }

    //endregion

    //region Private

    private fun setDefaultUserProperties() {
        val props = mutableMapOf<String, Any>(
            "app_version" to com.flyfleet.vehicledevice.BuildConfig.VERSION_NAME,
            "device_model" to android.os.Build.MODEL,
            "android_version" to android.os.Build.VERSION.RELEASE,
            "sdk_version" to android.os.Build.VERSION.SDK_INT
        )

        deviceId?.let { props["device_id"] = it }
        vehicleId?.let { props["vehicle_id"] = it }
        organizationId?.let { props["organization_id"] = it }

        postHogAnalytics.trackEvent("user_properties_set", props)
    }

    private fun logEvent(category: String, eventName: String, params: Bundle) {
        // Convert Bundle to map for PostHog
        val properties = mutableMapOf<String, Any>()
        properties["category"] = category
        for (key in params.keySet()) {
            @Suppress("DEPRECATION")
            params.get(key)?.let { properties[key] = it.toString() }
        }
        postHogAnalytics.trackEvent(eventName, properties)
    }

    //endregion
}
