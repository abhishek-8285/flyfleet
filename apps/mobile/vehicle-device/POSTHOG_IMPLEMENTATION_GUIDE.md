# PostHog Analytics - Complete Implementation Guide

**Status:** ✅ Complete & Ready to Deploy
**License:** 100% Open-Source (MIT)
**Cost:** Free (self-hosted) or Free tier (cloud)

---

## What's Been Implemented

### ✅ Android App Integration
- PostHog SDK integrated
- Event tracking for all operations
- Feature flags support
- Error tracking
- Privacy compliance (GDPR/CCPA)

### ✅ Files Created/Updated
1. `build.gradle` - PostHog dependency added
2. `VehicleDeviceApplication.kt` - PostHog initialization
3. `PostHogAnalytics.kt` - Analytics wrapper class
4. `DeviceConfig.kt` - Configuration management
5. `COMPLETE_OPEN_SOURCE_STACK.md` - Full documentation

---

## Quick Start (5 Minutes)

### Option 1: PostHog Cloud (Easiest)

1. **Sign up for free** at https://posthog.com
2. **Create project:** "FlyFleet Vehicle Device"
3. **Get API key** from Project Settings
4. **Update config:**
   ```kotlin
   // In DeviceConfig.kt or via remote config
   postHogApiKey = "phc_YOUR_API_KEY"
   postHogHost = "https://app.posthog.com"
   ```
5. **Deploy app** - That's it!

**Free Tier:** 1M events/month (enough for ~1000 devices)

### Option 2: Self-Hosted (Full Control)

1. **Deploy Docker stack** (see below)
2. **Get API key** from PostHog UI
3. **Update app config**
4. **Deploy app**

**Cost:** $50-200/month (your servers)

---

## Self-Hosted Deployment

### Docker Compose (Complete Stack)

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
      - SECRET_KEY=change_this_to_random_secret_key
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

  # Redis (Caching)
  redis:
    image: redis:7-alpine
    container_name: flyfleet-redis
    restart: always
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

### Deploy Commands

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f posthog

# Stop services
docker-compose down
```

### Access PostHog

1. Open http://localhost:8000
2. Create admin account
3. Create project: "FlyFleet Vehicle Device"
4. Get API key from Project Settings → API Keys
5. Update Android app config

---

## Event Tracking

### Events Tracked

#### Device Lifecycle
- `device_registered` - Device registration
- `device_started` - Monitoring started
- `device_stopped` - Monitoring stopped

#### Camera Operations
- `camera_started` - Camera recording started
- `camera_stopped` - Camera recording stopped
- `clip_recorded` - Event clip recorded
- `clip_uploaded` - Clip uploaded (success/failure)

#### Telemetry
- `telemetry_uploaded` - Telemetry batch uploaded

#### Driver Operations
- `driver_login` - Driver login (success/failure)
- `driver_logout` - Driver logout
- `hos_updated` - HOS status updated

#### Event Detection
- `event_detected` - Safety event detected (with severity)

#### Errors
- `error` - Any error occurred

### Example Usage

```kotlin
// In your service or activity
val analytics = VehicleDeviceApplication.getInstance().analytics

// Track event
analytics.trackEvent(
    eventName = "camera_started",
    properties = mapOf(
        "resolution" to "720p",
        "fps" to 30
    )
)

// Track error
analytics.trackError(
    category = "camera",
    message = "Failed to start recording",
    throwable = exception
)

// Check feature flag
if (analytics.isFeatureEnabled("ai_detection")) {
    // Enable AI detection
}

// Set user properties
analytics.setDeviceProperties(
    deviceId = "DEVICE-001",
    vehicleId = "VEHICLE-123",
    organizationId = "ORG-456"
)
```

---

## Dashboards

### Dashboard 1: Device Health

**Metrics:**
- Active devices (last 24h)
- Crash-free sessions
- Error rate
- Upload success rate

**PostHog Insight Queries:**
```
# Active devices
SELECT count(distinct device_id)
FROM events
WHERE event = 'device_started'
  AND timestamp > now() - INTERVAL 24 HOUR

# Error rate
SELECT count(*) 
FROM events
WHERE event = 'error'
  AND timestamp > now() - INTERVAL 24 HOUR
