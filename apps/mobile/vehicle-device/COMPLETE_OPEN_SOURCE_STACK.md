# Complete Open-Source Analytics Stack

**100% Free & Open-Source - No Vendor Lock-in**

---

## Recommended Stack

### 1. **PostHog** - Product Analytics & Feature Flags
- **License:** MIT
- **Purpose:** Event tracking, feature flags, A/B testing, session recording
- **Website:** https://posthog.com
- **Self-Hosted:** Yes (Docker)
- **Cloud:** Free tier available (1M events/month)

### 2. **Metabase** - Business Intelligence Dashboards
- **License:** AGPL
- **Purpose:** SQL-based dashboards, reports, analytics
- **Website:** https://metabase.com
- **Self-Hosted:** Yes (Docker, JAR)
- **Cloud:** Free for basic use

### 3. **Grafana** - Real-time Monitoring Dashboards
- **License:** AGPL
- **Purpose:** Time-series metrics, alerting, monitoring
- **Website:** https://grafana.com
- **Self-Hosted:** Yes
- **Cloud:** Free tier available

### 4. **Prometheus** - Metrics Collection
- **License:** Apache 2.0
- **Purpose:** Metrics collection, alerting
- **Website:** https://prometheus.io
- **Self-Hosted:** Yes

### 5. **Meilisearch** - Search & Analytics (Optional)
- **License:** MIT
- **Purpose:** Full-text search, event search
- **Website:** https://meilisearch.com
- **Self-Hosted:** Yes

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
│  └────────────────────┬─────────────────────────────┐  │
│                       │                              │  │
│                       │ HTTPS                        │  │
│                       │                              │  │
└───────────────────────┼──────────────────────────────┘  │
                        │                                  │
                        ▼                                  │
┌─────────────────────────────────────────────────────────┐
│           Self-Hosted Analytics Stack                   │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   PostHog    │  │   Metabase   │  │   Grafana    │  │
│  │  (Analytics) │  │     (BI)     │  │  (Monitoring)│  │
│  │  ClickHouse  │  │   PostgreSQL │  │  Prometheus  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │  Prometheus  │  │ Meilisearch  │                     │
│  │   (Metrics)  │  │   (Search)   │                     │
│  └──────────────┘  └──────────────┘                     │
└─────────────────────────────────────────────────────────┘
```

---

## Complete Docker Compose

```yaml
version: '3.8'

services:
  # PostHog (Product Analytics)
  posthog:
    image: posthog/posthog:latest
    container_name: flyfleet-posthog
    restart: always
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgres://posthog:posthog_password@posthog-db:5432/posthog
      - REDIS_URL=redis://redis:6379/
      - SECRET_KEY=your_secret_key_change_this_in_production
      - DEBUG=0
      - SITE_URL=https://analytics.flyfleet.com
      - CLICKHOUSE_HOST=clickhouse
      - CLICKHOUSE_DATABASE=posthog
      - CLICKHOUSE_USER=posthog
      - CLICKHOUSE_PASSWORD=posthog_password
    depends_on:
      - posthog-db
      - redis
      - clickhouse
    volumes:
      - posthog_data:/var/lib/posthog
    networks:
      - flyfleet-network

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

  # ClickHouse (for PostHog analytics)
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

  # Metabase (BI Dashboards)
  metabase:
    image: metabase/metabase:latest
    container_name: flyfleet-metabase
    restart: always
    ports:
      - "3001:3000"
    environment:
      - MB_DB_TYPE=postgres
      - MB_DB_DBNAME=metabase
      - MB_DB_PORT=5432
      - MB_DB_USER=metabase
      - MB_DB_PASS=metabase_password
      - MB_DB_HOST=metabase-db
    depends_on:
      - metabase-db
    volumes:
      - metabase_data:/metabase-data
    networks:
      - flyfleet-network

  metabase-db:
    image: postgres:14-alpine
    container_name: flyfleet-metabase-db
    restart: always
    environment:
      POSTGRES_USER: metabase
      POSTGRES_PASSWORD: metabase_password
      POSTGRES_DB: metabase
    volumes:
      - metabase_postgres_data:/var/lib/postgresql/data
    networks:
      - flyfleet-network

  # Grafana (Real-time Monitoring)
  grafana:
    image: grafana/grafana:latest
    container_name: flyfleet-grafana
    restart: always
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin_password_change_me
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    depends_on:
      - prometheus
    networks:
      - flyfleet-network

  # Prometheus (Metrics Collection)
  prometheus:
    image: prom/prometheus:latest
    container_name: flyfleet-prometheus
    restart: always
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
    networks:
      - flyfleet-network

  # Redis (Caching for PostHog)
  redis:
    image: redis:7-alpine
    container_name: flyfleet-redis
    restart: always
    networks:
      - flyfleet-network

  # Meilisearch (Optional - for event search)
  meilisearch:
    image: getmeili/meilisearch:latest
    container_name: flyfleet-meilisearch
    restart: always
    ports:
      - "7700:7700"
    environment:
      - MEILI_MASTER_KEY=your_meilisearch_master_key
    volumes:
      - meilisearch_data:/meili_data
    networks:
      - flyfleet-network

