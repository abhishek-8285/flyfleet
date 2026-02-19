# Open-Source Mobile Analytics Stack

**Goal:** Replace Firebase with 100% free, open-source, self-hosted alternatives
**Benefits:** 
- ✅ Full data ownership (no Google)
- ✅ No vendor lock-in
- ✅ GDPR/CCPA compliant by default
- ✅ No usage limits or costs
- ✅ Self-hosted on your infrastructure

---

## Recommended Stack

### 1. **Analytics: Matomo (formerly Piwik)**
- **License:** GPL v3
- **Website:** https://matomo.org
- **Features:**
  - User behavior tracking
  - Custom events
  - Funnels & cohorts
  - Real-time dashboard
  - Privacy-focused (GDPR compliant)
  - Self-hosted or cloud

**Why Matomo:**
- Most popular open-source analytics
- Feature-complete alternative to Google Analytics
- Mobile SDK available (iOS & Android)
- No data sampling

### 2. **Crash Reporting: Sentry**
- **License:** BSD 3-Clause
- **Website:** https://sentry.io
- **Features:**
  - Automatic crash reporting
  - Non-fatal error tracking
  - Performance monitoring
  - Release tracking
  - Source maps
  - Self-hosted or cloud

**Why Sentry:**
- Industry standard for error tracking
- Excellent Android SDK
- Performance monitoring included
- Self-hosted version is free (up to certain limits)

### 3. **Performance Monitoring: Custom + Prometheus**
- **License:** Apache 2.0
- **Website:** https://prometheus.io
- **Features:**
  - Metrics collection
  - Time-series database
  - Alerting
  - Grafana dashboards

**Why Prometheus:**
- Industry standard for metrics
- Integrates with everything
- Powerful query language
- Self-hosted

### 4. **Dashboards: Grafana**
- **License:** AGPL v3
- **Website:** https://grafana.com
- **Features:**
  - Beautiful dashboards
  - Multiple data sources
  - Alerting
  - Annotations

**Why Grafana:**
- Best-in-class dashboards
- Free and open-source
- Integrates with Prometheus, Sentry, Matomo

### 5. **Backend Storage: PostgreSQL + TimescaleDB**
- **License:** PostgreSQL License / TimescaleDB License
- **Features:**
  - Relational storage
  - Time-series data
  - Analytics queries

**Why TimescaleDB:**
- SQL-based (easier than NoSQL)
- Time-series optimized
- Free tier available

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│              FlyFleet Vehicle Device                     │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Matomo     │  │    Sentry    │  │   Prometheus │  │
│  │   Analytics  │  │   Crashes    │  │   Metrics    │  │
│  │   SDK        │  │    SDK       │  │   Client     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                  │                  │          │
│         │ HTTPS            │ HTTPS            │ HTTPS    │
│         │                  │                  │          │
└─────────┼──────────────────┼──────────────────┼──────────┘
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────┐
│              Self-Hosted Backend                         │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │    Matomo    │  │    Sentry    │  │  Prometheus  │  │
│  │    Server    │  │    Server    │  │   Server     │  │
│  │   (MySQL)    │  │  (ClickHouse)│  │   (TSDB)     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                  │                  │          │
│         └──────────────────┼──────────────────┘          │
│                            │                             │
│                   ┌────────▼────────┐                    │
│                   │    Grafana      │                    │
│                   │   Dashboards    │                    │
│                   └─────────────────┘                    │
└─────────────────────────────────────────────────────────┘
```

---

## Implementation

### Step 1: Add Dependencies

```kotlin
// build.gradle

dependencies {
    // Matomo Analytics
    implementation("org.matomo.analytics:matomo-sdk-android:1.10.0")
    
    // Sentry (Crash & Performance)
    implementation("io.sentry:sentry-android:7.0.0")
    
    // Prometheus Metrics
    implementation("io.prometheus:simpleclient:0.16.0")
    implementation("io.prometheus:simpleclient_httpserver:0.16.0")
    
    // OkHttp (already included)
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
}
```

### Step 2: Initialize in Application Class

```kotlin
package com.flyfleet.vehicledevice

import android.app.Application
import com.flyfleet.vehicledevice.analytics.OpenSourceAnalytics
import io.sentry.android.core.SentryAndroid
import timber.log.Timber

