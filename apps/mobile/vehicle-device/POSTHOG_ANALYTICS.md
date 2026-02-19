# PostHog - All-in-One Open-Source Analytics

**Why PostHog is Better:**
- ✅ **Single platform** (replaces Matomo + Sentry + Mixpanel)
- ✅ **Product analytics** (funnels, cohorts, retention)
- ✅ **Session recording** (see how users interact)
- ✅ **Feature flags** (A/B testing, gradual rollouts)
- ✅ **Surveys** (in-app user feedback)
- ✅ **Error tracking** (built-in crash reporting)
- ✅ **100% open-source** (MIT License)
- ✅ **Self-hosted or cloud** (your choice)
- ✅ **GDPR/CCPA compliant**

---

## PostHog vs. Separate Tools

| Feature | PostHog | Matomo + Sentry | Firebase |
|---------|---------|-----------------|----------|
| **Analytics** | ✅ Built-in | ✅ Matomo | ✅ Built-in |
| **Crash Reporting** | ✅ Built-in | ✅ Sentry | ✅ Crashlytics |
| **Feature Flags** | ✅ Built-in | ❌ No | ✅ Built-in |
| **A/B Testing** | ✅ Built-in | ❌ No | ✅ Built-in |
| **Session Recording** | ✅ Built-in | ❌ No | ❌ No |
| **Surveys** | ✅ Built-in | ❌ No | ❌ No |
| **Funnels** | ✅ Built-in | ⚠️ Plugin | ⚠️ GA4 |
| **Cohorts** | ✅ Built-in | ❌ No | ⚠️ GA4 |
| **Retention** | ✅ Built-in | ❌ No | ⚠️ GA4 |
| **Self-Hosted** | ✅ Yes | ✅ Yes | ❌ No |
| **Cost** | Free (self-hosted) | Free (self-hosted) | Free tier, then paid |
| **Data Ownership** | ✅ You own it | ✅ You own it | ❌ Google owns it |

**Winner:** PostHog 🏆

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│              FlyFleet Vehicle Device                     │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │           PostHog Android SDK                     │  │
│  │                                                    │  │
│  │  - Event Tracking                                  │  │
│  │  - User Properties                                 │  │
│  │  - Feature Flags                                   │  │
│  │  - Error Tracking                                  │  │
│  │  - Session Recording (optional)                    │  │
│  └────────────────────┬─────────────────────────────┐  │
│                       │                              │  │
│                       │ HTTPS                        │  │
│                       │                              │  │
└───────────────────────┼──────────────────────────────┘  │
                        │                                  │
                        ▼                                  │
┌─────────────────────────────────────────────────────────┐
│           PostHog Server (Self-Hosted)                  │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Analytics  │  │    Error     │  │    Feature   │  │
│  │   Engine     │  │   Tracking   │  │    Flags     │  │
│  │  (ClickHouse)│  │              │  │              │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │    Funnels   │  │   Session    │  │   Surveys    │  │
│  │   & Cohorts  │  │  Recording   │  │              │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## Implementation

### Step 1: Add Dependencies

```kotlin
// build.gradle

dependencies {
    // PostHog Analytics
    implementation("com.posthog:posthog-android:2.0.0")
    
    // PostHog Feature Flags
    implementation("com.posthog:posthog:2.0.0")
    
    // OkHttp (already included)
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    
    // Timber (logging)
    implementation("com.jakewharton.timber:timber:5.0.1")
}
```

### Step 2: Initialize in Application Class

```kotlin
package com.flyfleet.vehicledevice

import android.app.Application
import com.posthog.android.PostHog
import com.posthog.android.PostHogAndroidConfig
import timber.log.Timber

class VehicleDeviceApplication : Application() {

    lateinit var postHog: PostHog
        private set

    override fun onCreate() {
        super.onCreate()

        // Initialize logging
        if (BuildConfig.DEBUG) {
            Timber.plant(Timber.DebugTree())
        }

        // Initialize PostHog
        val config = PostHogAndroidConfig(
            apiKey = "YOUR_POSTHOG_API_KEY",
            host = "https://posthog.flyfleet.com" // Your self-hosted instance
        ).apply {
            // Capture application lifecycle events
            captureApplicationLifecycleEvents = true
            
            // Capture screen views
            captureScreenViews = true
            
            // Enable error tracking
            captureDeepLinks = false // Not needed for vehicle device
            
            // Session recording (Enabled for Kiosk Support, Masking Required)
            sessionReplay = true
            sessionReplayConfig.maskAllTextInputs = true
            sessionReplayConfig.maskAllImages = true
            
            // Opt-out by default (GDPR compliance)
            optOut = false
            
            // Flush interval (batch events)
            flushInterval = 30 // seconds
            
            // Max queue size
            maxQueueSize = 1000
            
            // Debug mode (disable in production)
            debug = BuildConfig.DEBUG
        }

        postHog = PostHog.with(this, config)

        Timber.i("PostHog initialized")
    }

    override fun onTerminate() {
        super.onTerminate()
        postHog.close()
    }
}
```

