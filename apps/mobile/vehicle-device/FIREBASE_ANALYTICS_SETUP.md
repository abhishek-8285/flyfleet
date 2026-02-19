# Production Mobile Analytics Setup Guide

**Platform:** Firebase (Google)
**Components:** Analytics, Crashlytics, Performance Monitoring, Remote Config
**Timeline:** 1-2 hours setup

---

## Step 1: Firebase Project Setup

### 1.1 Create Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click "Add project"
3. Enter project name: `flyfleet-vehicle-device`
4. Enable/disable Google Analytics (enable it)
5. Select analytics account
6. Click "Create project"

### 1.2 Add Android App to Firebase

1. In Firebase Console, click "Add app" → Android
2. Enter package name: `com.flyfleet.vehicledevice`
3. Enter app nickname: "FlyFleet Vehicle Device"
4. Enter debug signing certificate SHA-1:
   ```bash
   # Get debug SHA-1
   keytool -list -v -keystore ~/.android/debug.keystore -alias androiddebugkey -storepass android -keypass android
   ```
5. Download `google-services.json`
6. Place file in: `apps/mobile/vehicle-device/android-device/google-services.json`

### 1.3 Add Release Signing Certificate

For production crash reporting:

```bash
# Get release SHA-1
keytool -list -v -keystore /path/to/keystore.jks -alias your_alias_name
```

Add SHA-1 to Firebase Console:
- Project Settings → Your apps → Add fingerprint

---

## Step 2: Enable Firebase Services

### 2.1 Firebase Analytics

**Already enabled by default**

Configuration:
- Go to Analytics → Settings
- Enable "Google Signals" for cross-device tracking
- Set data retention to 14 months (GDPR compliant)

### 2.2 Firebase Crashlytics

1. Go to Crashlytics in Firebase Console
2. Click "Get started"
3. Follow setup wizard (already configured in code)

Features:
- Automatic crash reporting
- Non-fatal error tracking
- Custom keys and logs
- User tracking

### 2.3 Firebase Performance Monitoring

1. Go to Performance in Firebase Console
2. Click "Get started"
3. Already configured in code

Features:
- Automatic HTTP/S monitoring
- App start time tracking
- Custom trace tracking
- Network request monitoring

### 2.4 Firebase Remote Config

1. Go to Remote Config in Firebase Console
2. Click "Get started"

Features:
- Feature flags
- A/B testing
- Parameter customization
- Remote app configuration

---

## Step 3: Analytics Event Tracking

### 3.1 Automatic Events (No Code Required)

Firebase automatically tracks:
- First open
- App updates
- Session start
- Screen views (if enabled)
- User engagement

### 3.2 Custom Events (Implemented)

Our app tracks these custom events:

#### Device Events
- `device_registered` - Device registration
- `device_started` - Monitoring started
- `device_stopped` - Monitoring stopped
- `config_updated` - Config changed

#### Camera Events
- `camera_started` - Camera recording started
- `camera_stopped` - Camera recording stopped
- `clip_recorded` - Event clip recorded
- `clip_uploaded` - Clip uploaded to backend

#### Telemetry Events
- `telemetry_uploaded` - Telemetry batch uploaded

#### Driver Events
- `driver_login` - Driver login
- `driver_logout` - Driver logout
- `hos_updated` - HOS status updated

#### Event Detection
- `event_detected` - Safety event detected

#### Error Events
- `error_occurred` - Error occurred

### 3.3 User Properties

Tracked user properties:
- `device_id` - Unique device identifier
- `device_model` - Device model name
- `android_version` - Android OS version
- `app_version` - App version
- `driver_id` - Current driver ID
- `vehicle_id` - Assigned vehicle ID
- `organization_id` - Fleet organization ID

---

## Step 4: Viewing Analytics

### 4.1 Real-time Dashboard

Firebase Console → Analytics → Realtime
- See active users
- Current events
- User properties

### 4.2 Standard Reports

Firebase Console → Analytics → Reports
- User engagement
- Retention
- Demographics
- Tech analysis

### 4.3 Custom Dashboards

Firebase Console → Analytics → Custom Reports
- Create custom funnels
- Build custom segments
- Analyze specific events

### 4.4 Crashlytics Dashboard

Firebase Console → Crashlytics
- Crash-free users
- Crash trends
- Issue details
- Stack traces

### 4.5 Performance Dashboard

Firebase Console → Performance
- App start time
- HTTP request latency
- Custom traces
- Network performance

---

## Step 5: Privacy & Compliance

### 5.1 GDPR Compliance

```kotlin
// Disable analytics for users who opt-out
analytics.setAnalyticsEnabled(false)

// Reset analytics data (right to erasure)
analytics.resetAnalyticsData()
```

### 5.2 CCPA Compliance

```kotlin
// Don't sell personal data (Firebase doesn't sell data)
// Provide opt-out mechanism
analytics.setAnalyticsEnabled(false)
```

### 5.3 Data Retention

Default Firebase retention:
- Analytics: 14 months (configurable 2-50 months)
- Crashlytics: 90 days
- Performance: 30 days

---

## Step 6: Production Monitoring Dashboard

### 6.1 Key Metrics to Monitor

#### Reliability Metrics
- Crash-free users: Target >99.9%
- ANR rate: Target <0.1%
- Error rate: Target <1%