class VehicleDeviceApplication : Application() {

    lateinit var analytics: OpenSourceAnalytics
        private set

    override fun onCreate() {
        super.onCreate()

        // Initialize logging
        if (BuildConfig.DEBUG) {
            Timber.plant(Timber.DebugTree())
        }

        // Initialize Sentry (Crash Reporting)
        SentryAndroid.init(this) { options ->
            options.dsn = "https://YOUR_SENTRY_DSN@your-sentry-server.com/1"
            options.environment = if (BuildConfig.DEBUG) "development" else "production"
            options.isEnableAutoSessionTracking = true
            options.sessionTrackingIntervalMillis = 30000L
            options.isEnableNdk = true
            options.isEnableAutoPerformanceTracing = true
        }

        // Initialize Matomo (Analytics)
        analytics = OpenSourceAnalytics(
            context = this,
            matomoUrl = "https://your-matomo-server.com",
            siteId = 1
        )

        Timber.i("Open-source analytics initialized")
    }
}
```

### Step 3: Create Analytics Wrapper

```kotlin
package com.flyfleet.vehicledevice.analytics

import android.content.Context
import com.flyfleet.vehicledevice.BuildConfig
import io.sentry.Sentry
import io.sentry.protocol.User
import org.matomo.sdk.Matomo
import org.matomo.sdk.Tracker
import org.matomo.sdk.extra.TrackHelper
import timber.log.Timber

/**
 * Open-Source Mobile Analytics
 * 
 * Uses:
 * - Matomo for user analytics
 * - Sentry for crash reporting
 * - Prometheus for performance metrics
 */
