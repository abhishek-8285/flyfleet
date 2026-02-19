# Production Monitoring Dashboard

**Platform:** Firebase Console + Custom Dashboard
**Purpose:** Real-time monitoring of fleet device health, performance, and business metrics

---

## Dashboard 1: Device Health Overview

### Key Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **Active Devices** | - | - |
| **Crash-Free Users** | >99.9% | <99% |
| **ANR Rate** | <0.1% | >0.5% |
| **Error Rate** | <1% | >5% |
| **Service Uptime** | >99.9% | <99% |

### Widgets

1. **Active Devices Over Time** (Line Chart)
   - Metric: `active_devices`
   - Time range: Last 24 hours
   - Group by: Organization

2. **Crash-Free Users** (Scorecard)
   - Metric: `crash_free_users`
   - Comparison: vs last week

3. **Errors by Category** (Pie Chart)
   - Dimension: `error_category`
   - Time range: Last 24 hours

4. **Device States** (Pie Chart)
   - Dimension: `runtime_state`
   - Values: BOOTING, IDLE, ACTIVE_DRIVE, DEGRADED, RECOVERY

---

## Dashboard 2: Performance Monitoring

### Key Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **Cold Start** | <3s | >5s |
| **Camera Ready** | <2s | >3s |
| **Upload Latency (p95)** | <10s | >30s |
| **Memory Usage** | <200MB | >300MB |
| **CPU Usage** | <20% | >50% |

### Widgets

1. **App Start Time** (Time Series)
   - Metric: `app_start_time`
   - Percentiles: p50, p95, p99
   - Time range: Last 24 hours

2. **API Call Latency** (Heat Map)
   - Metric: `api_call_duration_ms`
   - Dimension: `endpoint`
   - Time range: Last 24 hours

3. **Memory Usage** (Time Series)
   - Metric: `memory_usage_mb`
   - Threshold line: 200MB
   - Time range: Last 24 hours

4. **CPU Usage** (Time Series)
   - Metric: `cpu_usage_percent`
   - Threshold line: 20%
   - Time range: Last 24 hours

5. **Storage Usage** (Gauge)
   - Metric: `storage_usage_percent`
   - Threshold: 90%
   - Time range: Current

---

## Dashboard 3: Camera & Recording

### Key Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **Recording Success Rate** | >99% | <95% |
| **Dropped Frame Rate** | <3% | >10% |
| **Event Clips/Day** | - | - |
| **Upload Success Rate** | >99% | <95% |

### Widgets

1. **Recording Status** (Pie Chart)
   - Dimension: `recording_status`
   - Values: recording, stopped, error

2. **Dropped Frame Rate** (Time Series)
   - Metric: `dropped_frame_percent`
   - Threshold line: 3%
   - Time range: Last 24 hours

3. **Event Clips Recorded** (Bar Chart)
   - Metric: `clip_recorded_count`
   - Dimension: `event_type`
   - Time range: Last 24 hours

4. **Upload Success Rate** (Time Series)
   - Metric: `upload_success_rate`
   - Threshold line: 99%
   - Time range: Last 24 hours

5. **Storage Usage by Type** (Stacked Area)
   - Metrics: `clips_storage_mb`, `ring_buffer_storage_mb`
   - Time range: Last 24 hours

---

## Dashboard 4: Telemetry & GPS

### Key Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **GPS Accuracy** | <10m | >50m |
| **Telemetry Upload Success** | >99% | <95% |
| **Points Collected/Minute** | 60 | <30 |
| **Batch Upload Time** | <5s | >15s |

### Widgets

1. **GPS Signal Quality** (Histogram)
   - Metric: `gps_accuracy_meters`
   - Buckets: 0-10m, 10-50m, 50-100m, >100m
   - Time range: Last 24 hours

2. **Telemetry Points Collected** (Time Series)
   - Metric: `telemetry_points_count`
   - Time range: Last 24 hours

3. **Upload Success Rate** (Time Series)
   - Metric: `telemetry_upload_success_rate`
   - Threshold line: 99%
   - Time range: Last 24 hours

4. **Network Type Distribution** (Pie Chart)
   - Dimension: `network_type`
   - Values: WiFi, Cellular_4G, Cellular_3G, Cellular_2G, None

---

## Dashboard 5: Driver & HOS

### Key Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **Driver Login Rate** | >90% | <80% |
| **HOS Violations/Day** | - | - |
| **Shift Duration (avg)** | 8-10 hours | >14 hours |
| **Login Success Rate** | >99% | <95% |

### Widgets

1. **Active Drivers** (Time Series)
   - Metric: `active_drivers_count`
   - Time range: Last 24 hours

2. **Duty Status Distribution** (Pie Chart)
   - Dimension: `duty_status`
   - Values: ON_DUTY, DRIVING, SLEEPER_BERTH, OFF_DUTY

3. **HOS Violations** (Bar Chart)
   - Metric: `hos_violation_count`
   - Dimension: `violation_type`
   - Time range: Last 7 days

4. **Shift Duration Distribution** (Histogram)
   - Metric: `shift_duration_hours`
   - Buckets: 0-4h, 4-8h, 8-12h, 12-14h, >14h

---

## Dashboard 6: Event Detection

### Key Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **Events Detected/Day** | - | - |
| **Detection Accuracy** | >95% | <90% |
| **False Positive Rate** | <5% | >10% |
| **Critical Events** | - | Immediate alert |

### Widgets

1. **Events by Type** (Bar Chart)
   - Metric: `event_detected_count`
   - Dimension: `event_type`
   - Time range: Last 24 hours