networks:
  flyfleet-network:
    driver: bridge

volumes:
  posthog_data:
  posthog_postgres_data:
  clickhouse_data:
  metabase_data:
  metabase_postgres_data:
  grafana_data:
  prometheus_data:
  meilisearch_data:
```

---

## Implementation

### Android Application Class

```kotlin
package com.flyfleet.vehicledevice

import android.app.Application
import com.posthog.android.PostHogAndroid
import com.posthog.android.PostHogAndroidConfig
import com.flyfleet.vehicledevice.analytics.PostHogAnalytics
import timber.log.Timber

class VehicleDeviceApplication : Application() {

    lateinit var postHog: PostHogAndroid
        private set

    lateinit var analytics: PostHogAnalytics
        private set

    override fun onCreate() {
        super.onCreate()

        // Initialize logging
        if (BuildConfig.DEBUG) {
            Timber.plant(Timber.DebugTree())
        }

        // Initialize PostHog
        val config = PostHogAndroidConfig(
            apiKey = "phc_YOUR_POSTHOG_PROJECT_API_KEY", // Get from PostHog UI
            host = "https://analytics.flyfleet.com" // Your PostHog instance
        ).apply {
            // Capture application lifecycle events
            captureApplicationLifecycleEvents = true
            
            // Don't capture screen views (no UI in vehicle device)
            captureScreenViews = false
            
            // Disable session replay (privacy)
            sessionReplay = false
            
            // Batch events for efficiency
            flushInterval = 30 // seconds
            maxQueueSize = 1000
            
            // Debug mode (disable in production)
            debug = BuildConfig.DEBUG
            
            // Opt-out default (GDPR compliance)
            optOut = false
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

### PostHog Analytics Class

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

---

## Setup Instructions

### 1. Deploy the Stack

```bash
# Clone the repo
cd /path/to/flyfleet

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f posthog
```

### 2. Configure PostHog

1. Open http://localhost:8000
2. Create admin account
3. Create new project: "FlyFleet Vehicle Device"
4. Get API key from Project Settings
5. Update `VehicleDeviceApplication.kt` with your API key

### 3. Configure Metabase

1. Open http://localhost:3001
2. Create admin account
3. Add data source: PostHog PostgreSQL database
4. Create dashboards

### 4. Configure Grafana

1. Open http://localhost:3000
2. Login: admin / admin_password_change_me
3. Add Prometheus data source
4. Import dashboards

---

## Cost Breakdown

| Component | Self-Hosted Cost | Cloud Cost |
|-----------|-----------------|------------|
| **PostHog** | $50/month (servers) | Free-365$/month |
| **Metabase** | Free | Free-500$/month |
| **Grafana** | Free | Free-200$/month |
| **Prometheus** | Free | Free |
| **Total** | **$50-200/month** | **$0-1000+/month** |

**Best Value:** Self-hosted (full control, lower cost)

---

## Next Steps

1. **Deploy Docker stack** (1 hour)
2. **Configure PostHog** (30 min)
3. **Update Android app** (1 hour)
4. **Test event tracking** (1 hour)
5. **Create dashboards** (2-4 hours)

**Total:** 5-7 hours to full implementation

Ready to deploy? Let me know!