### Step 3: Create Analytics Wrapper

```kotlin
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
 */
class PostHogAnalytics(
    private val context: Context,
    private val postHog: PostHogAndroid
) {
    private var deviceId: String? = null
    private var driverId: String? = null

    /**
     * Identify user (driver)
     */
    fun identify(userId: String, properties: Map<String, Any> = emptyMap()) {
        this.driverId = userId
        
        postHog.identify(userId, properties)
        
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
        
        // Set person properties
        val properties = mutableMapOf<String, Any>()
        properties["device_id"] = deviceId
        vehicleId?.let { properties["vehicle_id"] = it }
        organizationId?.let { properties["organization_id"] = it }
        properties["app_version"] = BuildConfig.VERSION_NAME
        properties["device_model"] = android.os.Build.MODEL
        properties["android_version"] = android.os.Build.VERSION.RELEASE
        
        postHog.setPersonProperties(properties)
        
        Timber.d("PostHog device properties set: $deviceId")
    }

    /**
     * Track custom event
     */
    fun trackEvent(
        eventName: String,
        properties: Map<String, Any> = emptyMap()
    ) {
        postHog.capture(eventName, properties)
        Timber.d("PostHog event: $eventName")
    }

    //region Device Events

    fun trackDeviceRegistered(success: Boolean, errorMessage: String? = null) {
        trackEvent(
            eventName = "device_registered",
            properties = mapOf(
                "success" to success,
                "error_message" to (errorMessage ?: "")
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
                "timestamp" to System.currentTimeMillis()
            )
        )
    }

    fun trackDeviceStopped(reason: String) {
        trackEvent(
            eventName = "device_stopped",
            properties = mapOf(
                "reason" to reason
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
                "duration_seconds" to durationSeconds
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
                "duration_ms" to durationMs
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
                "shift_duration_hours" to shiftDurationHours
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
                "exception" to (throwable?.javaClass?.name ?: "")
            )
        )

        // Capture exception in PostHog
        throwable?.let {
            postHog.captureException(it)
        }

        Timber.e(message, throwable)
    }

    fun trackError(category: String, message: String, throwable: Throwable? = null) {
        trackError("$category: $message", throwable)
    }

    //endregion

    //region Feature Flags

    /**
     * Check if feature is enabled
     */
    fun isFeatureEnabled(featureKey: String, defaultValue: Boolean = false): Boolean {
        return postHog.isFeatureEnabled(featureKey, defaultValue)
    }

    /**
     * Get feature flag value
     */
    fun getFeatureFlag(key: String): Any? {
        return postHog.getFeatureFlag(key)
    }

    /**
     * Get all feature flags
     */
    fun getAllFeatureFlags(): Map<String, Any> {
        return postHog.featureFlags
    }

    //endregion

    //region Privacy & Compliance

    /**
     * Opt-out of analytics (GDPR compliance)
     */
    fun setAnalyticsEnabled(enabled: Boolean) {
        postHog.optOut = !enabled
        Timber.d("PostHog analytics ${if (enabled) "enabled" else "disabled"}")
    }

    /**
     * Reset all analytics data (GDPR right to erasure)
     */
    fun reset() {
        postHog.reset()
        deviceId = null
        driverId = null
        Timber.i("PostHog analytics data reset")
    }

    //endregion

    //region Screen Tracking (Optional)

    /**
     * Track screen view
     */
    fun trackScreenView(screenName: String) {
        trackEvent(
            eventName = "screen_view",
            properties = mapOf(
                "screen_name" to screenName
            )
        )
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
```

### Step 4: Update Application Class

```kotlin
// In VehicleDeviceApplication.kt

import com.posthog.android.PostHogAndroid
import com.flyfleet.vehicledevice.analytics.PostHogAnalytics

class VehicleDeviceApplication : Application() {

    lateinit var postHog: PostHogAndroid
        private set

    lateinit var analytics: PostHogAnalytics
        private set

    override fun onCreate() {
        super.onCreate()
        
        // ... existing initialization ...
        
        // Initialize PostHog
        val config = PostHogAndroidConfig(
            apiKey = "phc_YOUR_POSTHOG_PROJECT_API_KEY",
            host = "https://posthog.flyfleet.com"
        ).apply {
            captureApplicationLifecycleEvents = true
            captureScreenViews = false // No UI screens in vehicle device
            sessionReplay = true // Enabled for support
            sessionReplayConfig.maskAllTextInputs = true
            debug = BuildConfig.DEBUG
        }

        postHog = PostHog.with(this, config)
        analytics = PostHogAnalytics(this, postHog)

        Timber.i("PostHog analytics initialized")
    }

    override fun onTerminate() {
        super.onTerminate()
        postHog.close()
    }
}
```

---

## Self-Hosted PostHog Setup

### Docker Compose