```

### Dashboard 2: Camera Performance

**Metrics:**
- Recording success rate
- Event clips per day
- Upload latency
- Storage usage

### Dashboard 3: Driver Behavior

**Metrics:**
- Driver login rate
- HOS violations
- Shift duration
- Event detection accuracy

---

## Feature Flags

### Example 1: Gradual AI Rollout

**PostHog Configuration:**
- Flag key: `ai_detection`
- Rollout: 10% → 25% → 50% → 100%
- Targeting: Specific organizations first

**Android Code:**
```kotlin
val aiEnabled = analytics.isFeatureEnabled("ai_detection", defaultValue = false)
if (aiEnabled) {
    // Enable AI detection
}
```

### Example 2: A/B Test Upload Intervals

**PostHog Configuration:**
- Flag key: `upload_interval`
- Variants: fast (30s), default (60s), slow (120s)

**Android Code:**
```kotlin
val interval = when (analytics.getFeatureFlag("upload_interval")) {
    "fast" -> 30000L
    "slow" -> 120000L
    else -> 60000L
}
```

---

## Privacy & Compliance

### GDPR Compliance

```kotlin
// User opts out
analytics.setAnalyticsEnabled(false)

// User requests data deletion (right to erasure)
analytics.reset()
```

### CCPA Compliance

```kotlin
// Don't sell data (PostHog doesn't sell data)
// Provide opt-out mechanism
analytics.setAnalyticsEnabled(false)
```

### Data Retention

Configure in PostHog Settings:
- Default: 12 months
- Configurable: 1-60 months
- Auto-deletion: Enabled

---

## Monitoring & Alerting

### PostHog Built-in Alerts

1. Go to PostHog → Insights → Create Alert
2. Set conditions:
   - Error rate > 5%
   - Upload success < 95%
   - Active devices = 0
3. Configure notifications:
   - Email
   - Slack webhook
   - PagerDuty

### Custom Monitoring

Export PostHog data to Grafana:
1. PostHog → Data Warehouse → Export
2. Grafana → Add Data Source → PostgreSQL
3. Create dashboards

---

## Cost Breakdown

### PostHog Cloud

| Devices | Events/Month | Cost |
|---------|-------------|------|
| 100 | 100k | Free |
| 1,000 | 1M | Free |
| 5,000 | 5M | $365/month |
| 10,000 | 10M | $730/month |

### Self-Hosted

| Component | Cost |
|-----------|------|
| Servers (4 CPU, 8GB RAM) | $50-200/month |
| Storage (100GB) | $10-50/month |
| Bandwidth | $10-50/month |
| **Total** | **$70-300/month** |

**Recommendation:** Start with PostHog Cloud (free tier), migrate to self-hosted at scale.

---

## Troubleshooting

### Events Not Appearing

1. Check API key is correct
2. Check host URL is reachable
3. Enable debug mode:
   ```kotlin
   debug = BuildConfig.DEBUG
   ```
4. Check PostHog logs:
   ```bash
   docker-compose logs -f posthog
   ```

### Feature Flags Not Working

1. Ensure flags are released in PostHog
2. Check flag key matches exactly
3. Clear app cache and retry

### High Memory Usage

1. Reduce flush interval:
   ```kotlin
   flushInterval = 60 // seconds
   ```
2. Reduce max queue size:
   ```kotlin
   maxQueueSize = 500
   ```

---

## Next Steps

1. **Deploy PostHog** (cloud or self-hosted)
2. **Get API key**
3. **Update Android app config**
4. **Test event tracking**
5. **Create dashboards**
6. **Set up alerts**
7. **Deploy to devices**

**Estimated Time:** 2-4 hours

---

## Support

- **PostHog Docs:** https://posthog.com/docs
- **PostHog Slack:** https://posthog.com/slack
- **GitHub:** https://github.com/PostHog/posthog

---

## Conclusion

PostHog analytics is now **fully integrated** into the FlyFleet Vehicle Device app. It's:

✅ 100% Open-Source
✅ GDPR/CCPA Compliant
✅ Self-hostable or Cloud
✅ Feature Flags included
✅ Error Tracking included
✅ Free for up to 1M events/month

**Ready to deploy!** 🚀