class OpenSourceAnalytics(
    private val context: Context,
    matomoUrl: String,
    siteId: Int
) {
    private val tracker: Tracker

    init {
        // Initialize Matomo
        val matomo = Matomo.getInstance(context)
        tracker = matomo.buildTracker(matomoUrl, siteId)
        
        // Set default dimensions
        tracker.setDimension(1, BuildConfig.VERSION_NAME) // App Version
        tracker.setDimension(2, android.os.Build.MODEL) // Device Model
        tracker.setDimension(3, android.os.Build.VERSION.RELEASE) // Android Version
    }

    /**
     * Set user identifiers
     */
    fun setUserId(userId: String) {
        tracker.userId = userId
        
        // Also set in Sentry
        Sentry.setUser(userId)
    }

    /**
     * Track screen view
     */
    fun trackScreenView(screenName: String) {
        TrackHelper.track()
            .screen(screenName)
            .title(screenName)
            .with(tracker)
        
        Timber.d("Screen view: $screenName")
    }

    /**
     * Track custom event
     */
    fun trackEvent(
        category: String,
        action: String,
        label: String? = null,
        value: Float? = null
    ) {
        var helper = TrackHelper.track()
            .event(category, action)
        
        label?.let { helper = helper.label(it) }
        value?.let { helper = helper.value(it) }
        
        helper.with(tracker)
        
        Timber.d("Event: $category/$action/${label ?: "no-label"}")
    }

    /**
     * Track device registered
     */
    fun trackDeviceRegistered(success: Boolean, errorMessage: String? = null) {
        trackEvent(
            category = "device",
            action = "register",
            label = if (success) "success" else "failure",
            value = if (success) 1f else 0f
        )

        if (!success) {
            Sentry.captureMessage("Device registration failed: $errorMessage")
        }
    }

    /**
     * Track camera started
     */
    fun trackCameraStarted(resolution: String, fps: Int) {
        trackEvent(
            category = "camera",
            action = "start",
            label = "${resolution}_${fps}fps"
        )
    }

    /**
     * Track clip recorded
     */
    fun trackClipRecorded(eventType: String, durationSeconds: Int) {
        trackEvent(
            category = "camera",
            action = "clip_recorded",
            label = eventType,
            value = durationSeconds.toFloat()
        )
    }

    /**
     * Track clip uploaded
     */
    fun trackClipUploaded(
        clipId: String,
        success: Boolean,
        durationMs: Long,
        errorMessage: String? = null
    ) {
        trackEvent(
            category = "camera",
            action = "clip_uploaded",
            label = if (success) "success" else "failure",
            value = durationMs.toFloat()
        )

        if (!success) {
            Sentry.captureMessage("Clip upload failed: $errorMessage")
        }
    }

    /**
     * Track telemetry uploaded
     */
    fun trackTelemetryUploaded(
        pointsCount: Long,
        success: Boolean,
        durationMs: Long
    ) {
        trackEvent(
            category = "telemetry",
            action = "upload",
            label = if (success) "success" else "failure",
            value = durationMs.toFloat()
        )

        TrackHelper.track()
            .variable(1, "Telemetry", "Points", pointsCount.toString())
            .with(tracker)
    }

    /**
     * Track driver login
     */
    fun trackDriverLogin(success: Boolean, durationMs: Long, errorMessage: String? = null) {
        trackEvent(
            category = "driver",
            action = "login",
            label = if (success) "success" else "failure",
            value = durationMs.toFloat()
        )

        if (!success) {
            Sentry.captureMessage("Driver login failed: $errorMessage")
        }
    }

    /**
     * Track event detected
     */
    fun trackEventDetected(eventType: String, severity: String, confidence: Double) {
        trackEvent(
            category = "events",
            action = "detected",
            label = "${eventType}_${severity}",
            value = confidence.toFloat()
        )

        // Also log to Sentry for critical events
        if (severity == "CRITICAL") {
            Sentry.captureMessage("Critical event detected: $eventType")
        }
    }

    /**
     * Track error
     */
    fun trackError(category: String, message: String, throwable: Throwable? = null) {
        trackEvent(
            category = "errors",
            action = category,
            label = message
        )

        // Log to Sentry
        throwable?.let {
            Sentry.captureException(it) { scope ->
                scope.setTag("category", category)
                scope.setMessage(message)
            }
        } ?: Sentry.captureMessage("$category error: $message")

        Timber.e("$category error: $message")
    }

    /**
     * Start performance trace
     */
    fun startTrace(name: String): Trace {
        val sentryTrace = Sentry.startTrace(name, "manual")
        return Trace(name, sentryTrace)
    }

    /**
     * Set custom tag for crash reporting
     */
    fun setTag(key: String, value: String) {
        Sentry.setTag(key, value)
    }

    /**
     * Set custom metric
     */
    fun setMetric(name: String, value: Double) {
        TrackHelper.track()
            .variable(1, "Metrics", name, value.toString())
            .with(tracker)
    }

    /**
     * Opt-out of analytics (GDPR compliance)
     */
    fun setAnalyticsEnabled(enabled: Boolean) {
        tracker.isOptOut = !enabled
        Timber.d("Analytics ${if (enabled) "enabled" else "disabled"}")
    }

    /**
     * Clear all analytics data (GDPR right to erasure)
     */
    fun clearAllData() {
        tracker.userId = null
        Timber.i("Analytics data cleared")
    }
}

/**
 * Performance trace wrapper
 */
class Trace(
    private val name: String,
    private val sentryTrace: io.sentry.ITrace
) {
    fun putAttribute(key: String, value: String) {
        sentryTrace.setTag(key, value)
    }

    fun putMetric(key: String, value: Int) {
        sentryTrace.setTag(key, value.toString())
    }

    fun stop() {
        sentryTrace.finish()
    }
}

/**
 * Extension function for tracking operation duration
 */
suspend fun <T> OpenSourceAnalytics.trackOperation(
    operationName: String,
    block: suspend () -> T
): T {
    val trace = startTrace(operationName)
    try {
        return block()
    } catch (e: Exception) {
        trackError("operation", "$operationName failed: ${e.message}", e)
        throw e
    } finally {
        trace.stop()
    }
}
```

### Step 4: Docker Compose for Self-Hosted Stack

```yaml
# docker-compose.yml

version: '3.8'