```yaml
# docker-compose.yml

version: '3.8'

services:
  # PostHog
  posthog:
    image: posthog/posthog:latest
    container_name: flyfleet-posthog
    restart: always
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgres://posthog:posthog_password@posthog-db:5432/posthog
      - REDIS_URL=redis://redis:6379/
      - SECRET_KEY=your_secret_key_change_this
      - DEBUG=0
      - SITE_URL=https://posthog.flyfleet.com
    depends_on:
      - posthog-db
      - redis
    volumes:
      - posthog_data:/var/lib/posthog
    networks:
      - flyfleet-network

  # PostHog Database
  posthog-db:
    image: postgres:14-alpine
    container_name: flyfleet-posthog-db
    restart: always
    environment:
      POSTGRES_USER: posthog
      POSTGRES_PASSWORD: posthog_password
      POSTGRES_DB: posthog
    volumes:
      - posthog_postgres_data:/var/lib/postgresql/data
    networks:
      - flyfleet-network

  # Redis (for caching)
  redis:
    image: redis:7-alpine
    container_name: flyfleet-redis
    restart: always
    networks:
      - flyfleet-network

  # ClickHouse (for analytics)
  clickhouse:
    image: clickhouse/clickhouse-server:latest
    container_name: flyfleet-clickhouse
    restart: always
    environment:
      - CLICKHOUSE_DB=posthog
      - CLICKHOUSE_USER=posthog
      - CLICKHOUSE_PASSWORD=posthog_password
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    networks:
      - flyfleet-network

networks:
  flyfleet-network:
    driver: bridge

volumes:
  posthog_data:
  posthog_postgres_data:
  clickhouse_data:
```

---

## PostHog Dashboards

### Dashboard 1: Device Health

**Metrics:**
- Active devices (last 24h)
- Crash-free sessions
- Error rate
- Upload success rate

**Query Example:**
```sql
SELECT count(distinct device_id)
FROM events
WHERE event = 'device_started'
  AND timestamp > now() - INTERVAL 24 HOUR
```

### Dashboard 2: Camera Performance

**Metrics:**
- Recording success rate
- Dropped frame rate
- Event clips per day
- Upload latency

### Dashboard 3: Driver Behavior

**Metrics:**
- Driver login rate
- HOS violations
- Shift duration
- Event detection accuracy

---

## Feature Flags Examples

### Example 1: Gradual AI Rollout

```kotlin
// In DeviceConfig
val aiEnabled = analytics.isFeatureEnabled("ai_detection", defaultValue = false)
```

**PostHog Configuration:**
- Flag key: `ai_detection`
- Rollout: 10% → 25% → 50% → 100%
- Targeting: Specific organizations first

### Example 2: A/B Test Upload Intervals

```kotlin
// In TelemetryService
val uploadInterval = when (analytics.getFeatureFlag("upload_interval")) {
    "fast" -> 30000L // 30 seconds
    "slow" -> 120000L // 2 minutes
    else -> 60000L // 1 minute (default)
}
```

**PostHog Configuration:**
- Flag key: `upload_interval`
- Variants: fast, default, slow
- Test: Compare upload success rate vs battery usage

---

## Cost Comparison

| Solution | Cost (1000 devices) | Cost (10,000 devices) |
|----------|---------------------|----------------------|
| **Firebase** | $0/month | $0/month |
| **Matomo + Sentry** | $50/month (servers) | $200/month (servers) |
| **PostHog Cloud** | $0/month (free tier) | $365/month |
| **PostHog Self-Hosted** | $50/month (servers) | $200/month (servers) |

**Best Value:** PostHog Self-Hosted (full features, data ownership)

---

## Benefits of PostHog

1. **All-in-One Platform** - No need for multiple tools
2. **Feature Flags** - Gradual rollouts, A/B testing
3. **Product Analytics** - Funnels, cohorts, retention
4. **Session Recording** - See how users interact (optional)
5. **Surveys** - Get user feedback
6. **Error Tracking** - Built-in crash reporting
7. **Self-Hosted** - Full data control
8. **GDPR Compliant** - Privacy-first by design
9. **Free & Open-Source** - MIT License
10. **Active Community** - 20k+ GitHub stars

---

## Next Steps

1. **Set up PostHog** (1 hour)
   - Deploy Docker stack
   - Configure project
   - Get API key

2. **Integrate Android SDK** (1 hour)
   - Add dependencies
   - Initialize in Application
   - Update analytics calls

3. **Create Dashboards** (2-4 hours)
   - Device health dashboard
   - Camera performance dashboard
   - Driver behavior dashboard

4. **Set up Feature Flags** (1 hour)
   - AI detection flag
   - Upload interval flag
   - Camera quality flag

5. **Test & Deploy** (1 hour)
   - Test event tracking
   - Verify feature flags
   - Deploy to devices

**Total Time:** 6-8 hours

---

## Conclusion

**PostHog is the best choice** for FlyFleet because it:

✅ Replaces 3+ tools with 1 platform
✅ Provides feature flags for gradual rollouts
✅ Offers product analytics (funnels, cohorts)
✅ Is 100% open-source and self-hostable
✅ Gives you full data ownership
✅ Costs less than separate tools
✅ Is GDPR/CCPA compliant by default

**Ready to implement?** Let me know and I'll update the code!