#### Performance Metrics
- Cold start time: Target <3 seconds
- Upload latency (p95): Target <10 seconds
- Memory usage: Target <200MB

#### Business Metrics
- Events detected per day
- Miles driven per day
- Upload success rate: Target >99%
- Driver login rate

### 6.2 Alerting Setup

Set up alerts in Firebase Console:

1. Go to Crashlytics → Settings → Alerts
2. Create alert for:
   - New crash issue
   - Crash rate increase >1%
   - ANR rate increase >0.5%

3. Configure email/Slack notifications

### 6.3 BigQuery Export (Optional)

For advanced analytics:

1. Firebase Console → Project Settings → Integrations → BigQuery
2. Link BigQuery project
3. Select data to export:
   - Analytics events
   - Crashlytics crashes
   - Performance data

Query example:
```sql
SELECT
  event_name,
  COUNT(*) as event_count,
  user_pseudo_id
FROM `project.analytics_*.events_*`
WHERE event_name = 'event_detected'
GROUP BY event_name, user_pseudo_id
ORDER BY event_count DESC
LIMIT 100
```

---

## Step 7: A/B Testing (Optional)

### 7.1 Create A/B Test

1. Firebase Console → A/B Testing → Create Experiment
2. Select target (e.g., Remote Config)
3. Define audience (e.g., specific organization)
4. Set variants (e.g., different upload intervals)
5. Define success metric (e.g., upload success rate)

### 7.2 Example: Test Upload Interval

```kotlin
// In DeviceConfig
val uploadIntervalMs = FirebaseRemoteConfig.getInstance()
    .getLong("telemetry_upload_interval_ms")
```

Remote Config parameter:
- Key: `telemetry_upload_interval_ms`
- Default: 60000 (60 seconds)
- Variant A: 30000 (30 seconds)
- Variant B: 120000 (120 seconds)

---

## Step 8: Feature Flags

### 8.1 Setup Remote Config

```kotlin
// In DeviceConfig.kt
val remoteConfig = FirebaseRemoteConfig.getInstance()

// Set defaults
remoteConfig.setDefaultsAsync(mapOf(
    "camera_enabled" to true,
    "ai_enabled" to false,
    "telemetry_batch_size" to 60,
    "upload_interval_ms" to 60000L
))

// Fetch config
remoteConfig.fetchAndActivate()
```

### 8.2 Feature Flag Examples

```kotlin
// Enable/disable camera
val cameraEnabled = remoteConfig.getBoolean("camera_enabled")

// AI feature flag (gradual rollout)
val aiEnabled = remoteConfig.getBoolean("ai_enabled")

// Configuration parameters
val batchSize = remoteConfig.getLong("telemetry_batch_size").toInt()
val uploadInterval = remoteConfig.getLong("upload_interval_ms")
```

---

## Step 9: Production Checklist

### Before Launch

- [ ] Firebase project created
- [ ] Android app registered
- [ ] `google-services.json` added
- [ ] Release SHA-1 added
- [ ] Analytics enabled
- [ ] Crashlytics enabled
- [ ] Performance Monitoring enabled
- [ ] Remote Config enabled
- [ ] Test events appearing in Realtime
- [ ] Test crash appearing in Crashlytics
- [ ] Privacy policy updated
- [ ] GDPR compliance implemented

### After Launch

- [ ] Monitor crash-free users
- [ ] Review error logs daily
- [ ] Check performance metrics weekly
- [ ] Update Remote Config as needed
- [ ] Export data to BigQuery (optional)
- [ ] Set up automated alerts

---

## Step 10: Troubleshooting

### Analytics Not Working

1. Check `google-services.json` is in correct location
2. Verify package name matches
3. Check SHA-1 fingerprint
4. Wait 24 hours for data to appear
5. Check Realtime dashboard

### Crashlytics Not Reporting

1. Ensure app is not in debug mode (Crashlytics disabled in debug by default)
2. Force test crash:
   ```kotlin
   FirebaseCrashlytics.getInstance().crash()
   ```
3. Wait 5 minutes for crash to appear
4. Check ProGuard rules if using release build

### Performance Data Missing

1. Ensure app has network connectivity
2. Check Firebase SDK version
3. Wait 24 hours for data to process
4. Verify custom traces are started/stopped correctly

---

## Cost Estimation

### Firebase Free Tier (Spark Plan)

Includes:
- Analytics: Unlimited
- Crashlytics: Unlimited
- Performance: Unlimited
- Remote Config: 500K subscribers

### Firebase Paid Tier (Blaze Plan)

Pay-as-you-go:
- Analytics: Free
- Crashlytics: Free
- Performance: Free
- Remote Config: Free up to 10M subscribers

**Estimated Cost for 1000 Devices:**
- $0/month (all services free for this scale)

**Estimated Cost for 10,000 Devices:**
- $0/month (still within free tier)

**Estimated Cost for 100,000 Devices:**
- Remote Config: ~$25/month (if >10M subscribers)
- BigQuery (optional): ~$50-100/month
- **Total: ~$75-125/month**

---

## Conclusion

Firebase provides a comprehensive, production-grade analytics and monitoring platform at no cost for typical fleet deployments. The integration is complete and ready for production use.

**Next Steps:**
1. Create Firebase project
2. Add Android app
3. Download `google-services.json`
4. Test analytics and crash reporting
5. Set up alerts
6. Launch!