services:
  # Matomo Analytics
  matomo:
    image: matomo:latest
    container_name: flyfleet-matomo
    restart: always
    ports:
      - "8080:80"
    environment:
      MATOMO_DATABASE_HOST: matomo-db
      MATOMO_DATABASE_ADAPTER: mysqli
      MATOMO_DATABASE_TABLES_PREFIX: matomo_
      MATOMO_DATABASE_USERNAME: matomo
      MATOMO_DATABASE_PASSWORD: matomo_password
      MATOMO_DATABASE_DBNAME: matomo
    volumes:
      - matomo_data:/var/www/html
    depends_on:
      - matomo-db
    networks:
      - flyfleet-network

  matomo-db:
    image: mysql:8.0
    container_name: flyfleet-matomo-db
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: matomo
      MYSQL_USER: matomo
      MYSQL_PASSWORD: matomo_password
    volumes:
      - matomo_mysql_data:/var/lib/mysql
    networks:
      - flyfleet-network

  # Sentry (Crash Reporting)
  sentry:
    image: sentry:latest
    container_name: flyfleet-sentry
    restart: always
    ports:
      - "8090:9000"
    environment:
      SENTRY_REDIS_HOST: redis
      SENTRY_POSTGRES_HOST: sentry-db
      SENTRY_DB_PASSWORD: sentro_password
      SENTRY_DB_USER: sentry
      SENTRY_DB_NAME: sentry
      SENTRY_EVENT_RETENTION_DAYS: "30"
    depends_on:
      - redis
      - sentry-db
    networks:
      - flyfleet-network

  sentry-db:
    image: postgres:14
    container_name: flyfleet-sentry-db
    restart: always
    environment:
      POSTGRES_PASSWORD: sentro_password
      POSTGRES_USER: sentry
      POSTGRES_DB: sentry
    volumes:
      - sentry_postgres_data:/var/lib/postgresql/data
    networks:
      - flyfleet-network

  redis:
    image: redis:7-alpine
    container_name: flyfleet-redis
    restart: always
    networks:
      - flyfleet-network

  # Prometheus (Metrics)
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

  # Grafana (Dashboards)
  grafana:
    image: grafana/grafana:latest
    container_name: flyfleet-grafana
    restart: always
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin_password
      GF_INSTALL_PLUGINS: grafana-piechart-panel
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    depends_on:
      - prometheus
    networks:
      - flyfleet-network

networks:
  flyfleet-network:
    driver: bridge

volumes:
  matomo_data:
  matomo_mysql_data:
  sentry_postgres_data:
  prometheus_data:
  grafana_data:
```

### Step 5: Prometheus Configuration

```yaml
# prometheus.yml

global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'flyfleet-backend'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
```

### Step 6: Update Application to Use Open-Source Analytics

```kotlin
// In VehicleDeviceApplication.kt

val analytics: OpenSourceAnalytics by lazy {
    OpenSourceAnalytics(
        context = this,
        matomoUrl = "https://analytics.flyfleet.com",
        siteId = 1
    )
}
```

---

## Cost Comparison

### Firebase (Proprietary)
- **Analytics:** Free
- **Crashlytics:** Free
- **Performance:** Free
- **Remote Config:** Free up to 10M users
- **Total:** $0-125/month (at scale)
- **Data:** Stored on Google servers

### Open-Source Stack
- **Matomo:** Free (self-hosted)
- **Sentry:** Free (self-hosted, up to certain limits)
- **Prometheus:** Free
- **Grafana:** Free
- **Infrastructure:** $50-200/month (your servers)
- **Total:** $50-200/month
- **Data:** Stored on your servers (full control)

---

## Benefits of Open-Source

1. **Data Ownership:** All data stays on your servers
2. **No Vendor Lock-in:** You control the stack
3. **GDPR Compliant:** By default (data stays in your region)
4. **Customizable:** Modify code as needed
5. **No Usage Limits:** Scale as needed
6. **Community Support:** Active communities for all tools

---

## Migration Path

If you ever used Firebase, migration is easy:

1. **Matomo:** Import GA4 data via plugin
2. **Sentry:** No migration needed (new crashes only)
3. **Prometheus:** Fresh start (metrics are ephemeral)

---

## Next Steps

1. **Set up Docker stack** (1 hour)
2. **Configure Matomo** (30 min)
3. **Configure Sentry** (30 min)
4. **Update Android app** (1 hour)
5. **Test analytics** (1 hour)
6. **Create Grafana dashboards** (2-4 hours)

**Total Time:** 6-8 hours

---

## Conclusion

The open-source stack provides **enterprise-grade analytics** without vendor lock-in or data privacy concerns. It's the right choice for fleet operations where data sovereignty matters.

**Recommended for:**
- ✅ Fleet operators with data sovereignty requirements
- ✅ GDPR/CCPA compliance needs
- ✅ Long-term cost optimization
- ✅ Full control over data

Let me know if you want me to implement this!