2. **Events by Severity** (Pie Chart)
   - Dimension: `severity`
   - Values: INFO, WARN, CRITICAL

3. **Detection Confidence** (Histogram)
   - Metric: `detection_confidence`
   - Buckets: 0-50%, 50-70%, 70-90%, 90-100%

4. **Event Upload Success** (Time Series)
   - Metric: `event_upload_success_rate`
   - Threshold line: 99%
   - Time range: Last 24 hours

---

## Dashboard 7: Backend API Health

### Key Metrics

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **API Success Rate** | >99% | <95% |
| **API Latency (p95)** | <1s | >5s |
| **Backend Errors** | 0 | >10/hour |

### Widgets

1. **API Call Success Rate** (Time Series)
   - Metric: `api_success_rate`
   - Dimension: `endpoint`
   - Threshold line: 99%
   - Time range: Last 24 hours

2. **API Latency by Endpoint** (Heat Map)
   - Metric: `api_latency_ms`
   - Dimension: `endpoint`
   - Percentiles: p50, p95, p99

3. **Backend Errors** (Time Series)
   - Metric: `backend_error_count`
   - Dimension: `error_type`
   - Time range: Last 24 hours

4. **Request Volume** (Time Series)
   - Metric: `api_request_count`
   - Dimension: `endpoint`
   - Time range: Last 24 hours

---

## Alerts Configuration

### Critical Alerts (Immediate Notification)

| Alert | Condition | Notification |
|-------|-----------|--------------|
| **Service Down** | Active devices = 0 for 5 min | SMS + Email + Slack |
| **Crash Rate Spike** | Crash rate >5% | SMS + Email + Slack |
| **Upload Failure** | Upload success <90% for 10 min | Email + Slack |
| **Storage Critical** | Storage >95% full | Email + Slack |
| **Backend Error Spike** | Backend errors >100/hour | SMS + Email + Slack |

### Warning Alerts (Daily Digest)

| Alert | Condition | Notification |
|-------|-----------|--------------|
| **ANR Rate High** | ANR rate >0.5% | Email |
| **Memory Pressure** | Memory >250MB avg | Email |
| **GPS Accuracy Poor** | GPS accuracy >50m for 1 hour | Email |
| **Driver Login Failures** | Login success <95% | Email |

---

## Firebase Console Setup

### Step 1: Create Custom Dashboards

1. Go to Firebase Console → Analytics → Dashboards
2. Click "Create Dashboard"
3. Add widgets for each metric above
4. Save dashboard

### Step 2: Set Up Alerts

1. Go to Firebase Console → Crashlytics → Settings → Alerts
2. Create alert for each critical metric
3. Configure notification channels:
   - Email: ops@flyfleet.com
   - Slack: #flyfleet-alerts webhook
   - SMS: On-call engineer phone

### Step 3: Configure BigQuery Export

1. Firebase Console → Project Settings → Integrations → BigQuery
2. Link BigQuery project
3. Enable export for:
   - Analytics events
   - Crashlytics crashes
   - Performance data

### Step 4: Create Custom Reports

1. Firebase Console → Analytics → Custom Reports
2. Create reports for:
   - Device health trends
   - Event detection accuracy
   - Driver behavior analysis
   - Upload performance

---

## Grafana Dashboard (Optional)

For advanced monitoring, set up Grafana with Firebase data:

### Data Sources

1. **Firebase BigQuery Export**
   - Connector: BigQuery
   - Query: Custom SQL queries

2. **Firebase Realtime Database**
   - Connector: Firebase
   - Query: Real-time metrics

### Dashboard Panels

Import pre-built panels from Firebase or create custom:

```json
{
  "dashboard": {
    "title": "FlyFleet Device Health",
    "panels": [
      {
        "title": "Active Devices",
        "type": "graph",
        "targets": [
          {
            "query": "SELECT COUNT(DISTINCT device_id) FROM analytics WHERE timestamp > NOW() - INTERVAL 1 HOUR"
          }
        ]
      }
    ]
  }
}
```

---

## Monitoring Best Practices

### 1. Set Meaningful Thresholds

- Base thresholds on historical data
- Use percentiles (p95, p99) not averages
- Set different thresholds for different times of day

### 2. Avoid Alert Fatigue

- Only alert on actionable items
- Use warning vs critical levels
- Implement alert suppression (don't alert on same issue repeatedly)

### 3. Regular Review

- Weekly: Review all alerts, adjust thresholds
- Monthly: Review dashboard relevance
- Quarterly: Add new metrics as needed

### 4. Documentation

- Document what each metric means
- Document alert response procedures
- Keep runbooks up to date

---

## Incident Response Runbook

### Service Outage

**Symptom:** Active devices = 0

**Response:**
1. Check backend service health
2. Check Firebase status
3. Check recent deployments
4. Rollback if needed
5. Notify customers

### High Crash Rate

**Symptom:** Crash-free users <99%

**Response:**
1. Check Crashlytics for new issues
2. Identify affected devices/versions
3. Check recent changes
4. Deploy fix or rollback
5. Notify affected users

### Upload Failures

**Symptom:** Upload success <95%

**Response:**
1. Check backend API health
2. Check network connectivity
3. Check storage quota
4. Check backend logs
5. Scale backend if needed

---

## Conclusion

This monitoring setup provides comprehensive visibility into device health, performance, and business metrics. With proper alerting and dashboards, the operations team can quickly identify and resolve issues before they impact customers.

**Estimated Setup Time:** 4-6 hours
**Ongoing Maintenance:** 1-2 hours/week
